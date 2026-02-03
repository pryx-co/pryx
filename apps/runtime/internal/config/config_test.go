package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPartialEnvironment(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Set only one environment variable
	t.Setenv("PRYX_LISTEN_ADDR", "localhost:9000")

	// Clear the other
	os.Unsetenv("PRYX_DB_PATH")

	config := Load()

	if config.ListenAddr != "localhost:9000" {
		t.Errorf("Expected ListenAddr 'localhost:9000', got '%s'", config.ListenAddr)
	}

	// DatabasePath should still have default
	expected := filepath.Join(home, ".pryx", "pryx.db")
	if config.DatabasePath != expected {
		t.Errorf("Expected DatabasePath '%s', got '%s'", expected, config.DatabasePath)
	}
}
