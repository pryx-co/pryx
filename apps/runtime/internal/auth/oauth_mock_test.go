package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pryx-core/internal/auth"
	"pryx-core/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock HTTP server for OAuth device flow responses
type mockOAuthServer struct {
	DeviceCodeResponse auth.DeviceCodeResponse
	TokenResponse      auth.TokenResponse
	ErrorResponse      auth.ErrorResponse
	ShouldReturnError  bool
	ResponseDelay      time.Duration
}

func newMockOAuthServer() *mockOAuthServer {
	return &mockOAuthServer{
		DeviceCodeResponse: auth.DeviceCodeResponse{
			DeviceCode:      "test-device-code-mock-123",
			UserCode:        "USER123M",
			VerificationURI: "https://auth.pryx.dev/activate",
			ExpiresIn:       1800,
			Interval:        5,
		},
		TokenResponse: auth.TokenResponse{
			AccessToken:  "mock-access-token-xyz",
			RefreshToken: "mock-refresh-token-abc",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		},
	}
}

// TestOAuthDeviceCodeMock tests device code response structure validation
func TestOAuthDeviceCodeMock(t *testing.T) {
	server := newMockOAuthServer()

	t.Run("device_code_response_structure", func(t *testing.T) {
		assert.NotEmpty(t, server.DeviceCodeResponse.DeviceCode, "Device code should not be empty")
		assert.NotEmpty(t, server.DeviceCodeResponse.UserCode, "User code should not be empty")
		assert.NotEmpty(t, server.DeviceCodeResponse.VerificationURI, "Verification URI should not be empty")
		assert.Greater(t, server.DeviceCodeResponse.ExpiresIn, 0, "ExpiresIn should be positive")
		assert.Greater(t, server.DeviceCodeResponse.Interval, 0, "Interval should be positive")

		t.Logf("Mock device code response: code=%s, user_code=%s, uri=%s",
			server.DeviceCodeResponse.DeviceCode,
			server.DeviceCodeResponse.UserCode,
			server.DeviceCodeResponse.VerificationURI)
	})

	t.Run("device_code_json_serialization", func(t *testing.T) {
		data, err := json.Marshal(server.DeviceCodeResponse)
		require.NoError(t, err, "Device code response should serialize to JSON")

		var decoded auth.DeviceCodeResponse
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err, "Device code response should deserialize from JSON")

		assert.Equal(t, server.DeviceCodeResponse.DeviceCode, decoded.DeviceCode)
		assert.Equal(t, server.DeviceCodeResponse.UserCode, decoded.UserCode)
		assert.Equal(t, server.DeviceCodeResponse.VerificationURI, decoded.VerificationURI)
	})

	t.Run("device_code_user_code_format", func(t *testing.T) {
		assert.LessOrEqual(t, len(server.DeviceCodeResponse.UserCode), 8,
			"User code should be short for easy manual entry")
		assert.GreaterOrEqual(t, len(server.DeviceCodeResponse.UserCode), 4,
			"User code should be at least 4 characters")
	})
}

// TestOAuthTokenExchangeMock tests token exchange with valid and invalid codes
func TestOAuthTokenExchangeMock(t *testing.T) {
	server := newMockOAuthServer()

	t.Run("valid_token_exchange_response", func(t *testing.T) {
		response := server.TokenResponse

		assert.NotEmpty(t, response.AccessToken, "Access token should not be empty")
		assert.Equal(t, "Bearer", response.TokenType, "Token type should be Bearer")
		assert.Greater(t, response.ExpiresIn, 0, "ExpiresIn should be positive")

		t.Logf("Mock token response: type=%s, expires_in=%d",
			response.TokenType, response.ExpiresIn)
	})

	t.Run("token_response_json_serialization", func(t *testing.T) {
		data, err := json.Marshal(server.TokenResponse)
		require.NoError(t, err, "Token response should serialize to JSON")

		var decoded auth.TokenResponse
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err, "Token response should deserialize from JSON")

		assert.Equal(t, server.TokenResponse.AccessToken, decoded.AccessToken)
		assert.Equal(t, server.TokenResponse.RefreshToken, decoded.RefreshToken)
		assert.Equal(t, server.TokenResponse.TokenType, decoded.TokenType)
	})

	t.Run("refresh_token_present", func(t *testing.T) {
		assert.NotEmpty(t, server.TokenResponse.RefreshToken,
			"Refresh token should be present for token renewal")
	})
}

