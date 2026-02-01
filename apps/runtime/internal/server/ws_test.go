package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "X-Forwarded-For",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1, 198.51.100.1"},
			want:       "203.0.113.1",
		},
		{
			name:       "X-Real-Ip",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{"X-Real-Ip": "198.51.100.2"},
			want:       "198.51.100.2",
		},
		{
			name:       "RemoteAddr fallback",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{},
			want:       "192.168.1.1",
		},
		{
			name:       "IPv6 RemoteAddr",
			remoteAddr: "[::1]:12345",
			headers:    map[string]string{},
			want:       "[::1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := getClientIP(req)
			if got != tt.want {
				t.Errorf("getClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRateLimiter_AllowUnderLimit(t *testing.T) {
	limiter := getRateLimiter("test-ip", 10) // 10 per second

	// Should allow first 5 requests immediately (burst size)
	for i := 0; i < 5; i++ {
		if !limiter.Allow() {
			t.Errorf("Request %d should be allowed within burst limit", i)
		}
	}
}

func TestRateLimiter_BlockOverLimit(t *testing.T) {
	// Use a very low rate to ensure blocking
	limiter := getRateLimiter("test-block-ip", 1) // 1 per second with burst of 5

	// Consume burst
	for i := 0; i < 5; i++ {
		limiter.Allow()
	}

	// Next request should be blocked
	if limiter.Allow() {
		t.Error("Request should be blocked after burst exhausted")
	}
}

func TestWebSocketBufferSize(t *testing.T) {
	if WebSocketBufferSize != 256 {
		t.Errorf("WebSocketBufferSize = %d, want 256", WebSocketBufferSize)
	}
}

func TestGenerateConnectionID(t *testing.T) {
	ip := "192.168.1.1"
	id1 := generateConnectionID(ip)
	time.Sleep(1 * time.Millisecond)
	id2 := generateConnectionID(ip)

	if id1 == id2 {
		t.Error("Generated connection IDs should be unique")
	}

	if !strings.HasPrefix(id1, ip) {
		t.Errorf("Connection ID should start with IP, got %s", id1)
	}
}

func TestDefaultConstants(t *testing.T) {
	if defaultMaxMessageSize != 10*1024*1024 {
		t.Errorf("defaultMaxMessageSize = %d, want %d", defaultMaxMessageSize, 10*1024*1024)
	}

	if defaultMaxConnections != 1000 {
		t.Errorf("defaultMaxConnections = %d, want 1000", defaultMaxConnections)
	}

	if defaultRateLimitPerMinute != 60 {
		t.Errorf("defaultRateLimitPerMinute = %d, want 60", defaultRateLimitPerMinute)
	}
}
