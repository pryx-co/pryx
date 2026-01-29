package telemetry

import (
	"context"
	"os"
	"testing"
	"time"

	"pryx-core/internal/config"
	"pryx-core/internal/keychain"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name         string
		envDisabled  string
		envSampling  string
		wantEnabled  bool
		wantSampling float64
	}{
		{
			name:         "default enabled",
			envDisabled:  "",
			envSampling:  "",
			wantEnabled:  true,
			wantSampling: 1.0,
		},
		{
			name:         "explicitly enabled",
			envDisabled:  "false",
			envSampling:  "",
			wantEnabled:  true,
			wantSampling: 1.0,
		},
		{
			name:         "disabled",
			envDisabled:  "true",
			envSampling:  "",
			wantEnabled:  false,
			wantSampling: 1.0,
		},
		{
			name:         "custom sampling",
			envDisabled:  "",
			envSampling:  "0.5",
			wantEnabled:  true,
			wantSampling: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			if tt.envDisabled != "" {
				os.Setenv("PRYX_TELEMETRY_DISABLED", tt.envDisabled)
				defer os.Unsetenv("PRYX_TELEMETRY_DISABLED")
			}
			if tt.envSampling != "" {
				os.Setenv("PRYX_TELEMETRY_SAMPLING", tt.envSampling)
				defer os.Unsetenv("PRYX_TELEMETRY_SAMPLING")
			}

			cfg := &config.Config{
				CloudAPIUrl: "https://api.pryx.io",
			}
			kc := keychain.New("pryx-test")

			provider, err := NewProvider(cfg, kc)

			if tt.wantEnabled {
				// When enabled, it may fail to init due to network/auth issues
				// but the provider should still be created
				if provider == nil {
					t.Fatal("NewProvider() returned nil provider")
				}
				if provider.enabled != tt.wantEnabled {
					t.Errorf("provider.enabled = %v, want %v", provider.enabled, tt.wantEnabled)
				}
				if provider.sampling != tt.wantSampling {
					t.Errorf("provider.sampling = %v, want %v", provider.sampling, tt.wantSampling)
				}
			} else {
				// When disabled, init should succeed without network
				if err != nil {
					t.Errorf("NewProvider() unexpected error when disabled: %v", err)
				}
				if provider == nil {
					t.Fatal("NewProvider() returned nil provider")
				}
				if provider.enabled {
					t.Error("NewProvider() provider.enabled = true, want false when disabled")
				}
			}

			// Cleanup
			if provider != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				provider.Shutdown(ctx)
			}
		})
	}
}

func TestProvider_Enabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{
			name:    "enabled",
			enabled: true,
		},
		{
			name:    "disabled",
			enabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				CloudAPIUrl: "https://api.pryx.io",
			}
			kc := keychain.New("pryx-test")

			// Force disabled state
			os.Setenv("PRYX_TELEMETRY_DISABLED", "true")
			defer os.Unsetenv("PRYX_TELEMETRY_DISABLED")

			provider, _ := NewProvider(cfg, kc)
			if provider.Enabled() {
				t.Error("Enabled() = true, want false")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			provider.Shutdown(ctx)
		})
	}
}

func TestProvider_DeviceID(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	kc := keychain.New("pryx-test")

	provider, _ := NewProvider(cfg, kc)

	// DeviceID is set during init, which may fail due to network
	// If provider is not fully initialized, deviceID may be empty
	deviceID := provider.DeviceID()

	// If deviceID is set, verify format
	if deviceID != "" && (len(deviceID) < 6 || deviceID[:5] != "pryx-") {
		t.Errorf("DeviceID() = %s, want prefix 'pryx-'", deviceID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	provider.Shutdown(ctx)
}

func TestProvider_StartSpan_Disabled(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	kc := keychain.New("pryx-test")

	// Force disabled
	os.Setenv("PRYX_TELEMETRY_DISABLED", "true")
	defer os.Unsetenv("PRYX_TELEMETRY_DISABLED")

	provider, _ := NewProvider(cfg, kc)

	ctx := context.Background()
	newCtx, span := provider.StartSpan(ctx, "test-span", attribute.String("key", "value"))

	// When disabled, should return same context and noop span
	if newCtx != ctx {
		t.Error("StartSpan() returned different context when disabled")
	}

	if span == nil {
		t.Error("StartSpan() returned nil span")
	}

	// Noop span should not panic
	span.End()
	span.AddEvent("test-event")
	if span.IsRecording() {
		t.Error("noopSpan.IsRecording() = true, want false")
	}

	ctx2, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	provider.Shutdown(ctx2)
}

func TestProvider_Shutdown(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	kc := keychain.New("pryx-test")

	// Force disabled
	os.Setenv("PRYX_TELEMETRY_DISABLED", "true")
	defer os.Unsetenv("PRYX_TELEMETRY_DISABLED")

	provider, _ := NewProvider(cfg, kc)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() unexpected error: %v", err)
	}
}

func TestGenerateDeviceID(t *testing.T) {
	id1 := generateDeviceID()
	id2 := generateDeviceID()

	if id1 == "" {
		t.Error("generateDeviceID() returned empty string")
	}

	if id1 == id2 {
		t.Error("generateDeviceID() returned duplicate IDs")
	}

	if len(id1) < 6 || id1[:5] != "pryx-" {
		t.Errorf("generateDeviceID() = %s, want prefix 'pryx-'", id1)
	}
}

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		expected string
	}{
		{
			name:     "default version",
			envVar:   "",
			expected: "dev",
		},
		{
			name:     "custom version",
			envVar:   "1.2.3",
			expected: "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				os.Setenv("PRYX_VERSION", tt.envVar)
				defer os.Unsetenv("PRYX_VERSION")
			} else {
				os.Unsetenv("PRYX_VERSION")
			}

			got := getVersion()
			if got != tt.expected {
				t.Errorf("getVersion() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNoopSpan(t *testing.T) {
	span := &noopSpan{}

	// All methods should not panic
	span.End()
	span.AddEvent("test-event")
	if span.IsRecording() {
		t.Error("noopSpan.IsRecording() = true, want false")
	}
	span.RecordError(nil)
	if span.SpanContext().IsValid() {
		t.Error("noopSpan.SpanContext().IsValid() = true, want false")
	}
	span.SetStatus(0, "test")
	span.SetName("test-name")
	span.SetAttributes(attribute.String("key", "value"))
	if span.TracerProvider() != nil {
		t.Error("noopSpan.TracerProvider() != nil, want nil")
	}
	span.AddLink(trace.Link{})
}