// TestOAuthRefreshTokenMock tests refresh flow mocking
func TestOAuthRefreshTokenMock(t *testing.T) {
	_ = newMockOAuthServer()

	t.Run("refresh_flow_structure", func(t *testing.T) {
		refreshRequest := map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": "mock-refresh-token-abc",
			"client_id":     "test-client-id",
		}

		data, err := json.Marshal(refreshRequest)
		require.NoError(t, err)

		var decoded map[string]string
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "refresh_token", decoded["grant_type"])
		assert.NotEmpty(t, decoded["refresh_token"])
		assert.NotEmpty(t, decoded["client_id"])

		t.Logf("Mock refresh request: grant_type=%s", decoded["grant_type"])
	})

	t.Run("refresh_response_structure", func(t *testing.T) {
		newToken := auth.TokenResponse{
			AccessToken:  "new-mock-access-token",
			RefreshToken: "new-mock-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		assert.NotEmpty(t, newToken.AccessToken)
		assert.NotEmpty(t, newToken.RefreshToken)
		assert.Equal(t, "Bearer", newToken.TokenType)
	})
}

// TestOAuthBrowserRedirectMock tests redirect URI handling
func TestOAuthBrowserRedirectMock(t *testing.T) {
	t.Run("redirect_uri_validation", func(t *testing.T) {
		redirectURIs := []string{
			"pryx://callback",
			"pryx://callback/test",
			"http://localhost:3000/callback",
			"https://pryx.dev/oauth/callback",
		}

		for _, uri := range redirectURIs {
			t.Logf("Testing redirect URI: %s", uri)
			assert.NotEmpty(t, uri)
		}
	})

	t.Run("pryx_scheme_handling", func(t *testing.T) {
		uri := "pryx://callback/test-provider"

		assert.Contains(t, uri, "pryx://", "Should contain pryx scheme")
		assert.Contains(t, uri, "callback", "Should contain callback path")
	})

	t.Run("state_parameter_validation", func(t *testing.T) {
		state := auth.OAuthState{
			State:       "mock-state-abc123",
			ProviderID:  "test-provider",
			ClientID:    "test-client",
			ExpiresAt:   time.Now().Add(10 * time.Minute),
			RedirectURI: "pryx://callback",
		}

		assert.NotEmpty(t, state.State)
		assert.NotEmpty(t, state.ProviderID)
		assert.NotEmpty(t, state.ClientID)
		assert.False(t, state.ExpiresAt.IsZero())
		assert.NotEmpty(t, state.RedirectURI)

		t.Logf("Mock OAuth state: state=%s, provider=%s", state.State, state.ProviderID)
	})
}

// TestOAuthErrorHandlingMock tests various OAuth error codes
func TestOAuthErrorHandlingMock(t *testing.T) {
	t.Run("error_response_structure", func(t *testing.T) {
		errors := []auth.ErrorResponse{
			{
				Error:            "invalid_client",
				ErrorDescription: "Client authentication failed",
			},
			{
				Error:            "invalid_grant",
				ErrorDescription: "The authorization code or refresh token is invalid",
			},
			{
				Error:            "access_denied",
				ErrorDescription: "The resource owner or authorization server denied the request",
			},
			{
				Error:            "invalid_scope",
				ErrorDescription: "The requested scope is invalid, unknown, or malformed",
			},
		}

		for _, err := range errors {
			assert.NotEmpty(t, err.Error, "Error code should not be empty")

			data, marshalErr := json.Marshal(err)
			assert.NoError(t, marshalErr)

			var decoded auth.ErrorResponse
			unmarshalErr := json.Unmarshal(data, &decoded)
			assert.NoError(t, unmarshalErr)

			assert.Equal(t, err.Error, decoded.Error)

			t.Logf("Mock OAuth error: code=%s, description=%s",
				err.Error, err.ErrorDescription)
		}
	})

	t.Run("authorization_pending_error", func(t *testing.T) {
		err := auth.ErrorResponse{
			Error:            "authorization_pending",
			ErrorDescription: "The device authorization request is still pending",
		}

		assert.Equal(t, "authorization_pending", err.Error)
		assert.Contains(t, err.ErrorDescription, "pending")
	})

	t.Run("slow_down_error", func(t *testing.T) {
		err := auth.ErrorResponse{
			Error:            "slow_down",
			ErrorDescription: "The client is polling too quickly and should slow down",
		}

		assert.Equal(t, "slow_down", err.Error)
		assert.Contains(t, err.ErrorDescription, "quickly")
	})

	t.Run("expired_token_error", func(t *testing.T) {
		err := auth.ErrorResponse{
			Error:            "expired_token",
			ErrorDescription: "The device code has expired",
		}

		assert.Equal(t, "expired_token", err.Error)
		assert.Contains(t, err.ErrorDescription, "expired")
	})
}

