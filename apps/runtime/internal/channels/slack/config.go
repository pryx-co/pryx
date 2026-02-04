package slack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultConfigDir  = ".pryx/config"
	defaultConfigFile = "slack.json"
)

type Config struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	BotToken        string    `json:"bot_token"`
	AppToken        string    `json:"app_token"`
	AllowedChannels []string  `json:"allowed_channels"`
	AllowedDMs      bool      `json:"allowed_dms"`
	SignatureKey    string    `json:"signature_key,omitempty"`
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (c *Config) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("config ID is required")
	}

	if c.Name == "" {
		return fmt.Errorf("config name is required")
	}

	if c.BotToken == "" {
		return fmt.Errorf("bot token is required")
	}

	return nil
}

func (c *Config) IsChannelAllowed(channelID string) bool {
	if len(c.AllowedChannels) == 0 {
		return true
	}

	for _, id := range c.AllowedChannels {
		if id == channelID {
			return true
		}
	}

	return false
}

func (c *Config) SetDefaults() {
	if len(c.AllowedChannels) == 0 {
		c.AllowedChannels = []string{}
	}
}

type ConfigManager struct {
	configPath string
}

func NewSlackConfigManager() *ConfigManager {
	home, _ := os.UserHomeDir()
	return &ConfigManager{
		configPath: filepath.Join(home, defaultConfigDir, defaultConfigFile),
	}
}

func (cm *ConfigManager) LoadAll() ([]Config, error) {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var configs []Config
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	for i := range configs {
		configs[i].SetDefaults()
	}

	return configs, nil
}

func (cm *ConfigManager) SaveAll(configs []Config) error {
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configsToSave := make([]Config, len(configs))
	for i, config := range configs {
		configsToSave[i] = config
		configsToSave[i].BotToken = ""
		configsToSave[i].AppToken = ""
	}

	data, err := json.MarshalIndent(configsToSave, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (cm *ConfigManager) Get(id string) (*Config, error) {
	configs, err := cm.LoadAll()
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.ID == id {
			return &config, nil
		}
	}

	return nil, fmt.Errorf("slack config not found: %s", id)
}

func (cm *ConfigManager) Save(config Config) error {
	configs, err := cm.LoadAll()
	if err != nil {
		return err
	}

	now := time.Now()
	config.UpdatedAt = now

	found := false
	for i, c := range configs {
		if c.ID == config.ID {
			config.CreatedAt = c.CreatedAt
			configs[i] = config
			found = true
			break
		}
	}

	if !found {
		config.CreatedAt = now
		configs = append(configs, config)
	}

	return cm.SaveAll(configs)
}

func (cm *ConfigManager) Delete(id string) error {
	configs, err := cm.LoadAll()
	if err != nil {
		return err
	}

	filtered := make([]Config, 0, len(configs))
	found := false
	for _, config := range configs {
		if config.ID != id {
			filtered = append(filtered, config)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("slack config not found: %s", id)
	}

	return cm.SaveAll(filtered)
}

func (cm *ConfigManager) Create(config Config) (*Config, error) {
	if config.ID == "" {
		config.ID = generateID()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	config.SetDefaults()

	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	if err := cm.Save(config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (cm *ConfigManager) Update(id string, updates map[string]interface{}) (*Config, error) {
	config, err := cm.Get(id)
	if err != nil {
		return nil, err
	}

	if name, ok := updates["name"].(string); ok {
		config.Name = name
	}
	if botToken, ok := updates["bot_token"].(string); ok {
		config.BotToken = botToken
	}
	if appToken, ok := updates["app_token"].(string); ok {
		config.AppToken = appToken
	}
	if allowedChannels, ok := updates["allowed_channels"].([]string); ok {
		config.AllowedChannels = allowedChannels
	}
	if allowedDMs, ok := updates["allowed_dms"].(bool); ok {
		config.AllowedDMs = allowedDMs
	}
	if signatureKey, ok := updates["signature_key"].(string); ok {
		config.SignatureKey = signatureKey
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		config.Enabled = enabled
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	config.UpdatedAt = time.Now()

	if err := cm.Save(*config); err != nil {
		return nil, err
	}

	return config, nil
}

func (cm *ConfigManager) List() ([]Config, error) {
	return cm.LoadAll()
}

func (cm *ConfigManager) ListEnabled() ([]Config, error) {
	configs, err := cm.LoadAll()
	if err != nil {
		return nil, err
	}

	enabled := make([]Config, 0)
	for _, config := range configs {
		if config.Enabled {
			enabled = append(enabled, config)
		}
	}

	return enabled, nil
}

func generateID() string {
	return fmt.Sprintf("slack-%d", time.Now().UnixNano())
}

func DefaultConfig() Config {
	return Config{
		AllowedChannels: []string{},
		AllowedDMs:      false,
		Enabled:         true,
	}
}

func NewBotConfig(name, botToken, appToken string) Config {
	config := DefaultConfig()
	config.Name = name
	config.BotToken = botToken
	config.AppToken = appToken
	return config
}
