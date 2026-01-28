package mcp

import (
	"context"
	"testing"
)

func TestBrowserProvider_ListTools(t *testing.T) {
	p := NewBrowserProvider()

	tools, err := p.ListTools(context.Background())
	if err != nil {
		t.Fatalf("failed to list tools: %v", err)
	}

	found := map[string]bool{}
	for _, tool := range tools {
		found[tool.Name] = true
	}
	for _, required := range []string{"install", "goto", "content", "screenshot", "evaluate"} {
		if !found[required] {
			t.Fatalf("missing tool %q", required)
		}
	}
}

func TestBrowserProvider_ServerInfo(t *testing.T) {
	p := NewBrowserProvider()

	info := p.ServerInfo()

	if info["name"] != "pryx-core/browser" {
		t.Errorf("unexpected name: %v", info["name"])
	}
	if info["title"] != "Pryx Browser (Bundled)" {
		t.Errorf("unexpected title: %v", info["title"])
	}
}

func TestBrowserProvider_UnknownTool(t *testing.T) {
	p := NewBrowserProvider()
	_, err := p.CallTool(context.Background(), "nope", map[string]interface{}{})
	if err == nil {
		t.Fatalf("expected error")
	}
}
