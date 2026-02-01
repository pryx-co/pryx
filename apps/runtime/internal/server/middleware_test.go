package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pryx-core/internal/config"
)

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: []string{"https://example.com", "https://app.example.com"},
	}

	handler := corsMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	allowedOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	if allowedOrigin != "https://example.com" {
		t.Errorf("Expected origin https://example.com, got %s", allowedOrigin)
	}

	credentials := rr.Header().Get("Access-Control-Allow-Credentials")
	if credentials != "true" {
		t.Errorf("Expected credentials true, got %s", credentials)
	}
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: []string{"https://example.com"},
	}

	handler := corsMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	allowedOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	if allowedOrigin != "" {
		t.Errorf("Expected no CORS headers for disallowed origin, got %s", allowedOrigin)
	}

	credentials := rr.Header().Get("Access-Control-Allow-Credentials")
	if credentials != "" {
		t.Errorf("Expected no credentials header for disallowed origin, got %s", credentials)
	}
}

func TestCORSMiddleware_PreflightAllowed(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: []string{"https://example.com"},
	}

	handler := corsMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for preflight, got %d", rr.Code)
	}
}

func TestCORSMiddleware_PreflightDisallowed(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: []string{"https://example.com"},
	}

	handler := corsMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for disallowed preflight, got %d", rr.Code)
	}
}

func TestCORSMiddleware_LocalhostAlwaysAllowed(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: []string{},
	}

	handler := corsMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	testCases := []string{
		"http://localhost:3000",
		"https://localhost:3000",
		"http://localhost:8080",
	}

	for _, origin := range testCases {
		t.Run(origin, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", origin)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rr.Code)
			}

			allowedOrigin := rr.Header().Get("Access-Control-Allow-Origin")
			if allowedOrigin != origin {
				t.Errorf("Expected origin %s, got %s", origin, allowedOrigin)
			}
		})
	}
}

func TestCORSMiddleware_EmptyOriginAllowed(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: []string{"https://example.com"},
	}

	handler := corsMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for empty origin, got %d", rr.Code)
	}

	// No CORS headers should be set for empty origin
	allowedOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	if allowedOrigin != "" {
		t.Errorf("Expected no CORS headers for empty origin, got %s", allowedOrigin)
	}
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		allowed []string
		want    bool
	}{
		{
			name:    "exact match",
			origin:  "https://example.com",
			allowed: []string{"https://example.com"},
			want:    true,
		},
		{
			name:    "no match",
			origin:  "https://evil.com",
			allowed: []string{"https://example.com"},
			want:    false,
		},
		{
			name:    "localhost http",
			origin:  "http://localhost:3000",
			allowed: []string{},
			want:    true,
		},
		{
			name:    "localhost https",
			origin:  "https://localhost:3000",
			allowed: []string{},
			want:    true,
		},
		{
			name:    "empty origin",
			origin:  "",
			allowed: []string{"https://example.com"},
			want:    true,
		},
		{
			name:    "multiple origins with match",
			origin:  "https://app.example.com",
			allowed: []string{"https://example.com", "https://app.example.com"},
			want:    true,
		},
		{
			name:    "case insensitive",
			origin:  "https://EXAMPLE.COM",
			allowed: []string{"https://example.com"},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOriginAllowed(tt.origin, tt.allowed)
			if got != tt.want {
				t.Errorf("isOriginAllowed(%q, %v) = %v, want %v", tt.origin, tt.allowed, got, tt.want)
			}
		})
	}
}
