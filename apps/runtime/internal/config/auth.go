package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type OAuthProvider struct {
	Name         string   `json:"name"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	Scopes       []string `json:"scopes"`
}

type AuthConfig struct {
	OAuthProviders map[string]*OAuthProvider `json:"oauth_providers"`
}

func DefaultAuthConfig() *AuthConfig {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return &AuthConfig{
			OAuthProviders: map[string]*OAuthProvider{},
		}
	}

	configDir := filepath.Join(homeDir, ".pryx")
	configFile := filepath.Join(configDir, "auth.json")

	// Load existing config if exists
	var config AuthConfig
	if data, err := os.ReadFile(configFile); err == nil {
		json.Unmarshal(data, &config)
	}

	return &config
}

func GetAuthConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".pryx", "auth.json"), nil
}

func SetAuthConfig(config *AuthConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".pryx")
	configFile := filepath.Join(configDir, "auth.json")

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

func AddOAuthProvider(id string, provider *OAuthProvider) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".pryx")
	configFile := filepath.Join(configDir, "auth.json")

	// Load existing config
	var config AuthConfig
	if data, err := os.ReadFile(configFile); err == nil {
		json.Unmarshal(data, &config)
	}

	if config.OAuthProviders == nil {
		config.OAuthProviders = map[string]*OAuthProvider{}
	}

	config.OAuthProviders[id] = provider

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}
