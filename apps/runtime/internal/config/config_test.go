package config

import (
	"os"
	"testing"
)

func TestPartialEnvironment(t *testing.T) {
	// Set only one environment variable
	os.Setenv("PRYX_LISTEN_ADDR", "localhost:9000")
	defer os.Unsetenv("PRYX_LISTEN_ADDR")

	// Clear the other
	os.Unsetenv("PRYX_DB_PATH")

	config := Load()

	if config.ListenAddr != "localhost:9000" {
		t.Errorf("Expected ListenAddr 'localhost:9000', got '%s'", config.ListenAddr)
	}

	// DatabasePath should still have default
	if config.DatabasePath != "pryx.db" {
		t.Errorf("Expected DatabasePath 'pryx.db', got '%s'", config.DatabasePath)
	}
}
