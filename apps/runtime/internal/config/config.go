// Package config provides configuration management for the Pryx runtime.
// Configuration can be loaded from YAML files and overridden via environment variables.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration settings for the Pryx runtime.
type Config struct {
	// ListenAddr is the address to listen on (e.g., ":3000" or ":0" for dynamic port).
	ListenAddr string `yaml:"listen_addr"`
	// DatabasePath is the path to the SQLite database file.
	DatabasePath string `yaml:"database_path"`
	// CloudAPIUrl is the URL of the Pryx Cloud API.
	CloudAPIUrl string `yaml:"cloud_api_url"`

	// AI Configuration
	// ModelProvider is the LLM provider to use (openai, anthropic, ollama, glm).
	ModelProvider string `yaml:"model_provider"`
	// ModelName is the specific model to use (e.g., gpt-4, claude-3-opus, llama3).
	ModelName string `yaml:"model_name"`
	// OllamaEndpoint is the URL of the Ollama server (when using Ollama provider).
	OllamaEndpoint string `yaml:"ollama_endpoint"`

	// Channels
	// TelegramToken is the bot token for Telegram integration.
	// TelegramEnabled enables or disables the Telegram bot.
	TelegramToken   string `yaml:"telegram_token"`
	TelegramEnabled bool   `yaml:"telegram_enabled"`
	// SlackAppToken and SlackBotToken are the tokens for Slack integration.
	// SlackEnabled enables or disables the Slack bot.
	SlackAppToken string `yaml:"slack_app_token"`
	SlackBotToken string `yaml:"slack_bot_token"`
	SlackEnabled  bool   `yaml:"slack_enabled"`

	// Memory Management
	// MaxMessagesPerSession limits the number of messages kept per session (0 = unlimited).
	MaxMessagesPerSession int `yaml:"max_messages_per_session"`
	// WebSocketBufferSize sets the WebSocket message buffer size.
	WebSocketBufferSize int `yaml:"websocket_buffer_size"`
	// EnableMemoryProfiling enables memory usage monitoring.
	EnableMemoryProfiling bool `yaml:"enable_memory_profiling"`
}

// ProviderKeyNames maps provider IDs to their keychain key names.
var ProviderKeyNames = map[string]string{
	"openai":     "provider:openai",
	"anthropic":  "provider:anthropic",
	"openrouter": "provider:openrouter",
	"together":   "provider:together",
	"groq":       "provider:groq",
	"xai":        "provider:xai",
	"mistral":    "provider:mistral",
	"cohere":     "provider:cohere",
	"google":     "provider:google",
	"glm":        "provider:glm",
	"slack":      "provider:slack",
}

// DefaultPath returns the default configuration file path.
// The config is stored in ~/.pryx/config.yaml.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pryx", "config.yaml")
}

// Load loads configuration from the default file and environment variables.
// Environment variables take precedence over file configuration.
// Returns a Config with default values if no configuration file exists.
func Load() *Config {
	cfg := &Config{
		ListenAddr:      ":0", // Use :0 for dynamic port allocation
		DatabasePath:    "pryx.db",
		CloudAPIUrl:     "https://pryx.dev/api",
		ModelProvider:   "ollama",
		ModelName:       "llama3",
		OllamaEndpoint:  "http://localhost:11434",
		TelegramEnabled: false,
		SlackEnabled:    false,
		SlackAppToken:   "",
		SlackBotToken:   "",
	}

	// Try loading from default file
	path := DefaultPath()
	if _, err := os.Stat(path); err == nil {
		if fileCfg, err := LoadFromFile(path); err == nil {
			*cfg = *fileCfg
		}
	}

	// Environment variables override file configuration
	if v := os.Getenv("PRYX_LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("PRYX_DB_PATH"); v != "" {
		cfg.DatabasePath = v
	}
	if v := os.Getenv("PRYX_CLOUD_API_URL"); v != "" {
		cfg.CloudAPIUrl = v
	}
	if v := os.Getenv("PRYX_SLACK_APP_TOKEN"); v != "" {
		cfg.SlackAppToken = v
	}
	if v := os.Getenv("PRYX_SLACK_BOT_TOKEN"); v != "" {
		cfg.SlackBotToken = v
	}
	if v := os.Getenv("PRYX_SLACK_ENABLED"); v != "" {
		cfg.SlackEnabled = true
	}

	return cfg
}

// LoadFromFile loads configuration from a specific YAML file path.
// Returns an error if the file cannot be read or parsed.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the configuration to a YAML file at the specified path.
// Creates parent directories if they don't exist.
// Returns an error if the file cannot be written.
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// getEnv returns the value of an environment variable or a fallback value.
// Returns fallback if the environment variable is not set.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
