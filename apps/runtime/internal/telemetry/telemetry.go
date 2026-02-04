// Package telemetry provides OpenTelemetry integration for Pryx runtime.
// Exports traces, metrics, and logs via OTLP HTTP to Cloudflare Workers.
package telemetry

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"pryx-core/internal/config"
	"pryx-core/internal/keychain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

const (
	tracerName  = "pryx-core"
	serviceName = "pryx-runtime"
)

// Provider manages OpenTelemetry configuration and export
type Provider struct {
	cfg      *config.Config
	keychain *keychain.Keychain
	tracer   trace.Tracer
	tp       *sdktrace.TracerProvider
	mp       *sdkmetric.MeterProvider
	meter    metric.Meter
	metrics  *Metrics
	logBatch *logBatcher
	enabled  bool
	sampling float64 // 0.0 to 1.0
	deviceID string
}

type Metrics struct {
	requests      metric.Int64Counter
	errors        metric.Int64Counter
	latencyMs     metric.Float64Histogram
	providerUsage metric.Int64Counter
}

type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

type logBatcher struct {
	exporter      *OTLPExporter
	flushInterval time.Duration
	maxBatch      int
	mu            sync.Mutex
	logs          []LogEntry
	stop          chan struct{}
	done          chan struct{}
}

var (
	globalProviderMu sync.RWMutex
	globalProvider   *Provider
)

// NewProvider creates a new telemetry provider
func NewProvider(cfg *config.Config, kc *keychain.Keychain) (*Provider, error) {
	p := &Provider{
		cfg:      cfg,
		keychain: kc,
		enabled:  true,
		sampling: 1.0, // Default: sample all
	}

	// Check if telemetry is disabled
	if os.Getenv("PRYX_TELEMETRY_DISABLED") == "true" {
		p.enabled = false
		return p, nil
	}

	// Parse sampling rate
	if v := os.Getenv("PRYX_TELEMETRY_SAMPLING"); v != "" {
		fmt.Sscanf(v, "%f", &p.sampling)
	}

	// Get device ID from keychain or generate
	deviceID, err := kc.Get("device_id")
	if err != nil {
		// Generate new device ID
		deviceID = generateDeviceID()
		kc.Set("device_id", deviceID)
	}
	p.deviceID = deviceID

	// Initialize OpenTelemetry
	if err := p.init(context.Background()); err != nil {
		log.Printf("Telemetry: initialization failed (will use noop): %v", err)
		// Still return a valid provider, just with noop tracing
		// This allows tests and callers to work even when network/auth fails
		p.tracer = &noopTracer{}
		setGlobalProvider(p)
		return p, nil
	}
	setGlobalProvider(p)

	return p, nil
}

// init sets up the OTLP exporter and tracer provider
func (p *Provider) init(ctx context.Context) error {
	if !p.enabled {
		return nil
	}

	// Get cloud access token for auth
	token, err := p.keychain.Get("cloud_access_token")
	if err != nil {
		// Not logged in - telemetry will be queued/buffered
		log.Println("Telemetry: No cloud token, will queue events")
	}

	// Create OTLP HTTP exporter
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(p.cfg.CloudAPIUrl),
		otlptracehttp.WithURLPath("/v1/traces"),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + token,
		}),
	}

	// Use insecure for local dev, TLS for production
	if os.Getenv("PRYX_DEV") == "true" {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	client := otlptracehttp.NewClient(opts...)
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service attributes
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("service.version", getVersion()),
			attribute.String("device.id", p.deviceID),
			attribute.String("host.arch", runtime.GOARCH),
			attribute.String("host.os", runtime.GOOS),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider with sampling
	p.tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(p.sampling)),
	)

	otel.SetTracerProvider(p.tp)
	p.tracer = p.tp.Tracer(tracerName)

	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(p.cfg.CloudAPIUrl),
		otlpmetrichttp.WithURLPath("/v1/metrics"),
		otlpmetrichttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + token,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create OTLP metrics exporter: %w", err)
	}

	reader := sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(10*time.Second))
	p.mp = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(p.mp)
	p.meter = p.mp.Meter(tracerName)

	metrics, err := newMetrics(p.meter)
	if err != nil {
		return fmt.Errorf("failed to create metrics: %w", err)
	}
	p.metrics = metrics

	logExporter := NewOTLPExporter(OTLPConfig{
		Endpoint:    p.cfg.CloudAPIUrl,
		APIKey:      token,
		ServiceName: serviceName,
	})
	p.logBatch = newLogBatcher(logExporter, 5*time.Second, 200)
	p.logBatch.Start()

	log.Printf("Telemetry: Initialized with sampling=%.2f", p.sampling)
	return nil
}

// StartSpan starts a new span with the given name and attributes
func (p *Provider) StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	if !p.enabled || p.tracer == nil {
		return ctx, &noopSpan{}
	}
	return p.tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// Shutdown gracefully shuts down the telemetry provider
func (p *Provider) Shutdown(ctx context.Context) error {
	if !p.enabled || p.tp == nil {
		return nil
	}
	if p.logBatch != nil {
		_ = p.logBatch.Shutdown(ctx)
	}
	if p.mp != nil {
		_ = p.mp.Shutdown(ctx)
	}
	return p.tp.Shutdown(ctx)
}

// Enabled returns whether telemetry is enabled
func (p *Provider) Enabled() bool {
	return p.enabled
}

