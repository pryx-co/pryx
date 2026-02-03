package doctor

import (
	"os"
	"testing"

	"pryx-core/internal/config"
)

func TestCheckInstallation(t *testing.T) {
	check := checkInstallation()

	if check.Name != "installation" {
		t.Errorf("Expected check name 'installation', got '%s'", check.Name)
	}

	// The executable should be accessible in test environment
	if check.Status != StatusOK && check.Status != StatusWarn {
		t.Errorf("Expected status OK or Warn, got %s", check.Status)
	}
}

func TestCheckDependencies(t *testing.T) {
	check := checkDependencies()

	if check.Name != "dependencies" {
		t.Errorf("Expected check name 'dependencies', got '%s'", check.Name)
	}

	// Most environments should have sh or warn
	if check.Status != StatusOK && check.Status != StatusWarn {
		t.Errorf("Expected status OK or Warn, got %s", check.Status)
	}
}

func TestCheckDatabase(t *testing.T) {
	// Create a temporary database for testing
	tmpFile, err := os.CreateTemp("", "pryx_doctor_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	cfg := &config.Config{
		DatabasePath: tmpFile.Name(),
	}

	check, dbConn := checkDatabase(cfg)

	if check.Name != "sqlite" {
		t.Errorf("Expected check name 'sqlite', got '%s'", check.Name)
	}

	if check.Status != StatusOK {
		t.Errorf("Expected status OK for valid database, got %s: %s", check.Status, check.Detail)
	}

	if dbConn == nil {
		t.Error("Expected database connection to be returned")
	} else {
		dbConn.Close()
	}
}

func TestCheckDatabaseMissingPath(t *testing.T) {
	cfg := &config.Config{
		DatabasePath: "",
	}

	check, dbConn := checkDatabase(cfg)

	if check.Status != StatusFail {
		t.Errorf("Expected status Fail for missing path, got %s", check.Status)
	}

	if dbConn != nil {
		t.Error("Expected nil database connection for missing path")
	}

	if check.Suggestion == "" {
		t.Error("Expected suggestion for failed check")
	}
}

func TestHealthURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{":3000", "http://127.0.0.1:3000/health"},
		{"localhost:8080", "http://localhost:8080/health"},
		{"0.0.0.0:9000", "http://0.0.0.0:9000/health"},
		{"http://localhost:3000", "http://localhost:3000/health"},
		{"https://localhost:3000", "https://localhost:3000/health"},
	}

	for _, test := range tests {
		result := healthURL(test.input)
		if result != test.expected {
			t.Errorf("healthURL(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestCheckChannelsNotFound(t *testing.T) {
	// Set workspace to a temporary directory with no channel config
	oldWd, _ := os.Getwd()
	oldHome := os.Getenv("HOME")
	oldUserProfile := os.Getenv("USERPROFILE")
	tmpDir, err := os.MkdirTemp("", "pryx_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}
	if err := os.Setenv("USERPROFILE", tmpDir); err != nil {
		t.Fatalf("Failed to set USERPROFILE: %v", err)
	}

	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	defer func() {
		_ = os.Setenv("HOME", oldHome)
		_ = os.Setenv("USERPROFILE", oldUserProfile)
	}()

	check := checkChannels()

	if check.Name != "channels" {
		t.Errorf("Expected check name 'channels', got '%s'", check.Name)
	}

	// Should warn when no channels config found
	if check.Status != StatusWarn {
		t.Errorf("Expected status Warn for missing channels config, got %s", check.Status)
	}
}
