package audit

import (
	"os/exec"
	"strings"
	"testing"
)

func TestAuditQueryCommand(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "query")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit query failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Audit") && !strings.Contains(outputStr, "Logs") {
		t.Errorf("Expected 'Audit' or 'Logs' in output, got: %s", outputStr)
	}
}

func TestAuditQueryWithSession(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "query", "--session", "test-session")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit query --session failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Audit") && !strings.Contains(outputStr, "Logs") {
		t.Errorf("Expected 'Audit' or 'Logs' in output, got: %s", outputStr)
	}
}

func TestAuditQueryWithSurface(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "query", "--surface", "cli")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit query --surface failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Audit") && !strings.Contains(outputStr, "Logs") {
		t.Errorf("Expected 'Audit' or 'Logs' in output, got: %s", outputStr)
	}
}

func TestAuditQueryWithJSON(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "query", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit query --json failed: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr != "" && outputStr != "null" && outputStr != "[]" {
		if !strings.HasPrefix(outputStr, "[") && !strings.HasPrefix(outputStr, "{") {
			t.Errorf("Expected JSON output, got: %s", outputStr)
		}
	}
}

func TestAuditQueryWithLimit(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "query", "--limit", "10")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit query --limit failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Audit") && !strings.Contains(outputStr, "Logs") {
		t.Errorf("Expected 'Audit' or 'Logs' in output, got: %s", outputStr)
	}
}

func TestAuditExportCommand(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "export")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit export failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Export") && !strings.Contains(outputStr, "Audit") {
		t.Errorf("Expected 'Export' or 'Audit' in output, got: %s", outputStr)
	}
}

func TestAuditExportToJSON(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "export", "--format", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit export --format json failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Export") {
		t.Errorf("Expected 'Export' in output, got: %s", outputStr)
	}
}

func TestAuditExportToCSV(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "export", "--format", "csv")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit export --format csv failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Export") {
		t.Errorf("Expected 'Export' in output, got: %s", outputStr)
	}
}

func TestAuditExportToJSONL(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "export", "--format", "jsonl")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit export --format jsonl failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Export") {
		t.Errorf("Expected 'Export' in output, got: %s", outputStr)
	}
}

func TestAuditStatsCommand(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "stats")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit stats failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Stats") && !strings.Contains(outputStr, "Statistics") {
		t.Errorf("Expected 'Stats' or 'Statistics' in output, got: %s", outputStr)
	}
}

func TestAuditHelpCommand(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 2 {
				t.Fatalf("audit help failed with unexpected exit code: %v", err)
			}
		}
	}

	expectedCommands := []string{"query", "export", "stats"}
	for _, cmdName := range expectedCommands {
		if !strings.Contains(string(output), cmdName) {
			t.Errorf("Expected '%s' in help output", cmdName)
		}
	}
}

func TestAuditQueryWithTimeFilter(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "query", "--since", "1h")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit query --since failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Audit") && !strings.Contains(outputStr, "Logs") {
		t.Errorf("Expected 'Audit' or 'Logs' in output, got: %s", outputStr)
	}
}

func TestAuditQueryWithToolFilter(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "query", "--tool", "bash")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit query --tool failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Audit") && !strings.Contains(outputStr, "Logs") {
		t.Errorf("Expected 'Audit' or 'Logs' in output, got: %s", outputStr)
	}
}

func TestAuditQueryWithMultipleFilters(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "query", "--session", "test", "--surface", "cli", "--limit", "5")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit query with multiple filters failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Audit") && !strings.Contains(outputStr, "Logs") {
		t.Errorf("Expected 'Audit' or 'Logs' in output, got: %s", outputStr)
	}
}

func TestAuditUnknownCommand(t *testing.T) {
	// Check if audit command exists
	cmd := exec.Command("/tmp/pryx-core", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "audit") {
		t.Skip("audit command not available in CLI")
	}

	cmd = exec.Command("/tmp/pryx-core", "audit", "unknown-cmd")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for unknown command")
	}

	if !strings.Contains(string(output), "Unknown") {
		t.Errorf("Expected 'Unknown command' in error output, got: %s", output)
	}
}
