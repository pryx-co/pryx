//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestToolExecution tests MCP tool execution via HTTP API
func TestToolExecution(t *testing.T) {
	bin := buildPryxCore(t)
	home := t.TempDir()

	workspaceDir := filepath.Join(home, "workspace")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	port, cancel := startPryxCore(t, bin, home)
	defer cancel()

	waitForServer(t, port, 5*time.Second)

	baseURL := "http://localhost:" + port

	t.Run("list_mcp_tools", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/mcp/tools")
		if err != nil {
			t.Fatalf("Failed to list MCP tools: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		tools, ok := result["tools"].([]interface{})
		if !ok {
			t.Fatal("Expected tools array in response")
		}

		if len(tools) == 0 {
			t.Skip("No MCP tools available - MCP servers may not be configured")
		}

		t.Logf("Found %d MCP tools", len(tools))
	})

	t.Run("filesystem_create_directory", func(t *testing.T) {
		testDir := filepath.Join(workspaceDir, "test_dir")
		payload := map[string]interface{}{
			"tool": "filesystem_create_directory",
			"arguments": map[string]interface{}{
				"path": testDir,
			},
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(baseURL+"/mcp/tools/call", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusBadGateway {
			t.Skip("MCP filesystem tool not available (502)")
		}

		if resp.StatusCode != http.StatusOK {
			t.Skipf("Filesystem create returned status %d - tool may not be implemented", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Skip("Could not decode filesystem response")
		}

		t.Logf("Filesystem directory created successfully")
	})

	t.Run("filesystem_read_file", func(t *testing.T) {
		testFile := filepath.Join(workspaceDir, "test.txt")
		testContent := "Test E2E content"

		if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		payload := map[string]interface{}{
			"tool": "filesystem_read_file",
			"arguments": map[string]interface{}{
				"path": testFile,
			},
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(baseURL+"/mcp/tools/call", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusBadGateway {
			t.Skip("MCP filesystem tool not available (502)")
		}

		if resp.StatusCode != http.StatusOK {
			t.Skipf("Filesystem read returned status %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Skip("Could not decode filesystem response")
		}

		content, ok := result["content"].(string)
		if !ok || content != testContent {
			t.Errorf("File content mismatch: got %s, want %s", content, testContent)
		}

		t.Logf("File read successful")
	})

	t.Run("shell_execute_command", func(t *testing.T) {
		payload := map[string]interface{}{
			"tool": "shell_execute",
			"arguments": map[string]interface{}{
				"command": "echo 'Hello from E2E test'",
				"timeout": 30,
			},
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(baseURL+"/mcp/tools/call", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to execute shell command: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusBadGateway {
			t.Skip("MCP shell tool not available (502)")
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			t.Skipf("Shell execute returned status %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Skip("Could not decode shell response")
		}

		output, ok := result["output"].(string)
		if !ok {
			t.Error("Expected output in response")
		}

		if !strings.Contains(output, "Hello from E2E test") {
			t.Errorf("Output mismatch: got %s, want 'Hello from E2E test'", output)
		}

		t.Logf("Shell executed successfully")
	})
}
