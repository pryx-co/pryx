package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path := DefaultPath()
	assert.Contains(t, path, ".pryx")
	assert.Contains(t, path, "config.yaml")
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Clear environment
	os.Unsetenv("PRYX_LISTEN_ADDR")
	os.Unsetenv("PRYX_DB_PATH")
	os.Unsetenv("PRYX_CLOUD_API_URL")

	cfg := Load()

	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.ListenAddr)
	assert.NotEmpty(t, cfg.DatabasePath)
}

func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Set environment variables
	os.Setenv("PRYX_LISTEN_ADDR", "127.0.0.1:8080")
	os.Setenv("PRYX_DB_PATH", "/custom/path.db")
	os.Setenv("PRYX_CLOUD_API_URL", "https://custom.api.com")

	defer func() {
		os.Unsetenv("PRYX_LISTEN_ADDR")
		os.Unsetenv("PRYX_DB_PATH")
		os.Unsetenv("PRYX_CLOUD_API_URL")
	}()

	cfg := Load()

	assert.Equal(t, "127.0.0.1:8080", cfg.ListenAddr)
	assert.Equal(t, "/custom/path.db", cfg.DatabasePath)
	assert.Equal(t, "https://custom.api.com", cfg.CloudAPIUrl)
}

func TestLoadFromEnvironment_Partial(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Set only one variable
	os.Setenv("PRYX_LISTEN_ADDR", ":9000")
	defer os.Unsetenv("PRYX_LISTEN_ADDR")

	// Clear others
	os.Unsetenv("PRYX_DB_PATH")
	os.Unsetenv("PRYX_CLOUD_API_URL")

	cfg := Load()

	assert.Equal(t, ":9000", cfg.ListenAddr)
	assert.Equal(t, filepath.Join(home, ".pryx", "pryx.db"), cfg.DatabasePath)
	assert.Equal(t, "https://pryx.dev/api", cfg.CloudAPIUrl) // Default
}

func TestLoadFromFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
listen_addr: ":4000"
database_path: "/data/pryx.db"
cloud_api_url: "https://test.api.com"
model_provider: "openai"
model_name: "gpt-4"
openai_key: "test-key"
telegram_enabled: true
telegram_token: "test-token"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := LoadFromFile(configPath)
	require.NoError(t, err)

	assert.Equal(t, ":4000", cfg.ListenAddr)
	assert.Equal(t, "/data/pryx.db", cfg.DatabasePath)
	assert.Equal(t, "https://test.api.com", cfg.CloudAPIUrl)
	assert.Equal(t, "openai", cfg.ModelProvider)
	assert.Equal(t, "gpt-4", cfg.ModelName)
	// API keys are stored in keychain, not config file
	assert.True(t, cfg.TelegramEnabled)
	assert.Equal(t, "test-token", cfg.TelegramToken)
}

func TestLoadFromFile_NotFound(t *testing.T) {
	cfg, err := LoadFromFile("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte("invalid: yaml: content: [}"), 0644)
	require.NoError(t, err)

	cfg, err := LoadFromFile(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	cfg := &Config{
		ListenAddr:    ":5000",
		DatabasePath:  "/test/db",
		CloudAPIUrl:   "https://save.test.com",
		ModelProvider: "anthropic",
		ModelName:     "claude-3",

		TelegramEnabled: true,
		TelegramToken:   "token123",
	}

	err := cfg.Save(configPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load and verify
	loaded, err := LoadFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, cfg.ListenAddr, loaded.ListenAddr)
	assert.Equal(t, cfg.DatabasePath, loaded.DatabasePath)
	assert.Equal(t, cfg.ModelProvider, loaded.ModelProvider)
	assert.Equal(t, cfg.TelegramEnabled, loaded.TelegramEnabled)
}

func TestConfig_Save_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create initial file
	cfg1 := &Config{ListenAddr: ":1000"}
	err := cfg1.Save(configPath)
	require.NoError(t, err)

	// Overwrite with new config
	cfg2 := &Config{ListenAddr: ":2000"}
	err = cfg2.Save(configPath)
	require.NoError(t, err)

	// Verify new content
	loaded, err := LoadFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, ":2000", loaded.ListenAddr)
}

func TestConfig_Save_CreateDir(t *testing.T) {
	tmpDir := t.TempDir()
	// Path with multiple nested directories that don't exist
	configPath := filepath.Join(tmpDir, "a", "b", "c", "config.yaml")

	cfg := &Config{ListenAddr: ":6000"}
	err := cfg.Save(configPath)
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(filepath.Join(tmpDir, "a", "b", "c"))
	assert.NoError(t, err)
}

func TestLoad_Combined(t *testing.T) {
	// Test that file values are used but env vars override
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	configContent := `
listen_addr: ":3000"
database_path: "/file/db"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment override
	os.Setenv("PRYX_DB_PATH", "/env/db")
	defer os.Unsetenv("PRYX_DB_PATH")

	// Unfortunately we can't easily test this since Load() uses DefaultPath()
	// This test documents the expected behavior
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback string
		expected string
	}{
		{"env set", "TEST_VAR", "value", "fallback", "value"},
		{"env not set", "UNSET_VAR", "", "fallback", "fallback"},
		{"env empty", "EMPTY_VAR", "", "fallback", "fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.fallback)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkLoad(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "pryx_home_bench_*")
	require.NoError(b, err)
	b.Cleanup(func() { _ = os.RemoveAll(tmpDir) })
	os.Setenv("HOME", tmpDir)

	os.Unsetenv("PRYX_LISTEN_ADDR")
	os.Unsetenv("PRYX_DB_PATH")
	os.Unsetenv("PRYX_CLOUD_API_URL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Load()
	}
}

func BenchmarkLoadFromFile(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
listen_addr: ":3000"
database_path: "pryx.db"
cloud_api_url: "https://pryx.dev/api"
model_provider: "ollama"
model_name: "llama3"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = LoadFromFile(configPath)
	}
}
