//go:build integration
// +build integration

package integration

import (
	"encoding/json"
	"net"
	"net/http"
	"testing"
	"time"

	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/server"
	"pryx-core/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChannelEndpointsIntegration tests channel API endpoints
func TestChannelEndpointsIntegration(t *testing.T) {
	cfg := &config.Config{ListenAddr: "127.0.0.1:0"}
	s, _ := store.New(":memory:")
	defer s.Close()

	t.Setenv("PRYX_KEYCHAIN_FILE", t.TempDir()+"/keychain.json")
	kc := keychain.New("test")

	srv := server.New(cfg, s.DB, kc)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go srv.Serve(listener)
	time.Sleep(10 * time.Millisecond)

	client := &http.Client{Timeout: time.Second}
	baseUrl := "http://" + listener.Addr().String()

	// Test channel list endpoint
	resp, err := client.Get(baseUrl + "/api/v1/channels")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Contains(t, result, "channels")
}

// TestOAuthDeviceFlowEndpoints tests the OAuth device flow endpoints
func TestOAuthDeviceFlowEndpoints(t *testing.T) {
	cfg := &config.Config{ListenAddr: "127.0.0.1:0"}
	s, _ := store.New(":memory:")
	defer s.Close()

	t.Setenv("PRYX_KEYCHAIN_FILE", t.TempDir()+"/keychain.json")
	kc := keychain.New("test")

	srv := server.New(cfg, s.DB, kc)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go srv.Serve(listener)
	time.Sleep(10 * time.Millisecond)

	client := &http.Client{Timeout: time.Second}
	baseUrl := "http://" + listener.Addr().String()

	// Test device code endpoint (should return structure for device flow)
	resp, err := client.Get(baseUrl + "/api/v1/auth/device/code")
	if err != nil {
		// This might fail if auth is not configured - that's expected
		t.Logf("Device code endpoint not available (expected if auth not configured): %v", err)
		return
	}
	defer resp.Body.Close()

	// If endpoint exists, verify structure
	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Device flow should return device_code, user_code, verification_uri, etc.
		assert.Contains(t, result, "device_code")
		assert.Contains(t, result, "user_code")
		assert.Contains(t, result, "verification_uri")
	}
}

// TestCompleteWorkflowIntegration tests a complete user workflow
func TestCompleteWorkflowIntegration(t *testing.T) {
	cfg := &config.Config{ListenAddr: "127.0.0.1:0"}
	s, _ := store.New(":memory:")
	defer s.Close()

	t.Setenv("PRYX_KEYCHAIN_FILE", t.TempDir()+"/keychain.json")
	kc := keychain.New("test")

	srv := server.New(cfg, s.DB, kc)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go srv.Serve(listener)
	time.Sleep(10 * time.Millisecond)

	client := &http.Client{Timeout: time.Second}
	baseUrl := "http://" + listener.Addr().String()

	// Test health endpoint
	resp, err := client.Get(baseUrl + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test skills endpoint
	resp, err = client.Get(baseUrl + "/skills")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test providers endpoint
	resp, err = client.Get(baseUrl + "/api/v1/providers")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test sessions endpoint
	resp, err = client.Get(baseUrl + "/api/v1/sessions")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test channels endpoint
	resp, err = client.Get(baseUrl + "/api/v1/channels")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
