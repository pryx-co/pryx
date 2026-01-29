package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"pryx-core/internal/config"
)

const (
	tokenValidity = time.Hour
)

type OAuthState struct {
	State       string    `json:"state"`
	ProviderID  string    `json:"provider_id"`
	ClientID    string    `json:"client_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	RedirectURI string    `json:"redirect_uri,omitempty"`
}

type Manager struct {
	cfg      *config.AuthConfig
	keychain Keychain
}

type Keychain interface {
	Set(user, password string) error
	Get(user string) (string, error)
	Delete(user string) error
}

func NewManager(cfg *config.AuthConfig, keychain Keychain) *Manager {
	if cfg == nil {
		cfg = &config.AuthConfig{OAuthProviders: map[string]*config.OAuthProvider{}}
	}
	if cfg.OAuthProviders == nil {
		cfg.OAuthProviders = map[string]*config.OAuthProvider{}
	}
	return &Manager{
		cfg:      cfg,
		keychain: keychain,
	}
}

func (m *Manager) InitiateDeviceFlow(ctx context.Context, providerID string, redirectURI string) (*OAuthState, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.keychain == nil {
		return nil, errors.New("keychain not available")
	}

	provider, ok := m.cfg.OAuthProviders[providerID]
	if !ok {
		return nil, errors.New("provider not found")
	}

	state, err := m.createOAuthState(providerID, provider.ClientID, redirectURI)
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	if err := m.keychain.Set("oauth_state_"+state.State, string(encoded)); err != nil {
		return nil, err
	}

	return state, nil
}

func (m *Manager) createOAuthState(providerID string, clientID string, redirectURI string) (*OAuthState, error) {
	state, err := generateRandomState()
	if err != nil {
		return nil, err
	}

	return &OAuthState{
		State:       state,
		ProviderID:  providerID,
		ClientID:    clientID,
		ExpiresAt:   time.Now().UTC().Add(tokenValidity),
		RedirectURI: redirectURI,
	}, nil
}

func (m *Manager) SetManualToken(ctx context.Context, providerID string, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.keychain == nil {
		return errors.New("keychain not available")
	}
	if token == "" {
		return errors.New("token cannot be empty")
	}

	return m.keychain.Set("oauth_token_"+providerID, token)
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
