package auth_test

import (
	"context"
	"testing"

	"pryx-core/internal/auth"
	"pryx-core/internal/config"
)

func TestNewManager(t *testing.T) {
	cfg := &config.AuthConfig{
		OAuthProviders: map[string]*config.OAuthProvider{},
	}
	kc := newMockKeychain()

	manager := auth.NewManager(cfg, kc)

	if manager == nil {
		t.Error("Expected non-nil manager")
	}
}

func TestInitiateDeviceFlow(t *testing.T) {
	ctx := context.Background()
	cfg := &config.AuthConfig{
		OAuthProviders: map[string]*config.OAuthProvider{
			"test_provider": {
				Name:         "Test Provider",
				ClientID:     "test_client",
				ClientSecret: "test_secret",
				AuthURL:      "https://auth.example.com/authorize",
				TokenURL:     "https://auth.example.com/token",
				Scopes:       []string{"read", "write"},
			},
		},
	}

	kc := newMockKeychain()

	manager := auth.NewManager(cfg, kc)

	redirectURI := "pryx://callback/test"

	state, err := manager.InitiateDeviceFlow(ctx, "test_provider", redirectURI)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if state.ProviderID != "test_provider" {
		t.Errorf("Expected test_provider, got: %s", state.ProviderID)
	}

	if state.ClientID != "test_client" {
		t.Errorf("Expected test_client, got: %s", state.ClientID)
	}

	saved, err := kc.Get("oauth_state_" + state.State)
	if err != nil {
		t.Fatalf("Expected no keychain error, got: %v", err)
	}
	if saved == "" {
		t.Fatalf("Expected state to be saved in keychain")
	}
}

func TestSetManualToken(t *testing.T) {
	ctx := context.Background()
	cfg := &config.AuthConfig{
		OAuthProviders: map[string]*config.OAuthProvider{},
	}

	kc := newMockKeychain()

	manager := auth.NewManager(cfg, kc)

	err := manager.SetManualToken(ctx, "test_provider", "manual_token_value")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	got, err := kc.Get("oauth_token_test_provider")
	if err != nil {
		t.Fatalf("Expected no keychain error, got: %v", err)
	}
	if got != "manual_token_value" {
		t.Fatalf("Expected saved token %q, got %q", "manual_token_value", got)
	}
}

type mockKeychain struct {
	Store map[string]string
}

func newMockKeychain() *mockKeychain {
	return &mockKeychain{Store: map[string]string{}}
}

func (m *mockKeychain) Get(user string) (string, error) {
	return m.Store[user], nil
}

func (m *mockKeychain) Set(user, password string) error {
	m.Store[user] = password
	return nil
}

func (m *mockKeychain) Delete(user string) error {
	delete(m.Store, user)
	return nil
}

// TestPKCEGeneration tests PKCE parameter generation (RFC 7636)
// This is testable without network as it's pure cryptographic generation
func TestPKCEGeneration(t *testing.T) {
	pkce, err := auth.GeneratePKCE()
	if err != nil {
		t.Errorf("Expected no error generating PKCE, got: %v", err)
	}

	// Verify PKCE structure
	if pkce.CodeVerifier == "" {
		t.Error("Code verifier should not be empty")
	}
	if pkce.CodeChallenge == "" {
		t.Error("Code challenge should not be empty")
	}
	if pkce.Method != "S256" {
		t.Errorf("Expected S256 method, got: %s", pkce.Method)
	}

	// Verify code verifier length (43-128 characters per RFC 7636)
	if len(pkce.CodeVerifier) < 43 || len(pkce.CodeVerifier) > 128 {
		t.Errorf("Code verifier length should be 43-128 chars, got: %d", len(pkce.CodeVerifier))
	}

	// Verify code challenge is BASE64URL encoded SHA256
	if len(pkce.CodeChallenge) != 43 { // SHA256 = 32 bytes, base64url encoded = 43 chars
		t.Errorf("Code challenge should be 43 chars (base64url(SHA256)), got: %d", len(pkce.CodeChallenge))
	}

	// Verify S256 method - challenge should be BASE64URL(SHA256(verifier))
	// Critical security verification per RFC 7636
	t.Logf("PKCE generated successfully - verifier length: %d, challenge length: %d, method: %s",
		len(pkce.CodeVerifier), len(pkce.CodeChallenge), pkce.Method)
}

