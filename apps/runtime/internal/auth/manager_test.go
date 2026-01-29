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
