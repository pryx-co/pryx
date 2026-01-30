package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OTLPConfig holds configuration for OTLP exporters
type OTLPConfig struct {
	Endpoint    string
	APIKey      string
	ServiceName string
}

// OTLPExporter handles OpenTelemetry-compatible export via HTTP
type OTLPExporter struct {
	config OTLPConfig
	client *http.Client
}

// TraceSpan represents a trace span for export
type TraceSpan struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	ParentID   string            `json:"parent_id,omitempty"`
	Name       string            `json:"name"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Status     string            `json:"status"`
}

// Metric represents a metric for export
type Metric struct {
	Name       string            `json:"name"`
	Value      float64           `json:"value"`
	Timestamp  time.Time         `json:"timestamp"`
	Type       string            `json:"type"` // counter, gauge, histogram
	Attributes map[string]string `json:"attributes,omitempty"`
}

// NewOTLPExporter creates a new OTLP exporter
func NewOTLPExporter(config OTLPConfig) *OTLPExporter {
	if config.Endpoint == "" {
		config.Endpoint = "http://localhost:4318"
	}
	if config.ServiceName == "" {
		config.ServiceName = "pryx-runtime"
	}

	return &OTLPExporter{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExportSpans exports trace spans to the OTLP endpoint
func (e *OTLPExporter) ExportSpans(spans []TraceSpan) error {
	if len(spans) == 0 {
		return nil
	}

	payload := map[string]interface{}{
		"service": e.config.ServiceName,
		"spans":   spans,
	}

	return e.sendRequest("/v1/traces", payload)
}

// ExportMetrics exports metrics to the OTLP endpoint
func (e *OTLPExporter) ExportMetrics(metrics []Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	payload := map[string]interface{}{
		"service": e.config.ServiceName,
		"metrics": metrics,
	}

	return e.sendRequest("/v1/metrics", payload)
}

// sendRequest sends data to the OTLP endpoint
func (e *OTLPExporter) sendRequest(path string, payload interface{}) error {
	url := e.config.Endpoint + path

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if e.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.config.APIKey)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("OTLP endpoint returned error: %s", resp.Status)
	}

	return nil
}

// RecordSpan creates and exports a trace span
func (e *OTLPExporter) RecordSpan(name string, startTime time.Time, attributes map[string]string) TraceSpan {
	return TraceSpan{
		TraceID:    generateID(),
		SpanID:     generateID(),
		Name:       name,
		StartTime:  startTime,
		EndTime:    time.Now(),
		Attributes: attributes,
		Status:     "ok",
	}
}

// RecordMetric creates a metric
func (e *OTLPExporter) RecordMetric(name string, value float64, metricType string, attributes map[string]string) Metric {
	return Metric{
		Name:       name,
		Value:      value,
		Timestamp:  time.Now(),
		Type:       metricType,
		Attributes: attributes,
	}
}

// IsHealthy checks if the OTLP endpoint is reachable
func (e *OTLPExporter) IsHealthy() bool {
	req, err := http.NewRequest("GET", e.config.Endpoint+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode < 400
}

// generateID creates a simple unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// TelemetryBatch holds a batch of telemetry data
type TelemetryBatch struct {
	Spans   []TraceSpan
	Metrics []Metric
}

// ExportBatch exports a batch of spans and metrics
func (e *OTLPExporter) ExportBatch(batch TelemetryBatch) error {
	if err := e.ExportSpans(batch.Spans); err != nil {
		return fmt.Errorf("failed to export spans: %w", err)
	}

	if err := e.ExportMetrics(batch.Metrics); err != nil {
		return fmt.Errorf("failed to export metrics: %w", err)
	}

	return nil
}
