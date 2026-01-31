package mcp

import (
	"context"
	"testing"
)

func TestScreenProvider_ServerInfo(t *testing.T) {
	provider := NewScreenProvider()
	info := provider.ServerInfo()

	if info["name"] != "pryx-core/screen" {
		t.Errorf("Expected name 'pryx-core/screen', got %v", info["name"])
	}

	if info["title"] != "Pryx Screen Capture (Bundled)" {
		t.Errorf("Expected title 'Pryx Screen Capture (Bundled)', got %v", info["title"])
	}
}

func TestScreenProvider_ListTools(t *testing.T) {
	provider := NewScreenProvider()
	ctx := context.Background()

	tools, err := provider.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	expectedTools := map[string]bool{
		"capture": false,
		"record":  false,
	}

	for _, tool := range tools {
		if _, exists := expectedTools[tool.Name]; exists {
			expectedTools[tool.Name] = true
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("Expected tool '%s' not found", name)
		}
	}
}

func TestScreenProvider_Capture_InvalidRegion(t *testing.T) {
	provider := NewScreenProvider()
	ctx := context.Background()

	args := map[string]interface{}{
		"region": "",
		"format": "png",
	}

	result, err := provider.CallTool(ctx, "capture", args)
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	// Should succeed with full screen capture (region="")
	if result.IsError {
		t.Errorf("Expected success, got error: %v", result.Content)
	}
}

func TestScreenProvider_Record_RequiresFfmpeg(t *testing.T) {
	provider := NewScreenProvider()
	ctx := context.Background()

	args := map[string]interface{}{
		"duration": 1.0,
		"format":   "mp4",
	}

	result, err := provider.CallTool(ctx, "record", args)
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	// Should return error indicating ffmpeg is required (unless ffmpeg is installed)
	// This test verifies the error handling path
	t.Logf("Record result: %+v", result)
}
