// Package config provides configuration management for the Pryx runtime.
// Configuration can be loaded from YAML files and overridden via environment variables.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration settings for the Pryx runtime.
type Config struct {
	// ListenAddr is the address to listen on (e.g., ":3000" or ":0" for dynamic port).
	ListenAddr string `yaml:"listen_addr"`
	// DatabasePath is the path to the SQLite database file.
	DatabasePath string `yaml:"database_path"`
	// SkillsPath is the directory where skills are installed.
	SkillsPath string `yaml:"skills_path"`
	// CachePath is the directory for cached data.
	CachePath string `yaml:"cache_path"`
	// CloudAPIUrl is the URL of the Pryx Cloud API.
	CloudAPIUrl string `yaml:"cloud_api_url"`

	// Agent Detection
	// AgentDetectEnabled enables automatic detection of external agents.
	AgentDetectEnabled bool `yaml:"agent_detect_enabled"`
	// AgentDetectInterval is how often to scan for agents.
	AgentDetectInterval time.Duration `yaml:"agent_detect_interval"`

	// AI Configuration
	// ModelProvider is the LLM provider to use (openai, anthropic, ollama, glm).
	ModelProvider string `yaml:"model_provider"`
	// ModelName is the specific model to use (e.g., gpt-4, claude-3-opus, llama3).
	ModelName string `yaml:"model_name"`
	// OllamaEndpoint is the URL of the Ollama server (when using Ollama provider).
	OllamaEndpoint string `yaml:"ollama_endpoint"`
	// ConfiguredProviders is the list of providers that have been explicitly configured.
	// This tracks providers added via 'provider add' even without API keys (e.g., Ollama).
	ConfiguredProviders []string `yaml:"configured_providers"`

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

	// RAG Memory System
	// MemoryEnabled enables the RAG memory system.
	MemoryEnabled bool `yaml:"memory_enabled"`
	// MemoryAutoFlush enables automatic memory flushing before context compaction.
	MemoryAutoFlush bool `yaml:"memory_auto_flush"`
	// MemoryFlushThresholdTokens triggers auto-flush when token count approaches this threshold.
	MemoryFlushThresholdTokens int `yaml:"memory_flush_threshold_tokens"`

	// Security Configuration
	// AllowedOrigins is a list of allowed CORS origins. Use specific origins in production.
	// Defaults include localhost for development.
	AllowedOrigins []string `yaml:"allowed_origins"`
	// WebSocketAllowedOrigins is a list of allowed WebSocket origins.
	// If empty, defaults to AllowedOrigins.
	WebSocketAllowedOrigins []string `yaml:"websocket_allowed_origins"`
	// MaxWebSocketConnections limits concurrent WebSocket connections (0 = unlimited).
	MaxWebSocketConnections int `yaml:"max_websocket_connections"`
	// MaxWebSocketMessageSize sets the maximum message size in bytes (default: 10MB).
	MaxWebSocketMessageSize int64 `yaml:"max_websocket_message_size"`
	// WebSocketRateLimitPerMinute sets max connections per minute per IP (default: 60).
	WebSocketRateLimitPerMinute int `yaml:"websocket_rate_limit_per_minute"`
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
	return filepath.Join(defaultPryxDir(), "config.yaml")
}

func defaultPryxDir() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ".pryx"
	}
	return filepath.Join(home, ".pryx")
}

// Load loads configuration from the default file and environment variables.
// Environment variables take precedence over file configuration.
// Returns a Config with default values if no configuration file exists.
func Load() *Config {
	pryxDir := defaultPryxDir()

	cfg := &Config{
		ListenAddr:                  ":0", // Use :0 for dynamic port allocation
		DatabasePath:                filepath.Join(pryxDir, "pryx.db"),
		SkillsPath:                  filepath.Join(pryxDir, "skills"),
		CachePath:                   filepath.Join(pryxDir, "cache"),
		CloudAPIUrl:                 "https://pryx.dev/api",
		ModelProvider:               "ollama",
		ModelName:                   "llama3",
		OllamaEndpoint:              "http://localhost:11434",
		TelegramEnabled:             false,
		SlackEnabled:                false,
		SlackAppToken:               "",
		SlackBotToken:               "",
		AgentDetectEnabled:          false,
		AgentDetectInterval:         30 * time.Second,
		MemoryEnabled:               true,
		MemoryAutoFlush:             true,
		MemoryFlushThresholdTokens:  100000,
		AllowedOrigins:              []string{}, // Defaults to localhost via middleware logic
		MaxWebSocketConnections:     1000,
		MaxWebSocketMessageSize:     10 * 1024 * 1024, // 10MB
		WebSocketRateLimitPerMinute: 60,
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

	_ = os.MkdirAll(pryxDir, 0o755)
	if strings.TrimSpace(cfg.SkillsPath) != "" {
		_ = os.MkdirAll(cfg.SkillsPath, 0o755)
	}
	if strings.TrimSpace(cfg.CachePath) != "" {
		_ = os.MkdirAll(cfg.CachePath, 0o755)
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
