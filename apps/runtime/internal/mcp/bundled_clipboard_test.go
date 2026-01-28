package mcp

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func skipIfNoClipboardTool(t *testing.T) {
	var tool string
	switch runtime.GOOS {
	case "darwin":
		tool = "pbpaste"
	case "linux":
		tool = "xclip"
	case "windows":
		tool = "powershell"
	default:
		t.Skip("unsupported platform")
		return
	}
	if _, err := exec.LookPath(tool); err != nil {
		t.Skipf("%s not available: %v", tool, err)
	}
}

func TestClipboardProvider_ListTools(t *testing.T) {
	skipIfNoClipboardTool(t)

	p := NewClipboardProvider()
	tools, err := p.ListTools(context.Background())
	if err != nil {
		t.Fatalf("failed to list tools: %v", err)
	}

	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	toolNames := []string{tools[0].Name, tools[1].Name}
	if !contains(toolNames, "read_clipboard") {
		t.Errorf("missing read_clipboard tool")
	}
	if !contains(toolNames, "write_clipboard") {
		t.Errorf("missing write_clipboard tool")
	}
}

func TestClipboardProvider_WriteRead(t *testing.T) {
	skipIfNoClipboardTool(t)

	p := NewClipboardProvider()

	testContent := "test-content-12345"

	writeArgs := map[string]interface{}{
		"content": testContent,
		"format":  "text",
	}

	writeRes, err := p.CallTool(context.Background(), "write_clipboard", writeArgs)
	if err != nil {
		t.Fatalf("failed to write clipboard: %v", err)
	}
	if writeRes.IsError {
		t.Errorf("write_clipboard returned error")
	}

	readArgs := map[string]interface{}{
		"format": "text",
	}

	readRes, err := p.CallTool(context.Background(), "read_clipboard", readArgs)
	if err != nil {
		t.Fatalf("failed to read clipboard: %v", err)
	}
	if readRes.IsError {
		t.Errorf("read_clipboard returned error")
	}

	if !strings.Contains(string(readRes.StructuredContent), testContent) {
		t.Errorf("clipboard content mismatch")
	}
}

func TestClipboardProvider_Base64(t *testing.T) {
	skipIfNoClipboardTool(t)

	p := NewClipboardProvider()

	testContent := "base64-test-content"

	writeArgs := map[string]interface{}{
		"content": testContent,
		"format":  "text",
	}

	_, err := p.CallTool(context.Background(), "write_clipboard", writeArgs)
	if err != nil {
		t.Fatalf("failed to write clipboard: %v", err)
	}

	readArgs := map[string]interface{}{
		"format": "base64",
	}

	readRes, err := p.CallTool(context.Background(), "read_clipboard", readArgs)
	if err != nil {
		t.Fatalf("failed to read clipboard: %v", err)
	}

	if !strings.Contains(string(readRes.StructuredContent), "base64") {
		t.Errorf("expected base64 format in response")
	}
}

func TestClipboardProvider_ServerInfo(t *testing.T) {
	p := NewClipboardProvider()
	info := p.ServerInfo()

	if info["name"] != "pryx-core/clipboard" {
		t.Errorf("unexpected name: %v", info["name"])
	}
	if info["title"] != "Pryx Clipboard (Bundled)" {
		t.Errorf("unexpected title: %v", info["title"])
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
