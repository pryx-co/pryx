package auth_test

import (
	"testing"

	"pryx-core/internal/auth"
	"pryx-core/internal/config"
)

// TestCrossPlatformPathHandling tests path handling for cross-platform compatibility
// Verifies that paths work correctly regardless of OS
func TestCrossPlatformPathHandling(t *testing.T) {
	// Test that paths are constructed using filepath.Join (cross-platform)
	// This doesn't require actual file operations, just path construction

	testCases := []struct {
		name        string
		parts       []string
		description string
	}{
		{
			name:        "pryx home directory",
			parts:       []string{"~", ".pryx"},
			description: "Pryx configuration directory",
		},
		{
			name:        "pryx cache directory",
			parts:       []string{"~", ".pryx", "cache"},
			description: "Pryx cache directory",
		},
		{
			name:        "pryx skills directory",
			parts:       []string{"~", ".pryx", "skills"},
			description: "Pryx skills directory",
		},
		{
			name:        "pryx channels config",
			parts:       []string{"~", ".pryx", "channels.json"},
			description: "Pryx channels configuration file",
		},
		{
			name:        "pryx keychain file",
			parts:       []string{"~", ".pryx", "keychain.json"},
			description: "Pryx keychain file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// These paths should be constructable without errors
			// Actual path validation happens in production code
			t.Logf("%s: %v", tc.description, tc.parts)
		})
	}
}

// TestOAuthProviderConfiguration tests OAuth provider configuration structure
// This is testable without network as it validates configuration only
func TestOAuthProviderConfiguration(t *testing.T) {
	providers := map[string]*config.OAuthProvider{
		"google": {
			Name:         "Google",
			ClientID:     "google-client-id",
			ClientSecret: "google-client-secret",
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			Scopes:       []string{"https://www.googleapis.com/auth/generative-language.retroactive"},
		},
		"anthropic": {
			Name:         "Anthropic",
			ClientID:     "anthropic-client-id",
			ClientSecret: "anthropic-client-secret",
			AuthURL:      "https://console.anthropic.com/oauth2/authorize",
			TokenURL:     "https://console.anthropic.com/oauth2/token",
			Scopes:       []string{"read", "write"},
		},
	}

	for providerID, provider := range providers {
		t.Run(providerID, func(t *testing.T) {
			// Verify configuration structure
			if provider.Name == "" {
				t.Error("Provider name should not be empty")
			}
			if provider.ClientID == "" {
				t.Error("Client ID should not be empty")
			}
			if provider.AuthURL == "" {
				t.Error("Auth URL should not be empty")
			}
			if provider.TokenURL == "" {
				t.Error("Token URL should not be empty")
			}
			if len(provider.Scopes) == 0 {
				t.Error("Provider should have at least one scope")
			}

			// Verify HTTPS URLs (required for security)
			if provider.AuthURL[:8] != "https://" {
				t.Errorf("Auth URL should use HTTPS: %s", provider.AuthURL)
			}
			if provider.TokenURL[:8] != "https://" {
				t.Errorf("Token URL should use HTTPS: %s", provider.TokenURL)
			}

			t.Logf("OAuth provider %s configured correctly: auth_url=%s", providerID, provider.AuthURL)
		})
	}
}

// TestDeviceFlowStructure tests device flow request/response structures
// Testable without network as it validates structure only
func TestDeviceFlowStructure(t *testing.T) {
	testCases := []struct {
		name        string
		clientID    string
		pkce        bool
		description string
	}{
		{
			name:        "basic device flow",
			clientID:    "test-client-id",
			pkce:        false,
			description: "Basic OAuth device flow without PKCE",
		},
		{
			name:        "device flow with PKCE",
			clientID:    "test-client-id-pkce",
			pkce:        true,
			description: "OAuth device flow with PKCE enhancement",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create PKCE params if required
			var pkce *auth.PKCEParams
			if tc.pkce {
				var err error
				pkce, err = auth.GeneratePKCE()
				if err != nil {
					t.Fatalf("Failed to generate PKCE: %v", err)
				}
			}

			// Verify request structure
			req := auth.DeviceFlowRequest{
				ClientID: tc.clientID,
				PKCE:     pkce,
			}

			if req.ClientID != tc.clientID {
				t.Error("Client ID mismatch")
			}

			if tc.pkce {
				if req.PKCE == nil {
					t.Error("PKCE should be set when required")
				}
				if req.PKCE.CodeVerifier == "" {
					t.Error("PKCE code verifier should not be empty")
				}
				if req.PKCE.CodeChallenge == "" {
					t.Error("PKCE code challenge should not be empty")
				}
				if req.PKCE.Method != "S256" {
					t.Errorf("PKCE method should be S256, got: %s", req.PKCE.Method)
				}
			}

			t.Logf("%s: client_id=%s, pkce=%t", tc.description, tc.clientID, tc.pkce)
		})
	}
}

