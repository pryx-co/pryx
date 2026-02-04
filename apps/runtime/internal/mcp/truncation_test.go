package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewTruncator(t *testing.T) {
	tmpDir := t.TempDir()
	truncator := NewTruncator(tmpDir)

	if truncator == nil {
		t.Fatal("NewTruncator() returned nil")
	}

	expectedDir := filepath.Join(tmpDir, DefaultOutputDir)
	if truncator.outputDir != expectedDir {
		t.Errorf("outputDir = %s, want %s", truncator.outputDir, expectedDir)
	}

	if truncator.maxLines != MaxLines {
		t.Errorf("maxLines = %d, want %d", truncator.maxLines, MaxLines)
	}

	if truncator.maxBytes != MaxBytes {
		t.Errorf("maxBytes = %d, want %d", truncator.maxBytes, MaxBytes)
	}
}

func TestTruncator_Process_NoTruncation(t *testing.T) {
	tmpDir := t.TempDir()
	truncator := NewTruncator(tmpDir)

	content := "line1\nline2\nline3"
	result, err := truncator.Process(content)

	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if result.Truncated {
		t.Error("Truncated = true, want false for small content")
	}

	if result.Content != content {
		t.Errorf("Content = %s, want %s", result.Content, content)
	}

	if result.OutputPath != "" {
		t.Error("OutputPath should be empty when not truncated")
	}
}

func TestTruncator_Process_TruncateByLines(t *testing.T) {
	tmpDir := t.TempDir()
	truncator := NewTruncator(tmpDir)
	truncator.maxLines = 5
	truncator.maxBytes = MaxBytes

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}
	content := strings.Join(lines, "\n")

	result, err := truncator.Process(content)

	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if !result.Truncated {
		t.Error("Truncated = false, want true for large content")
	}

	if result.OutputPath == "" {
		t.Error("OutputPath should not be empty when truncated")
	}

	if !strings.Contains(result.Content, "output truncated") {
		t.Error("Content should contain truncation hint")
	}

	if !strings.Contains(result.Content, "Use Grep to search") {
		t.Error("Content should contain hint about Grep")
	}
}

func TestTruncator_Process_TruncateByBytes(t *testing.T) {
	tmpDir := t.TempDir()
	truncator := NewTruncator(tmpDir)
	truncator.maxLines = MaxLines
	truncator.maxBytes = 100

	content := strings.Repeat("a", 1000)

	result, err := truncator.Process(content)

	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if !result.Truncated {
		t.Error("Truncated = false, want true for large byte content")
	}

	if result.OutputPath == "" {
		t.Error("OutputPath should not be empty when truncated")
	}
}

func TestTruncator_Process_SavesFullContent(t *testing.T) {
	tmpDir := t.TempDir()
	truncator := NewTruncator(tmpDir)
	truncator.maxLines = 5
	truncator.maxBytes = MaxBytes

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line content"
	}
	content := strings.Join(lines, "\n")

	result, err := truncator.Process(content)

	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if !result.Truncated {
		t.Fatal("Content should be truncated")
	}

	savedContent, err := os.ReadFile(result.OutputPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(savedContent) != content {
		t.Error("Saved content does not match original")
	}
}

func TestTruncator_ProcessToolResult_NoTruncation(t *testing.T) {
	tmpDir := t.TempDir()
	truncator := NewTruncator(tmpDir)

	result := ToolResult{
		Content: []ToolContent{
			{Type: "text", Text: "small content"},
		},
		IsError: false,
	}

	processed := truncator.ProcessToolResult(result)

	if len(processed.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(processed.Content))
	}

	if processed.Content[0].Text != "small content" {
		t.Error("Content should not be modified")
	}
}

func TestTruncator_ProcessToolResult_WithTruncation(t *testing.T) {
	tmpDir := t.TempDir()
	truncator := NewTruncator(tmpDir)
	truncator.maxLines = 5
	truncator.maxBytes = MaxBytes

	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line content"
	}
	largeContent := strings.Join(lines, "\n")

	result := ToolResult{
		Content: []ToolContent{
			{Type: "text", Text: largeContent},
		},
		IsError: false,
	}

	processed := truncator.ProcessToolResult(result)

	if len(processed.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(processed.Content))
	}

	if !strings.Contains(processed.Content[0].Text, "output truncated") {
		t.Error("Content should contain truncation hint")
	}
}

func TestTruncator_ProcessToolResult_ErrorResult(t *testing.T) {
	tmpDir := t.TempDir()
	truncator := NewTruncator(tmpDir)

	result := ToolResult{
		Content: []ToolContent{
			{Type: "text", Text: strings.Repeat("error ", 10000)},
		},
		IsError: true,
	}

	processed := truncator.ProcessToolResult(result)

	if processed.Content[0].Text != result.Content[0].Text {
		t.Error("Error results should not be truncated")
	}
}

func TestTruncator_ProcessToolResult_NonTextContent(t *testing.T) {
	tmpDir := t.TempDir()
	truncator := NewTruncator(tmpDir)

	result := ToolResult{
		Content: []ToolContent{
			{Type: "image", Data: "base64data", MimeType: "image/png"},
		},
		IsError: false,
	}

	processed := truncator.ProcessToolResult(result)

	if processed.Content[0].Type != "image" {
		t.Error("Non-text content should not be modified")
	}
}

func TestTruncator_CleanupOldFiles(t *testing.T) {
	tmpDir := t.TempDir()
	truncator := NewTruncator(tmpDir)

	if err := os.MkdirAll(truncator.outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	oldFile := filepath.Join(truncator.outputDir, "old_file.txt")
	if err := os.WriteFile(oldFile, []byte("old content"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}

	oldTime := time.Now().Add(-8 * 24 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set file time: %v", err)
	}

	recentFile := filepath.Join(truncator.outputDir, "recent_file.txt")
	if err := os.WriteFile(recentFile, []byte("recent content"), 0644); err != nil {
		t.Fatalf("Failed to create recent file: %v", err)
	}

	if err := truncator.CleanupOldFiles(); err != nil {
		t.Fatalf("CleanupOldFiles() error = %v", err)
	}

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file should have been deleted")
	}

	if _, err := os.Stat(recentFile); os.IsNotExist(err) {
		t.Error("Recent file should still exist")
	}
}

func TestGlobalTruncator(t *testing.T) {
	tmpDir := t.TempDir()

	InitTruncator(tmpDir)

	truncator := GetTruncator()
	if truncator == nil {
		t.Fatal("GetTruncator() returned nil after InitTruncator()")
	}

	expectedDir := filepath.Join(tmpDir, DefaultOutputDir)
	if truncator.GetOutputDir() != expectedDir {
		t.Errorf("GetOutputDir() = %s, want %s", truncator.GetOutputDir(), expectedDir)
	}

	result := ToolResult{
		Content: []ToolContent{
			{Type: "text", Text: "test"},
		},
	}

	processed := TruncateToolResult(result)
	if len(processed.Content) != 1 {
		t.Fatal("TruncateToolResult failed")
	}
}

func TestCleanup_NoInit(t *testing.T) {
	if err := Cleanup(); err != nil {
		t.Errorf("Cleanup() with no init should return nil, got %v", err)
	}
}
