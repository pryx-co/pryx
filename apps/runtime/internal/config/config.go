package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr   string `yaml:"listen_addr"`
	DatabasePath string `yaml:"database_path"`
	CloudAPIUrl  string `yaml:"cloud_api_url"`

	// AI Configuration
	ModelProvider  string `yaml:"model_provider"` // openai, anthropic, ollama, glm
	ModelName      string `yaml:"model_name"`     // e.g. gpt-4, claude-3-opus, llama3, glm-4-flash
	OllamaEndpoint string `yaml:"ollama_endpoint"`

	// Channels
	TelegramToken   string `yaml:"telegram_token"`
	TelegramEnabled bool   `yaml:"telegram_enabled"`
}

// ProviderKeyNames maps provider IDs to their keychain key names
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
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pryx", "config.yaml")
}

func Load() *Config {
	cfg := &Config{
		ListenAddr:     ":0", // Use :0 for dynamic port allocation (like OpenCode/moltbot)
		DatabasePath:   "pryx.db",
		CloudAPIUrl:    "https://pryx.dev/api",
		ModelProvider:  "ollama",
		ModelName:      "llama3",
		OllamaEndpoint: "http://localhost:11434",
	}

	// Try loading from default file
	path := DefaultPath()
	if _, err := os.Stat(path); err == nil {
		if fileCfg, err := LoadFromFile(path); err == nil {
			*cfg = *fileCfg
		}
	}

	// Env overrides
	if v := os.Getenv("PRYX_LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("PRYX_DB_PATH"); v != "" {
		cfg.DatabasePath = v
	}
	if v := os.Getenv("PRYX_CLOUD_API_URL"); v != "" {
		cfg.CloudAPIUrl = v
	}

	return cfg
}

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

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