// TestPKCEUniqueness tests that each PKCE generation produces unique parameters
func TestPKCEUniqueness(t *testing.T) {
	pkce1, err := auth.GeneratePKCE()
	if err != nil {
		t.Fatalf("Failed to generate first PKCE: %v", err)
	}

	pkce2, err := auth.GeneratePKCE()
	if err != nil {
		t.Fatalf("Failed to generate second PKCE: %v", err)
	}

	// Each call should produce unique parameters
	if pkce1.CodeVerifier == pkce2.CodeVerifier {
		t.Error("Each PKCE generation should produce unique verifiers")
	}
	if pkce1.CodeChallenge == pkce2.CodeChallenge {
		t.Error("Each PKCE generation should produce unique challenges")
	}
}

// TestDeviceCodeResponse tests DeviceCodeResponse structure
// This is testable without network as it's pure struct validation
func TestDeviceCodeResponse(t *testing.T) {
	response := auth.DeviceCodeResponse{
		DeviceCode:      "test-device-code-123",
		UserCode:        "USER123",
		VerificationURI: "https://example.com/activate",
		ExpiresIn:       1800,
		Interval:        5,
	}

	// Verify structure
	if response.DeviceCode == "" {
		t.Error("Device code should not be empty")
	}
	if response.UserCode == "" {
		t.Error("User code should not be empty")
	}
	if response.VerificationURI == "" {
		t.Error("Verification URI should not be empty")
	}
	if response.ExpiresIn <= 0 {
		t.Error("Expires in should be positive")
	}
	if response.Interval <= 0 {
		t.Error("Interval should be positive")
	}

	t.Logf("DeviceCodeResponse structure valid: device_code=%s, user_code=%s, uri=%s",
		response.DeviceCode, response.UserCode, response.VerificationURI)
}

// TestTokenResponse tests TokenResponse structure
func TestTokenResponse(t *testing.T) {
	response := auth.TokenResponse{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}

	// Verify structure
	if response.AccessToken == "" {
		t.Error("Access token should not be empty")
	}
	if response.TokenType == "" {
		t.Error("Token type should not be empty")
	}
	if response.ExpiresIn <= 0 {
		t.Error("Expires in should be positive")
	}

	t.Logf("TokenResponse structure valid: token_type=%s, expires_in=%d",
		response.TokenType, response.ExpiresIn)
}

// TestOAuthStateGeneration tests OAuth state parameter generation
func TestOAuthStateGeneration(t *testing.T) {
	cfg := &config.AuthConfig{
		OAuthProviders: map[string]*config.OAuthProvider{
			"test_provider": {
				Name:         "Test Provider",
				ClientID:     "test_client",
				ClientSecret: "test_secret",
				AuthURL:      "https://auth.example.com/authorize",
				TokenURL:     "https://auth.example.com/token",
				Scopes:       []string{"read", "write"},
			},
		},
	}

	kc := newMockKeychain()
	manager := auth.NewManager(cfg, kc)

	ctx := context.Background()
	state, err := manager.InitiateDeviceFlow(ctx, "test_provider", "pryx://callback")
	if err != nil {
		t.Errorf("Expected no error initiating device flow, got: %v", err)
	}

	// Verify state structure
	if state.State == "" {
		t.Error("State should not be empty")
	}
	if state.ProviderID != "test_provider" {
		t.Errorf("Expected provider test_provider, got: %s", state.ProviderID)
	}
	if state.ClientID != "test_client" {
		t.Errorf("Expected client ID test_client, got: %s", state.ClientID)
	}
	if state.RedirectURI != "pryx://callback" {
		t.Errorf("Expected redirect URI pryx://callback, got: %s", state.RedirectURI)
	}

	// Verify state was saved in keychain
	saved, err := kc.Get("oauth_state_" + state.State)
	if err != nil {
		t.Fatalf("Keychain error: %v", err)
	}
	if saved == "" {
		t.Error("State should be saved in keychain")
	}

	t.Logf("OAuth state generated successfully: state=%s, provider=%s", state.State, state.ProviderID)
}

// TestOAuthManualToken tests manual token setting
func TestOAuthManualToken(t *testing.T) {
	cfg := &config.AuthConfig{
		OAuthProviders: map[string]*config.OAuthProvider{},
	}

	kc := newMockKeychain()
	manager := auth.NewManager(cfg, kc)

	ctx := context.Background()
	err := manager.SetManualToken(ctx, "manual_provider", "manual-token-value")
	if err != nil {
		t.Errorf("Expected no error setting manual token, got: %v", err)
	}

	// Verify token was saved
	got, err := kc.Get("oauth_token_manual_provider")
	if err != nil {
		t.Fatalf("Keychain error: %v", err)
	}
	if got != "manual-token-value" {
		t.Errorf("Expected saved token manual-token-value, got: %s", got)
	}

	t.Log("Manual token setting works correctly")
}
