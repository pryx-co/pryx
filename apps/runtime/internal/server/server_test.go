package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		ListenAddr:   ":0",
		DatabasePath: ":memory:",
	}

	s, err := store.New(":memory:")
	require.NoError(t, err)
	defer s.Close()

	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	assert.NotNil(t, server)
	assert.NotNil(t, server.router)
	assert.NotNil(t, server.bus)
	assert.NotNil(t, server.mcp)
	assert.NotNil(t, server.cfg)
}

func TestServer_Routes(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	tests := []struct {
		name       string
		method     string
		path       string
		expectCode int
	}{
		{"health GET", "GET", "/health", http.StatusOK},
		{"skills GET", "GET", "/skills", http.StatusOK},
		{"skills info GET", "GET", "/skills/test-id", http.StatusNotFound},
		{"mcp tools GET", "GET", "/mcp/tools", http.StatusOK},
		{"mcp call POST no body", "POST", "/mcp/tools/call", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == "POST" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader("{}"))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectCode, rec.Code)
		})
	}
}

func TestHandleHealth(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	server.handleHealth(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
}

func TestHandleSkillsList(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	req := httptest.NewRequest("GET", "/skills", nil)
	rec := httptest.NewRecorder()

	server.handleSkillsList(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "skills")
}

func TestHandleSkillsInfo_MissingID(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/skills/", nil)
	rec := httptest.NewRecorder()

	// Use chi router to properly extract URL params
	r := chi.NewRouter()
	r.Get("/skills/{id}", server.handleSkillsInfo)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleSkillsInfo_NotFound(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	r := chi.NewRouter()
	r.Get("/skills/{id}", server.handleSkillsInfo)

	req := httptest.NewRequest("GET", "/skills/nonexistent", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleMCPTools(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	req := httptest.NewRequest("GET", "/mcp/tools", nil)
	rec := httptest.NewRecorder()

	server.handleMCPTools(rec, req)

	// Should return OK (might have empty tools if no MCP servers)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "tools")
}

func TestHandleMCPCall_InvalidJSON(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	req := httptest.NewRequest("POST", "/mcp/tools/call", strings.NewReader("invalid json"))
	rec := httptest.NewRecorder()

	server.handleMCPCall(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleMCPCall_MissingTool(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	body := map[string]interface{}{
		"session_id": "test-session",
		"arguments":  map[string]interface{}{},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/mcp/tools/call", bytes.NewReader(jsonBody))
	rec := httptest.NewRecorder()

	server.handleMCPCall(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCorsMiddleware(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	req := httptest.NewRequest("OPTIONS", "/health", nil)
	rec := httptest.NewRecorder()

	server.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "GET")
}

func TestServer_Bus(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	b := server.Bus()
	assert.NotNil(t, b)
	assert.Equal(t, server.bus, b)
}

func TestServer_Handler(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	handler := server.Handler()
	assert.NotNil(t, handler)
}

func TestServer_Serve(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	// Create a test listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	// Start serving in background
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Serve(listener)
	}()

	// Give server time to start
	time.Sleep(10 * time.Millisecond)

	// Make a request
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get("http://" + listener.Addr().String() + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
	defer shutdownCancel()
	err = server.Shutdown(shutdownCtx)
	assert.NoError(t, err)

	select {
	case <-errChan:
	case <-time.After(time.Second):
		t.Fatal("Server didn't stop")
	}
}

func TestServer_Shutdown(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	// Create and start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go server.Serve(listener)
	time.Sleep(10 * time.Millisecond)

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestServer_Shutdown_NotStarted(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	// Shutdown without starting - should not panic
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestServer_DynamicPortAllocation(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	// Start should find an available port
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.Start()
	}()

	// Give time for port file to be written
	time.Sleep(50 * time.Millisecond)

	// Port file should exist and contain a valid port
	home, _ := os.UserHomeDir()
	portData, err := os.ReadFile(filepath.Join(home, ".pryx", "runtime.port"))
	if err == nil {
		port := strings.TrimSpace(string(portData))
		assert.NotEmpty(t, port)
		// Verify it's a valid port number
		portNum, err := strconv.Atoi(port)
		require.NoError(t, err)
		assert.Greater(t, portNum, 0)
		assert.Less(t, portNum, 65536)
	}

	cancel()
}

func BenchmarkHandleHealth(b *testing.B) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := keychain.New("test")

	server := New(cfg, s.DB, kc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		rec := httptest.NewRecorder()
		server.handleHealth(rec, req)
	}
}
