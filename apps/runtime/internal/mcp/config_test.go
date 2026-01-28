package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadServersConfigFromFirstExisting_None(t *testing.T) {
	cfg, path, err := LoadServersConfigFromFirstExisting([]string{
		filepath.Join(t.TempDir(), "missing.json"),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path, got %q", path)
	}
	if cfg == nil || cfg.Servers == nil || len(cfg.Servers) != 0 {
		t.Fatalf("expected empty servers map")
	}
}

func TestLoadServersConfigFromFirstExisting_Parse(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "servers.json")
	if err := os.WriteFile(p, []byte(`{"servers":{"x":{"transport":"http","url":"http://example/mcp"}}}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, path, err := LoadServersConfigFromFirstExisting([]string{p})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if path != p {
		t.Fatalf("expected path %q, got %q", p, path)
	}
	if cfg.Servers["x"].Transport != "http" {
		t.Fatalf("expected server x transport http, got %q", cfg.Servers["x"].Transport)
	}
}
