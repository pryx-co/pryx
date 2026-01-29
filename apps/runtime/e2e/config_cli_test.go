package e2e

import (
	"strings"
	"testing"
)

func TestConfigCLI_Get(t *testing.T) {
	home := t.TempDir()
	_, _ = runPryxCoreWithEnv(t, home, nil, "config", "get", "database_path")
}

func TestConfigCLI_Set(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "config", "set", "listen_addr", ":4321")
	if code != 0 {
		t.Fatalf("config set failed (code %d):\n%s", code, out)
	}

	out, code = runPryxCoreWithEnv(t, home, nil, "config", "get", "listen_addr")
	if code != 0 {
		t.Fatalf("config get failed (code %d):\n%s", code, out)
	}
	if strings.TrimSpace(out) != ":4321" {
		t.Fatalf("expected listen_addr ':4321', got %q", strings.TrimSpace(out))
	}
}

func TestConfigCLI_List(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "config", "list")
	if code != 0 {
		t.Fatalf("config list failed (code %d):\n%s", code, out)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("Expected config list output")
	}
}

func TestConfigCLI_InvalidCommand(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "config", "invalid")
	if code == 0 {
		t.Error("Expected error for invalid config command")
	}

	if !strings.Contains(out, "Usage") && !strings.Contains(out, "invalid") {
		t.Logf("Output: %s", out)
	}
}

func TestConfigCLI_MissingValue(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "config", "set", "listen_addr")
	if code == 0 {
		t.Error("Expected error for missing value")
	}

	if strings.TrimSpace(out) == "" {
		t.Error("Expected error message")
	}
}
