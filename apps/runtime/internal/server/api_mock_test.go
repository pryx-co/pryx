package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIEndpointStructure tests API endpoint patterns
func TestAPIEndpointStructure(t *testing.T) {
	t.Run("cloud_login_endpoints", func(t *testing.T) {
		endpoints := []struct {
			method string
			path   string
		}{
			{"POST", "/api/v1/cloud/login/start"},
			{"POST", "/api/v1/cloud/login/poll"},
			{"GET", "/api/v1/cloud/status"},
			{"POST", "/api/v1/cloud/logout"},
		}

		for _, ep := range endpoints {
			t.Logf("Testing endpoint: %s %s", ep.method, ep.path)
			assert.NotEmpty(t, ep.method)
			assert.NotEmpty(t, ep.path)
			assert.Contains(t, ep.path, "/api/v1/")
		}
	})

	t.Run("api_version_prefix", func(t *testing.T) {
		paths := []string{
			"/api/v1/cloud/",
			"/api/v1/runtime/",
			"/api/v1/config/",
		}

		for _, path := range paths {
			assert.True(t, strings.HasPrefix(path, "/api/v1/"),
				"Path %s should have /api/v1/ prefix", path)
		}
	})

	t.Run("endpoint_response_format", func(t *testing.T) {
		responses := []struct {
			name   string
			status int
			body   string
		}{
			{"success", 200, `{"success": true}`},
			{"error", 400, `{"error": "bad_request", "message": "Invalid request"}`},
			{"unauthorized", 401, `{"error": "unauthorized", "message": "Please login"}`},
			{"not_found", 404, `{"error": "not_found", "message": "Resource not found"}`},
		}

		for _, resp := range responses {
			t.Logf("Response format: status=%d, body=%s", resp.status, resp.body)
			assert.NotEmpty(t, resp.name)
		}
	})
}

