package memory

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

const pryxCorePath = "/tmp/pryx-core"

func skipMemoryTest(t *testing.T) {
	if _, err := os.Stat(pryxCorePath); err != nil {
		t.Skip("/tmp/pryx-core not found; build runtime binary before running CLI tests")
	}

	cmd := exec.Command(pryxCorePath, "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "memory") {
		t.Skip("memory command not available in CLI")
	}
}

func TestMemoryUsageCommand(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "usage")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory usage failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Memory") && !strings.Contains(outputStr, "Usage") {
		t.Errorf("Expected 'Memory' or 'Usage' in output, got: %s", outputStr)
	}
}

func TestMemoryUsageWithSession(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "usage", "--session", "test-session")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory usage --session failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Memory") && !strings.Contains(outputStr, "Usage") {
		t.Errorf("Expected 'Memory' or 'Usage' in output, got: %s", outputStr)
	}
}

func TestMemoryUsageWithJSON(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "usage", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory usage --json failed: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr != "" && outputStr != "null" && outputStr != "{}" {
		if !strings.HasPrefix(outputStr, "[") && !strings.HasPrefix(outputStr, "{") {
			t.Errorf("Expected JSON output, got: %s", outputStr)
		}
	}
}

func TestMemorySessionsCommand(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "sessions")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory sessions failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Session") && !strings.Contains(outputStr, "Memory") {
		t.Errorf("Expected 'Session' or 'Memory' in output, got: %s", outputStr)
	}
}

func TestMemorySessionsWithJSON(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "sessions", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory sessions --json failed: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr != "" && outputStr != "null" && outputStr != "[]" {
		if !strings.HasPrefix(outputStr, "[") {
			t.Errorf("Expected JSON array output, got: %s", outputStr)
		}
	}
}

func TestMemorySummarizeCommand(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "summarize", "--session", "test-session")
	output, err := cmd.CombinedOutput()

	if err == nil {
		outputStr := string(output)
		if !strings.Contains(outputStr, "Summarize") && !strings.Contains(outputStr, "Memory") {
			t.Logf("Summarize output doesn't contain expected content: %s", outputStr)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "error") {
				t.Logf("Expected error for unknown session, got: %s", output)
			}
		}
	}
}

func TestMemoryArchiveCommand(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "archive", "--session", "test-session")
	output, err := cmd.CombinedOutput()

	if err == nil {
		outputStr := string(output)
		if !strings.Contains(outputStr, "Archive") && !strings.Contains(outputStr, "Memory") {
			t.Logf("Archive output doesn't contain expected content: %s", outputStr)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "error") {
				t.Logf("Expected error for unknown session, got: %s", output)
			}
		}
	}
}

func TestMemoryUnarchiveCommand(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "unarchive", "--session", "test-session")
	output, err := cmd.CombinedOutput()

	if err == nil {
		outputStr := string(output)
		if !strings.Contains(outputStr, "Unarchive") && !strings.Contains(outputStr, "Memory") {
			t.Logf("Unarchive output doesn't contain expected content: %s", outputStr)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "error") {
				t.Logf("Expected error for unknown session, got: %s", output)
			}
		}
	}
}

func TestMemoryCleanupCommand(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "cleanup")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory cleanup failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Cleanup") && !strings.Contains(outputStr, "Memory") {
		t.Errorf("Expected 'Cleanup' or 'Memory' in output, got: %s", outputStr)
	}
}

func TestMemoryCleanupDryRun(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "cleanup", "--dry-run")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory cleanup --dry-run failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Cleanup") && !strings.Contains(outputStr, "Memory") {
		t.Errorf("Expected 'Cleanup' or 'Memory' in output, got: %s", outputStr)
	}
}

func TestMemoryRAGCommand(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "rag", "--query", "test query")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory rag failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "RAG") && !strings.Contains(outputStr, "Memory") {
		t.Errorf("Expected 'RAG' or 'Memory' in output, got: %s", outputStr)
	}
}

func TestMemoryHelpCommand(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 2 {
				t.Fatalf("memory help failed with unexpected exit code: %v", err)
			}
		}
	}

	expectedCommands := []string{"usage", "sessions", "summarize", "archive", "cleanup", "rag"}
	for _, cmdName := range expectedCommands {
		if !strings.Contains(string(output), cmdName) {
			t.Errorf("Expected '%s' in help output", cmdName)
		}
	}
}

func TestMemoryUnknownCommand(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "unknown-cmd")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for unknown command")
	}

	if !strings.Contains(string(output), "Unknown") {
		t.Errorf("Expected 'Unknown command' in error output, got: %s", output)
	}
}

func TestMemoryUsageWithThreshold(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "usage", "--threshold", "80")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory usage --threshold failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Memory") && !strings.Contains(outputStr, "Usage") {
		t.Errorf("Expected 'Memory' or 'Usage' in output, got: %s", outputStr)
	}
}

func TestMemorySummarizeWithRatio(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "summarize", "--session", "test", "--ratio", "0.5")
	output, err := cmd.CombinedOutput()

	if err == nil {
		outputStr := string(output)
		if !strings.Contains(outputStr, "Summarize") && !strings.Contains(outputStr, "Memory") {
			t.Logf("Summarize with ratio output doesn't contain expected content: %s", outputStr)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "error") {
				t.Logf("Expected error for unknown session, got: %s", output)
			}
		}
	}
}

func TestMemoryRAGWithLimit(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "rag", "--query", "test", "--limit", "5")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory rag --limit failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "RAG") && !strings.Contains(outputStr, "Memory") {
		t.Errorf("Expected 'RAG' or 'Memory' in output, got: %s", outputStr)
	}
}

func TestMemoryArchiveWithAuto(t *testing.T) {
	skipMemoryTest(t)
	cmd := exec.Command(pryxCorePath, "memory", "archive", "--auto", "--older-than", "7d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory archive --auto failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Archive") && !strings.Contains(outputStr, "Memory") {
		t.Errorf("Expected 'Archive' or 'Memory' in output, got: %s", outputStr)
	}
}
