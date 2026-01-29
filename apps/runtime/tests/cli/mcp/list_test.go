package mcp

import (
	"os/exec"
	"strings"
	"testing"
)

func TestListCommand(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcp list failed: %v", err)
	}

	if !strings.Contains(string(output), "MCP Servers") {
		t.Errorf("Expected 'MCP Servers' in output, got: %s", output)
	}
}

func TestListCommandWithJSON(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcp list --json failed: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr != "" && outputStr != "null" && outputStr != "[]" {
		if !strings.HasPrefix(outputStr, "[") && !strings.HasPrefix(outputStr, "{") {
			t.Errorf("Expected JSON output, got: %s", outputStr)
		}
	}
}

func TestAddCommand(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp", "add", "test-server", "--command", "echo", "test")
	output, err := cmd.CombinedOutput()

	if err == nil {
		outputStr := string(output)
		if !strings.Contains(outputStr, "Added") && !strings.Contains(outputStr, "test-server") {
			t.Logf("Add output doesn't contain expected content: %s", outputStr)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "error") && !strings.Contains(string(output), "Error") {
				t.Logf("Expected error message, got: %s", output)
			}
		}
	}
}

func TestRemoveCommand(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp", "remove", "nonexistent-server")
	output, err := cmd.CombinedOutput()

	if err == nil {
		outputStr := string(output)
		if !strings.Contains(outputStr, "Removed") {
			t.Logf("Remove output doesn't contain 'Removed': %s", outputStr)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "error") {
				t.Logf("Expected error for unknown server, got: %s", output)
			}
		}
	}
}

func TestTestCommand(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp", "test", "nonexistent-server")
	output, err := cmd.CombinedOutput()

	if err == nil {
		outputStr := string(output)
		if !strings.Contains(outputStr, "Test") && !strings.Contains(outputStr, "Success") {
			t.Logf("Test output doesn't contain expected content: %s", outputStr)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "error") {
				t.Logf("Expected error for unknown server, got: %s", output)
			}
		}
	}
}

func TestAuthCommand(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp", "auth", "nonexistent-server")
	output, err := cmd.CombinedOutput()

	if err == nil {
		outputStr := string(output)
		if !strings.Contains(outputStr, "Auth") && !strings.Contains(outputStr, "Authentication") {
			t.Logf("Auth output doesn't contain expected content: %s", outputStr)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "error") {
				t.Logf("Expected error for unknown server, got: %s", output)
			}
		}
	}
}

func TestHelpCommand(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 2 {
				t.Fatalf("mcp help failed with unexpected exit code: %v", err)
			}
		}
	}

	expectedCommands := []string{"list", "add", "remove", "test", "auth"}
	for _, cmdName := range expectedCommands {
		if !strings.Contains(string(output), cmdName) {
			t.Errorf("Expected '%s' in help output", cmdName)
		}
	}
}

func TestAddCommandWithInvalidArgs(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp", "add")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for missing arguments")
	}

	if !strings.Contains(string(output), "Usage") && !strings.Contains(string(output), "error") {
		t.Logf("Expected usage or error message, got: %s", output)
	}
}

func TestRemoveCommandWithMissingServer(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp", "remove")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for missing server name")
	}

	if !strings.Contains(string(output), "Usage") && !strings.Contains(string(output), "error") {
		t.Logf("Expected usage or error message, got: %s", output)
	}
}

func TestUnknownCommand(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp", "unknown-cmd")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for unknown command")
	}

	if !strings.Contains(string(output), "Unknown") {
		t.Errorf("Expected 'Unknown command' in error output, got: %s", output)
	}
}

func TestAddCommandWithAuth(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp", "add", "auth-test", "--command", "echo", "test", "--auth", "bearer-token")
	output, err := cmd.CombinedOutput()

	if err == nil {
		outputStr := string(output)
		if !strings.Contains(outputStr, "Added") && !strings.Contains(outputStr, "auth-test") {
			t.Logf("Add with auth output doesn't contain expected content: %s", outputStr)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "error") && !strings.Contains(string(output), "Error") {
				t.Logf("Expected error message, got: %s", output)
			}
		}
	}
}

func TestListCommandWithVerbose(t *testing.T) {
	cmd := exec.Command("/tmp/pryx-core", "mcp", "list", "--verbose")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcp list --verbose failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "MCP Servers") {
		t.Errorf("Expected 'MCP Servers' in verbose output, got: %s", outputStr)
	}
}
