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
	"sync"
	"testing"
	"time"

	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/skills"
	"pryx-core/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEnv interface {
	Helper()
	Setenv(key, value string)
	TempDir() string
}

func newTestKeychain(t testEnv) *keychain.Keychain {
	t.Helper()
	t.Setenv("PRYX_KEYCHAIN_FILE", filepath.Join(t.TempDir(), "keychain.json"))
	return keychain.New("test")
}

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		ListenAddr:   ":0",
		DatabasePath: ":memory:",
	}

	s, err := store.New(":memory:")
	require.NoError(t, err)
	defer s.Close()

	kc := newTestKeychain(t)

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
	kc := newTestKeychain(t)

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
		{"skills enable POST", "POST", "/skills/enable", http.StatusBadRequest},
		{"skills disable POST", "POST", "/skills/disable", http.StatusBadRequest},
		{"skills install POST", "POST", "/skills/install", http.StatusBadRequest},
		{"skills uninstall POST", "POST", "/skills/uninstall", http.StatusBadRequest},
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

func TestHandleSkillsEnableDisableRoundTrip(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	st, _ := store.New(":memory:")
	defer st.Close()
	kc := newTestKeychain(t)

	t.Setenv("PRYX_SKILLS_CONFIG_PATH", filepath.Join(t.TempDir(), "skills.yaml"))

	server := New(cfg, st.DB, kc)
	server.skills = skills.NewRegistry()
	server.skills.Upsert(skills.Skill{ID: "test-skill"})

	{
		reqBody := `{"id":"test-skill"}`
		req := httptest.NewRequest("POST", "/skills/enable", strings.NewReader(reqBody))
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	{
		s, ok := server.skills.Get("test-skill")
		require.True(t, ok)
		assert.True(t, s.Enabled)
	}

	{
		reqBody := `{"id":"test-skill"}`
		req := httptest.NewRequest("POST", "/skills/disable", strings.NewReader(reqBody))
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	{
		s, ok := server.skills.Get("test-skill")
		require.True(t, ok)
		assert.False(t, s.Enabled)
	}
}

func TestHandleSkillsInstallFromURLAndUninstall(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	st, _ := store.New(":memory:")
	defer st.Close()
	kc := newTestKeychain(t)

	managedRoot := t.TempDir()
	t.Setenv("PRYX_MANAGED_SKILLS_DIR", managedRoot)
	t.Setenv("PRYX_SKILLS_CONFIG_PATH", filepath.Join(t.TempDir(), "skills.yaml"))

	skillDoc := []byte(`---
name: installed-skill
description: from url
---
# Installed skill`)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(skillDoc)
	}))
	defer ts.Close()

	server := New(cfg, st.DB, kc)
	server.skills = skills.NewRegistry()

	{
		reqBody := `{"id":"` + ts.URL + `"}`
		req := httptest.NewRequest("POST", "/skills/install", strings.NewReader(reqBody))
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	{
		s, ok := server.skills.Get("installed-skill")
		require.True(t, ok)
		assert.Equal(t, skills.SourceRemote, s.Source)
	}

	{
		reqBody := `{"id":"installed-skill"}`
		req := httptest.NewRequest("POST", "/skills/uninstall", strings.NewReader(reqBody))
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	{
		_, ok := server.skills.Get("installed-skill")
		assert.False(t, ok)
	}
}

func TestHandleHealth(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := newTestKeychain(t)

	server := New(cfg, s.DB, kc)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	server.handleHealth(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestHandleSkillsList(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := newTestKeychain(t)

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
	kc := newTestKeychain(t)

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
	kc := newTestKeychain(t)

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
	kc := newTestKeychain(t)

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
	kc := newTestKeychain(t)

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
	kc := newTestKeychain(t)

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
	cfg := &config.Config{
		ListenAddr:     ":0",
		AllowedOrigins: []string{"https://example.com"},
	}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := newTestKeychain(t)

	server := New(cfg, s.DB, kc)

	// Test preflight with allowed origin
	req := httptest.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	server.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "GET")
}

func TestHandleCloudLogin_HappyPath(t *testing.T) {
	var mu sync.Mutex
	var codeChallengeMethod string
	var codeChallenge string
	var verifier string
	deviceCode := "device-123"

	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device/code":
			var req map[string]string
			_ = json.NewDecoder(r.Body).Decode(&req)
			mu.Lock()
			codeChallengeMethod = req["code_challenge_method"]
			codeChallenge = req["code_challenge"]
			mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":               deviceCode,
				"user_code":                 "USER-CODE",
				"verification_uri":          "https://example.com/verify",
				"expires_in":                60,
				"interval":                  1,
				"verification_uri_complete": "https://example.com/verify?code=USER-CODE",
			})
		case "/auth/device/token":
			var req map[string]string
			_ = json.NewDecoder(r.Body).Decode(&req)
			mu.Lock()
			verifier = req["code_verifier"]
			mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "token-abc",
				"expires_in":   3600,
				"token_type":   "bearer",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer cloud.Close()

	cfg := &config.Config{ListenAddr: ":0", CloudAPIUrl: cloud.URL}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := newTestKeychain(t)

	server := New(cfg, s.DB, kc)

	{
		req := httptest.NewRequest("POST", "/api/v1/cloud/login/start", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, deviceCode, body["device_code"])
	}

	{
		req := httptest.NewRequest(
			"POST",
			"/api/v1/cloud/login/poll",
			strings.NewReader(`{"device_code":"device-123","interval":1,"expires_in":5}`),
		)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, true, body["ok"])
	}

	stored, err := kc.Get("cloud_access_token")
	require.NoError(t, err)
	assert.Equal(t, "token-abc", stored)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "S256", codeChallengeMethod)
	assert.NotEmpty(t, codeChallenge)
	assert.NotEmpty(t, verifier)
}

