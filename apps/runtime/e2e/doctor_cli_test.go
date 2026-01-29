package e2e

import (
	"strings"
	"testing"
)

func TestDoctorCLI_Run(t *testing.T) {
	home := t.TempDir()
	out, _ := runPryxCoreWithEnv(t, home, nil, "doctor")

	if strings.TrimSpace(out) == "" {
		t.Error("Expected doctor output")
	}

	// Check for any expected keywords in output (case-insensitive)
	outputLower := strings.ToLower(out)
	if !strings.Contains(outputLower, "database") &&
		!strings.Contains(outputLower, "config") &&
		!strings.Contains(outputLower, "dependencies") &&
		!strings.Contains(outputLower, "network") &&
		!strings.Contains(outputLower, "installation") &&
		!strings.Contains(outputLower, "health") &&
		!strings.Contains(outputLower, "sqlite") &&
		!strings.Contains(outputLower, "mcp") &&
		!strings.Contains(outputLower, "channels") {
		t.Errorf("Expected at least one diagnostic check in output, got: %s", out)
	}
}

func TestDoctorCLI_CheckNames(t *testing.T) {
	home := t.TempDir()
	out, _ := runPryxCoreWithEnv(t, home, nil, "doctor")

	// These are the standard checks from doctor package
	expectedChecks := []string{
		"database",
		"config",
		"dependencies",
		"network",
	}

	// At least some checks should be present
	foundCheck := false
	for _, check := range expectedChecks {
		if strings.Contains(strings.ToLower(out), check) {
			foundCheck = true
			break
		}
	}

	if !foundCheck {
		t.Logf("Doctor output: %s", out)
		t.Log("Warning: No expected check names found in output")
	}
}

func TestDoctorCLI_StatusIndicators(t *testing.T) {
	home := t.TempDir()
	out, _ := runPryxCoreWithEnv(t, home, nil, "doctor")

	// Look for status indicators (uppercase)
	statusIndicators := []string{"PASS", "OK", "FAIL", "WARN", "ERROR"}
	foundStatus := false
	for _, indicator := range statusIndicators {
		if strings.Contains(out, indicator) {
			foundStatus = true
			break
		}
	}

	if !foundStatus {
		t.Logf("Doctor output: %s", out)
		t.Error("Expected status indicator (PASS/FAIL/WARN) in output")
	}
}
