package mcp

import (
	"context"
	"os"
	"testing"
	"time"

	"pryx-core/internal/bus"
)

func TestManager_LoadAndConnect_DefaultsToBundledTier1(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PRYX_WORKSPACE_ROOT", t.TempDir())

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	m := NewManager(bus.New(), nil, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = m.LoadAndConnect(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	tools, err := m.ListToolsFlat(ctx, true)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(tools) == 0 {
		t.Fatalf("expected some tools")
	}

	want := map[string]bool{
		"filesystem:read_file":     false,
		"shell:exec":               false,
		"browser:goto":             false,
		"clipboard:read_clipboard": false,
	}
	for _, tool := range tools {
		if _, ok := want[tool.Name]; ok {
			want[tool.Name] = true
		}
	}
	for k, ok := range want {
		if !ok {
			t.Fatalf("missing expected tool %q", k)
		}
	}
}