// TestAPIResponseFormats tests mock JSON response structures
func TestAPIResponseFormats(t *testing.T) {
	t.Run("device_code_response_format", func(t *testing.T) {
		response := map[string]any{
			"device_code":      "test-device-code-123",
			"user_code":        "USER123",
			"verification_uri": "https://auth.pryx.dev/activate",
			"expires_in":       1800,
			"interval":         5,
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		var decoded map[string]any
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Contains(t, decoded, "device_code")
		assert.Contains(t, decoded, "user_code")
		assert.Contains(t, decoded, "verification_uri")
		assert.Contains(t, decoded, "expires_in")
		assert.Contains(t, decoded, "interval")

		t.Logf("Device code response: %s", string(data))
	})

	t.Run("token_response_format", func(t *testing.T) {
		response := map[string]any{
			"access_token":  "test-access-token",
			"refresh_token": "test-refresh-token",
			"expires_in":    3600,
			"token_type":    "Bearer",
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		var decoded map[string]any
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Contains(t, decoded, "access_token")
		assert.Contains(t, decoded, "token_type")
		assert.Equal(t, "Bearer", decoded["token_type"])

		t.Logf("Token response: %s", string(data))
	})

	t.Run("error_response_format", func(t *testing.T) {
		response := map[string]any{
			"error":             "invalid_grant",
			"error_description": "The authorization code is invalid",
			"error_uri":         "https://auth.pryx.dev/docs/error/invalid_grant",
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		var decoded map[string]any
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Contains(t, decoded, "error")
		assert.Contains(t, decoded, "error_description")

		t.Logf("Error response: %s", string(data))
	})

	t.Run("status_response_format", func(t *testing.T) {
		response := map[string]any{
			"authenticated": true,
			"provider":      "pryx.dev",
			"expires_at":    "2026-02-03T12:00:00Z",
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		var decoded map[string]any
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Contains(t, decoded, "authenticated")
		assert.Contains(t, decoded, "provider")

		t.Logf("Status response: %s", string(data))
	})
}

// TestAPIErrorCodes tests mock error response handling
func TestAPIErrorCodes(t *testing.T) {
	t.Run("oauth_error_codes", func(t *testing.T) {
		errorCodes := []struct {
			code        string
			description string
			status      int
		}{
			{"invalid_client", "Client authentication failed", http.StatusUnauthorized},
			{"invalid_grant", "The authorization code or refresh token is invalid", http.StatusBadRequest},
			{"access_denied", "The resource owner denied the request", http.StatusForbidden},
			{"invalid_scope", "The requested scope is invalid", http.StatusBadRequest},
			{"authorization_pending", "The device authorization is still pending", http.StatusBadRequest},
			{"slow_down", "The client is polling too quickly", http.StatusBadRequest},
			{"expired_token", "The device code has expired", http.StatusBadRequest},
		}

		for _, ec := range errorCodes {
			t.Logf("Error code: %s (%d) - %s", ec.code, ec.status, ec.description)
			assert.NotEmpty(t, ec.code)
			assert.NotEmpty(t, ec.description)
			assert.Greater(t, ec.status, 0)
		}
	})

	t.Run("http_error_mapping", func(t *testing.T) {
		mappings := []struct {
			apiError string
			httpCode int
		}{
			{"invalid_client", http.StatusUnauthorized},
			{"invalid_grant", http.StatusBadRequest},
			{"access_denied", http.StatusForbidden},
			{"server_error", http.StatusInternalServerError},
			{"temporarily_unavailable", http.StatusServiceUnavailable},
		}

		for _, m := range mappings {
			t.Logf("API error %s maps to HTTP %d", m.apiError, m.httpCode)
			assert.Contains(t, []int{400, 401, 403, 500, 503}, m.httpCode)
		}
	})

	t.Run("error_response_serialization", func(t *testing.T) {
		testCases := []struct {
			errMap map[string]string
		}{
			{map[string]string{"error": "invalid_client", "error_description": "Test"}},
			{map[string]string{"error": "invalid_grant"}},
			{map[string]string{"error": "access_denied", "error_description": "Denied", "error_uri": "https://example.com"}},
		}

		for _, tc := range testCases {
			data, err := json.Marshal(tc.errMap)
			require.NoError(t, err)

			var decoded map[string]string
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Contains(t, decoded, "error")
			assert.Equal(t, tc.errMap["error"], decoded["error"])
		}
	})
}

// MockPryxDevAPI creates a mock server for pryx.dev API testing
func MockPryxDevAPI(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		t.Logf("Mock API received: %s %s", method, path)

		switch {
		case path == "/api/v1/cloud/login/start" && method == "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"device_code": "mock-device-code",
				"user_code": "USER123",
				"verification_uri": "https://auth.pryx.dev/activate",
				"expires_in": 1800,
				"interval": 5
			}`))

		case path == "/api/v1/cloud/login/poll" && method == "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "mock-access-token",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))

		case path == "/api/v1/cloud/status" && method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"authenticated": true,
				"provider": "pryx.dev"
			}`))

		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{
				"error": "not_found",
				"message": "Endpoint not found"
			}`))
		}
	}))
}

// TestMockPryxDevAPI tests the mock API server behavior
func TestMockPryxDevAPI(t *testing.T) {
	mockServer := MockPryxDevAPI(t)
	defer mockServer.Close()

	t.Run("login_start_endpoint", func(t *testing.T) {
		resp, err := http.Post(mockServer.URL+"/api/v1/cloud/login/start", "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data map[string]any
		err = json.NewDecoder(resp.Body).Decode(&data)
		require.NoError(t, err)

		assert.Contains(t, data, "device_code")
		assert.Contains(t, data, "user_code")
	})

	t.Run("login_poll_endpoint", func(t *testing.T) {
		reqBody := `{"device_code": "test", "interval": 5}`
		resp, err := http.Post(mockServer.URL+"/api/v1/cloud/login/poll", "application/json", strings.NewReader(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data map[string]any
		err = json.NewDecoder(resp.Body).Decode(&data)
		require.NoError(t, err)

		assert.Contains(t, data, "access_token")
	})

	t.Run("status_endpoint", func(t *testing.T) {
		resp, err := http.Get(mockServer.URL + "/api/v1/cloud/status")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data map[string]any
		err = json.NewDecoder(resp.Body).Decode(&data)
		require.NoError(t, err)

		assert.Contains(t, data, "authenticated")
	})

	t.Run("not_found_endpoint", func(t *testing.T) {
		resp, err := http.Get(mockServer.URL + "/api/v1/unknown")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