// DeviceID returns the device identifier
func (p *Provider) DeviceID() string {
	return p.deviceID
}

func newMetrics(meter metric.Meter) (*Metrics, error) {
	requests, err := meter.Int64Counter("pryx.requests", metric.WithDescription("Total request count"))
	if err != nil {
		return nil, err
	}
	errors, err := meter.Int64Counter("pryx.errors", metric.WithDescription("Total error count"))
	if err != nil {
		return nil, err
	}
	latencyMs, err := meter.Float64Histogram("pryx.latency_ms", metric.WithDescription("Request latency in ms"))
	if err != nil {
		return nil, err
	}
	providerUsage, err := meter.Int64Counter("pryx.provider.usage", metric.WithDescription("Provider usage count"))
	if err != nil {
		return nil, err
	}

	return &Metrics{
		requests:      requests,
		errors:        errors,
		latencyMs:     latencyMs,
		providerUsage: providerUsage,
	}, nil
}

func (m *Metrics) RecordRequest(ctx context.Context, name string) {
	if m == nil {
		return
	}
	m.requests.Add(ctx, 1, metric.WithAttributes(attribute.String("name", name)))
}

func (m *Metrics) RecordError(ctx context.Context, name string) {
	if m == nil {
		return
	}
	m.errors.Add(ctx, 1, metric.WithAttributes(attribute.String("name", name)))
}

func (m *Metrics) RecordLatency(ctx context.Context, name string, latency time.Duration) {
	if m == nil {
		return
	}
	m.latencyMs.Record(ctx, float64(latency.Milliseconds()), metric.WithAttributes(attribute.String("name", name)))
}

func (m *Metrics) RecordProviderUsage(ctx context.Context, provider string) {
	if m == nil {
		return
	}
	m.providerUsage.Add(ctx, 1, metric.WithAttributes(attribute.String("provider", provider)))
}

func newLogBatcher(exporter *OTLPExporter, flushInterval time.Duration, maxBatch int) *logBatcher {
	if flushInterval <= 0 {
		flushInterval = 5 * time.Second
	}
	if maxBatch <= 0 {
		maxBatch = 200
	}
	return &logBatcher{
		exporter:      exporter,
		flushInterval: flushInterval,
		maxBatch:      maxBatch,
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
	}
}

func (b *logBatcher) Start() {
	go b.run()
}

func (b *logBatcher) Shutdown(ctx context.Context) error {
	select {
	case <-b.done:
		return nil
	default:
	}
	close(b.stop)
	select {
	case <-b.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *logBatcher) Add(level, message string, fields map[string]interface{}) {
	if b == nil {
		return
	}
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}
	b.add(entry)
}

func (b *logBatcher) run() {
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()
	defer close(b.done)
	for {
		select {
		case <-ticker.C:
			b.flush()
		case <-b.stop:
			b.flush()
			return
		}
	}
}

func (b *logBatcher) add(entry LogEntry) {
	b.mu.Lock()
	b.logs = append(b.logs, entry)
	shouldFlush := len(b.logs) >= b.maxBatch
	batch := b.logs
	if shouldFlush {
		b.logs = nil
	}
	b.mu.Unlock()
	if shouldFlush {
		b.export(batch)
	}
}

func (b *logBatcher) flush() {
	b.mu.Lock()
	if len(b.logs) == 0 {
		b.mu.Unlock()
		return
	}
	batch := b.logs
	b.logs = nil
	b.mu.Unlock()
	b.export(batch)
}

func (b *logBatcher) export(entries []LogEntry) {
	if b.exporter == nil || len(entries) == 0 {
		return
	}
	payload := map[string]interface{}{
		"service": b.exporter.config.ServiceName,
		"logs":    entries,
	}
	_ = b.exporter.sendRequest("/v1/logs", payload)
}

func setGlobalProvider(p *Provider) {
	globalProviderMu.Lock()
	globalProvider = p
	globalProviderMu.Unlock()
}

func GlobalProvider() *Provider {
	globalProviderMu.RLock()
	defer globalProviderMu.RUnlock()
	return globalProvider
}

// Helper functions

func generateDeviceID() string {
	// Simple device ID generation - in production, use proper crypto
	return fmt.Sprintf("pryx-%d", time.Now().UnixNano())
}

func getVersion() string {
	if v := os.Getenv("PRYX_VERSION"); v != "" {
		return v
	}
	return "dev"
}

// noopSpan is a no-op span for when telemetry is disabled
type noopSpan struct {
	embedded.Span
}

// noopTracer is a no-op tracer for when telemetry is disabled or init fails
type noopTracer struct {
	embedded.Tracer
}

func (t *noopTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ctx, &noopSpan{}
}

func (s *noopSpan) End(options ...trace.SpanEndOption)                  {}
func (s *noopSpan) AddEvent(name string, options ...trace.EventOption)  {}
func (s *noopSpan) IsRecording() bool                                   { return false }
func (s *noopSpan) RecordError(err error, options ...trace.EventOption) {}
func (s *noopSpan) SpanContext() trace.SpanContext                      { return trace.SpanContext{} }
func (s *noopSpan) SetStatus(code codes.Code, description string)       {}
func (s *noopSpan) SetName(name string)                                 {}
func (s *noopSpan) SetAttributes(kv ...attribute.KeyValue)              {}
func (s *noopSpan) TracerProvider() trace.TracerProvider                { return nil }
func (s *noopSpan) AddLink(link trace.Link)                             {}