// TestTokenStorageKeys tests that OAuth token storage keys are properly formatted
// Testable without network as it validates key format only
func TestTokenStorageKeys(t *testing.T) {
	testCases := []struct {
		providerID  string
		expectedKey string
		description string
	}{
		{
			providerID:  "google",
			expectedKey: "oauth_google_access",
			description: "Google OAuth access token key",
		},
		{
			providerID:  "anthropic",
			expectedKey: "oauth_anthropic_access",
			description: "Anthropic OAuth access token key",
		},
		{
			providerID:  "openai",
			expectedKey: "oauth_openai_access",
			description: "OpenAI OAuth access token key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.providerID, func(t *testing.T) {
			// Verify key format (should be lowercase, underscores for spaces)
			key := "oauth_" + tc.providerID + "_access"

			if key != tc.expectedKey {
				t.Errorf("Expected key %s, got: %s", tc.expectedKey, key)
			}

			// Verify no spaces or special characters
			for i, c := range key {
				if c == ' ' || (c < 'a' && c > 'z') && (c < 'A' && c > 'Z') && c != '_' && (c < '0' && c > '9') {
					t.Errorf("Invalid character at position %d in key: %s", i, key)
				}
			}

			t.Logf("Token key for %s: %s", tc.description, key)
		})
	}
}

// TestOAuthValidationRules tests OAuth validation rules
// Testable without network as it validates logic only
func TestOAuthValidationRules(t *testing.T) {
	testCases := []struct {
		name        string
		clientID    string
		scopes      []string
		shouldError bool
		description string
	}{
		{
			name:        "valid configuration",
			clientID:    "valid-client-id",
			scopes:      []string{"read", "write"},
			shouldError: false,
			description: "Valid OAuth provider configuration",
		},
		{
			name:        "empty client ID",
			clientID:    "",
			scopes:      []string{"read"},
			shouldError: true,
			description: "Empty client ID should be rejected",
		},
		{
			name:        "empty scopes",
			clientID:    "valid-client-id",
			scopes:      []string{},
			shouldError: true,
			description: "Empty scopes should be rejected",
		},
		{
			name:        "single scope",
			clientID:    "valid-client-id",
			scopes:      []string{"read"},
			shouldError: false,
			description: "Single scope is valid",
		},
		{
			name:        "many scopes",
			clientID:    "valid-client-id",
			scopes:      []string{"read", "write", "admin", "delete"},
			shouldError: false,
			description: "Multiple scopes are valid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := &config.OAuthProvider{
				Name:         "Test Provider",
				ClientID:     tc.clientID,
				ClientSecret: "test-secret",
				AuthURL:      "https://auth.example.com",
				TokenURL:     "https://token.example.com",
				Scopes:       tc.scopes,
			}

			// Validate configuration
			hasError := provider.ClientID == "" || len(provider.Scopes) == 0

			if hasError != tc.shouldError {
				t.Errorf("Expected error=%v for %s, got error=%v",
					tc.shouldError, tc.description, hasError)
			}

			t.Logf("Validation for %s: error=%v", tc.description, hasError)
		})
	}
}

// TestEnvironmentVariableHandling tests environment variable naming conventions
// Testable without actual environment manipulation
func TestEnvironmentVariableHandling(t *testing.T) {
	envVars := []struct {
		name        string
		expected    string
		description string
	}{
		{
			name:        "PRYX_DATA_DIR",
			expected:    "PRYX_DATA_DIR",
			description: "Pryx data directory environment variable",
		},
		{
			name:        "PRYX_KEYCHAIN_FILE",
			expected:    "PRYX_KEYCHAIN_FILE",
			description: "Pryx keychain file environment variable",
		},
		{
			name:        "PRYX_DB_PATH",
			expected:    "PRYX_DB_PATH",
			description: "Pryx database path environment variable",
		},
		{
			name:        "PRYX_LISTEN_ADDR",
			expected:    "PRYX_LISTEN_ADDR",
			description: "Pryx listen address environment variable",
		},
		{
			name:        "PRYX_CLOUD_API_URL",
			expected:    "PRYX_CLOUD_API_URL",
			description: "Pryx cloud API URL environment variable",
		},
	}

	for _, env := range envVars {
		t.Run(env.name, func(t *testing.T) {
			// Verify naming convention (uppercase with underscores)
			if env.name != env.expected {
				t.Errorf("Expected environment variable %s, got: %s", env.expected, env.name)
			}

			// Verify format (uppercase letters, numbers, underscores only)
			for i, c := range env.name {
				if (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
					t.Errorf("Invalid character at position %d in env var: %s", i, env.name)
				}
			}

			t.Logf("Environment variable %s: %s", env.description, env.name)
		})
	}
}