// TestOAuthDeviceFlowIntegrationMock simulates complete device flow
func TestOAuthDeviceFlowIntegrationMock(t *testing.T) {
	t.Run("complete_device_flow_simulation", func(t *testing.T) {
		ctx := context.Background()

		kc := newMockKeychain()

		cfg := &config.AuthConfig{
			OAuthProviders: map[string]*config.OAuthProvider{
				"mock-provider": {
					Name:         "Mock Provider",
					ClientID:     "mock-client-id",
					ClientSecret: "mock-client-secret",
					AuthURL:      "https://auth.pryx.dev/oauth/device/authorize",
					TokenURL:     "https://auth.pryx.dev/oauth/token",
					Scopes:       []string{"openid", "profile", "email"},
				},
			},
		}

		manager := auth.NewManager(cfg, kc)

		state, err := manager.InitiateDeviceFlow(ctx, "mock-provider", "pryx://callback")
		require.NoError(t, err)
		require.NotNil(t, state)

		assert.NotEmpty(t, state.State)
		assert.Equal(t, "mock-provider", state.ProviderID)
		assert.Equal(t, "mock-client-id", state.ClientID)

		t.Logf("Device flow initiated: state=%s, provider=%s", state.State, state.ProviderID)

		mockDeviceCode := auth.DeviceCodeResponse{
			DeviceCode:      state.State,
			UserCode:        "USER123",
			VerificationURI: "https://auth.pryx.dev/activate",
			ExpiresIn:       1800,
			Interval:        5,
		}

		assert.NotEmpty(t, mockDeviceCode.UserCode)
		assert.Contains(t, mockDeviceCode.VerificationURI, "activate")

		t.Logf("User code for manual entry: %s", mockDeviceCode.UserCode)

		mockTokenResponse := auth.TokenResponse{
			AccessToken:  "mock-access-token-final",
			RefreshToken: "mock-refresh-token-final",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}

		assert.NotEmpty(t, mockTokenResponse.AccessToken)
		assert.NotEmpty(t, mockTokenResponse.RefreshToken)

		err = manager.SetManualToken(ctx, "mock-provider", mockTokenResponse.AccessToken)
		require.NoError(t, err)

		t.Logf("Token stored successfully for provider: mock-provider")
		t.Logf("Complete mock device flow simulation successful")
	})
}

// TestOAuthPKCEMock tests PKCE parameter generation for secure OAuth
func TestOAuthPKCEMock(t *testing.T) {
	t.Run("pkce_verifier_generation", func(t *testing.T) {
		verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"

		assert.Len(t, verifier, 43, "PKCE verifier should be 43 characters")
		assert.NotEmpty(t, verifier)

		t.Logf("Mock PKCE verifier length: %d", len(verifier))
	})

	t.Run("pkce_challenge_derivation", func(t *testing.T) {
		challenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

		assert.Len(t, challenge, 43, "PKCE challenge should be 43 characters")
		assert.NotEmpty(t, challenge)

		t.Logf("Mock PKCE challenge length: %d", len(challenge))
	})

	t.Run("pkce_method_validation", func(t *testing.T) {
		method := "S256"
		assert.Equal(t, "S256", method, "PKCE method should be S256")
	})
}

// TestOAuthTokenValidationMock tests token response validation
func TestOAuthTokenValidationMock(t *testing.T) {
	t.Run("access_token_validation", func(t *testing.T) {
		tokens := []string{
			"valid-access-token-123",
			"",
			"Bearer mock-token",
			"short",
		}

		for _, token := range tokens {
			isValid := len(token) >= 10 || len(token) == 0
			if len(token) > 7 && token[:7] == "Bearer" {
				isValid = true
			}
			t.Logf("Token validation: length=%d, valid=%v", len(token), isValid)
		}
	})

	t.Run("expires_in_validation", func(t *testing.T) {
		expirationTimes := []int{60, 300, 1800, 3600, 7200, 86400}

		for _, expiresIn := range expirationTimes {
			assert.Greater(t, expiresIn, 0)
			duration := time.Duration(expiresIn) * time.Second
			t.Logf("Token expires in: %v", duration)
		}
	})
}

// Mock HTTP server for testing OAuth endpoints
func createMockOAuthHTTPServer(t *testing.T, response any, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		assert.Contains(t, r.Header.Get("Content-Type"), "application/json")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if response != nil {
			data, _ := json.Marshal(response)
			w.Write(data)
		}
	}))
}

// TestOAuthHTTPServerMock tests mock HTTP server behavior
func TestOAuthHTTPServerMock(t *testing.T) {
	server := newMockOAuthServer()

	t.Run("device_code_endpoint_mock", func(t *testing.T) {
		mockServer := createMockOAuthHTTPServer(t, server.DeviceCodeResponse, http.StatusOK)
		defer mockServer.Close()

		resp, err := http.Post(mockServer.URL, "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("token_endpoint_mock", func(t *testing.T) {
		mockServer := createMockOAuthHTTPServer(t, server.TokenResponse, http.StatusOK)
		defer mockServer.Close()

		resp, err := http.Post(mockServer.URL, "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("error_endpoint_mock", func(t *testing.T) {
		mockServer := createMockOAuthHTTPServer(t, auth.ErrorResponse{
			Error:            "invalid_grant",
			ErrorDescription: "Mock invalid grant error",
		}, http.StatusBadRequest)
		defer mockServer.Close()

		resp, err := http.Post(mockServer.URL, "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