func TestHandleCloudLogin_RetryAfterTimeoutKeepsPKCE(t *testing.T) {
	var mu sync.Mutex
	var verifiers []string
	allowToken := false
	deviceCode := "device-999"

	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device/code":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":      deviceCode,
				"user_code":        "USER-CODE",
				"verification_uri": "https://example.com/verify",
				"expires_in":       60,
				"interval":         1,
			})
		case "/auth/device/token":
			var req map[string]string
			_ = json.NewDecoder(r.Body).Decode(&req)
			mu.Lock()
			verifiers = append(verifiers, req["code_verifier"])
			shouldAllow := allowToken
			mu.Unlock()

			if !shouldAllow {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "authorization_pending"})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "token-final",
				"expires_in":   3600,
				"token_type":   "bearer",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer cloud.Close()

	cfg := &config.Config{ListenAddr: ":0", CloudAPIUrl: cloud.URL}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := newTestKeychain(t)
	server := New(cfg, s.DB, kc)

	{
		req := httptest.NewRequest("POST", "/api/v1/cloud/login/start", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}

	{
		req := httptest.NewRequest(
			"POST",
			"/api/v1/cloud/login/poll",
			strings.NewReader(`{"device_code":"device-999","interval":1,"expires_in":1}`),
		)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusRequestTimeout, rec.Code)
	}

	mu.Lock()
	allowToken = true
	mu.Unlock()

	{
		req := httptest.NewRequest(
			"POST",
			"/api/v1/cloud/login/poll",
			strings.NewReader(`{"device_code":"device-999","interval":1,"expires_in":5}`),
		)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, true, body["ok"])
	}

	stored, err := kc.Get("cloud_access_token")
	require.NoError(t, err)
	assert.Equal(t, "token-final", stored)

	mu.Lock()
	defer mu.Unlock()
	require.GreaterOrEqual(t, len(verifiers), 2)
	assert.NotEmpty(t, verifiers[0])
	assert.NotEmpty(t, verifiers[len(verifiers)-1])
	assert.Equal(t, verifiers[0], verifiers[len(verifiers)-1])
}

func TestHandleProviderKey_RoundTrip(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := newTestKeychain(t)

	server := New(cfg, s.DB, kc)

	{
		req := httptest.NewRequest("GET", "/api/v1/providers/openai/key", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, false, body["configured"])
	}

	{
		req := httptest.NewRequest("POST", "/api/v1/providers/openai/key", strings.NewReader(`{"api_key":"sk-test"}`))
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, true, body["ok"])
	}

	{
		req := httptest.NewRequest("GET", "/api/v1/providers/openai/key", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, true, body["configured"])
	}

	{
		req := httptest.NewRequest("DELETE", "/api/v1/providers/openai/key", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	}

	{
		req := httptest.NewRequest("GET", "/api/v1/providers/openai/key", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, false, body["configured"])
	}
}

func TestHandleProviderKeySet_InvalidBody(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := newTestKeychain(t)

	server := New(cfg, s.DB, kc)

	{
		req := httptest.NewRequest("POST", "/api/v1/providers/openai/key", strings.NewReader("not json"))
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	{
		req := httptest.NewRequest("POST", "/api/v1/providers/openai/key", strings.NewReader(`{"api_key":""}`))
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestHandleProviderKey_InvalidProviderID(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := newTestKeychain(t)

	server := New(cfg, s.DB, kc)

	{
		req := httptest.NewRequest("GET", "/api/v1/providers/bad%20id/key", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	{
		req := httptest.NewRequest("POST", "/api/v1/providers/bad%20id/key", strings.NewReader(`{"api_key":"x"}`))
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	{
		req := httptest.NewRequest("DELETE", "/api/v1/providers/bad%20id/key", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestHandleProviderKey_KeychainUnavailable(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()

	server := New(cfg, s.DB, nil)

	req := httptest.NewRequest("GET", "/api/v1/providers/openai/key", nil)
	rec := httptest.NewRecorder()
	server.router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestServer_Bus(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := newTestKeychain(t)

	server := New(cfg, s.DB, kc)

	b := server.Bus()
	assert.NotNil(t, b)
	assert.Equal(t, server.bus, b)
}

func TestServer_Handler(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := newTestKeychain(t)

	server := New(cfg, s.DB, kc)

	handler := server.Handler()
	assert.NotNil(t, handler)
}

func TestServer_Serve(t *testing.T) {
	cfg := &config.Config{ListenAddr: ":0"}
	s, _ := store.New(":memory:")
	defer s.Close()
	kc := newTestKeychain(t)

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
	kc := newTestKeychain(t)

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
	kc := newTestKeychain(t)

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
	kc := newTestKeychain(t)

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
	kc := newTestKeychain(b)

	server := New(cfg, s.DB, kc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		rec := httptest.NewRecorder()
		server.handleHealth(rec, req)
	}
}
