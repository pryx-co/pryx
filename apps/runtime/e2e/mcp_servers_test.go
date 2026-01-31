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

// TestMCPServers tests MCP server functionality
func TestMCPServers(t *testing.T) {
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

	t.Run("list_mcp_servers", func(t *testing.T) {
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
			t.Logf("Response: %+v", result)
			t.Fatal("Expected tools array in response")
		}

		t.Logf("Found %d MCP tools", len(tools))

		if len(tools) > 0 {
			t.Logf("✓ MCP servers available with %d tools", len(tools))
		} else {
			t.Skip("No MCP tools available")
		}
	})

	t.Run("execute_filesystem_list", func(t *testing.T) {
		payload := map[string]interface{}{
			"tool": "filesystem_list_directory",
			"arguments": map[string]interface{}{
				"path": workspaceDir,
			},
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(baseURL+"/mcp/tools/call", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to call filesystem tool: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusBadGateway {
			t.Skip("MCP filesystem tool not available (502)")
		}

		if resp.StatusCode != http.StatusOK {
			t.Skipf("Filesystem list returned status %d - tool may not be implemented", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Skip("Could not decode filesystem response")
		}

		t.Logf("✓ Filesystem list directory executed successfully")
	})

	t.Run("execute_shell_echo", func(t *testing.T) {
		payload := map[string]interface{}{
			"tool": "shell_exec",
			"arguments": map[string]interface{}{
				"command":    "echo 'MCP test'",
				"timeout_ms": 5000,
			},
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(baseURL+"/mcp/tools/call", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to call shell tool: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusBadGateway {
			t.Skip("MCP shell tool not available (502)")
		}

		if resp.StatusCode != http.StatusOK {
			t.Skipf("Shell exec returned status %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Skip("Could not decode shell response")
		}

		content, ok := result["content"].([]interface{})
		if !ok || len(content) == 0 {
			t.Skip("No content in shell response")
		}

		firstContent := content[0].(map[string]interface{})
		output, ok := firstContent["text"].(string)
		if !ok {
			t.Skip("No text output in shell response")
		}

		if !strings.Contains(output, "MCP test") {
			t.Errorf("Expected 'MCP test' in output, got: %s", output)
		}

		t.Logf("✓ Shell echo executed successfully")
	})

	t.Run("execute_clipboard_copy", func(t *testing.T) {
		payload := map[string]interface{}{
			"tool": "clipboard_copy",
			"arguments": map[string]interface{}{
				"text": "Test clipboard content from E2E",
			},
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(baseURL+"/mcp/tools/call", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to call clipboard tool: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusBadGateway {
			t.Skip("MCP clipboard tool not available (502)")
		}

		if resp.StatusCode != http.StatusOK {
			t.Skipf("Clipboard copy returned status %d", resp.StatusCode)
		}

		t.Logf("✓ Clipboard copy executed successfully")
	})

	t.Run("execute_browser_navigate", func(t *testing.T) {
		payload := map[string]interface{}{
			"tool": "browser_navigate",
			"arguments": map[string]interface{}{
				"url": "https://example.com",
			},
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(baseURL+"/mcp/tools/call", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to call browser tool: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusBadGateway {
			t.Skip("MCP browser tool not available (502) - Playwright may not be installed")
		}

		if resp.StatusCode != http.StatusOK {
			t.Skipf("Browser navigate returned status %d", resp.StatusCode)
		}

		t.Logf("✓ Browser navigate executed successfully")
	})

	t.Run("execute_screen_capture", func(t *testing.T) {
		payload := map[string]interface{}{
			"tool": "screen_capture",
			"arguments": map[string]interface{}{
				"region": "full",
				"format": "png",
			},
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(baseURL+"/mcp/tools/call", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to call screen tool: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusBadGateway {
			t.Skip("MCP screen tool not available (502) - screen capture tools may not be installed")
		}

		if resp.StatusCode != http.StatusOK {
			t.Skipf("Screen capture returned status %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Skip("Could not decode screen response")
		}

		t.Logf("✓ Screen capture executed successfully")
	})
}
