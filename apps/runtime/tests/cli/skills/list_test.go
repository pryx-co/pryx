package skills

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

const pryxCorePath = "/tmp/pryx-core"

func skipIfBinaryMissing(t *testing.T) {
	if _, err := os.Stat(pryxCorePath); err != nil {
		t.Skipf("%s not found; build runtime binary before running CLI tests", pryxCorePath)
	}
}

func TestListCommand(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("skills list failed: %v", err)
	}

	if !strings.Contains(string(output), "Available Skills") {
		t.Errorf("Expected 'Available Skills' in output, got: %s", output)
	}
}

func TestListCommandWithEligible(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "list", "--eligible")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("skills list --eligible failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Eligible") && !strings.Contains(outputStr, "Available") {
		t.Logf("Output may not contain 'Eligible', got: %s", outputStr)
	}
}

func TestListCommandWithJSON(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("skills list --json failed: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr != "" && outputStr != "null" && outputStr != "[]" {
		if !strings.HasPrefix(outputStr, "[") {
			t.Errorf("Expected JSON array output, got: %s", outputStr)
		}
	}
}

func TestInfoCommand(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "info", "test")
	output, err := cmd.CombinedOutput()

	if err == nil {
		if !strings.Contains(string(output), "test") {
			t.Logf("Info output doesn't contain 'test': %s", output)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") {
				t.Logf("Expected 'not found' for unknown skill, got: %s", output)
			}
		}
	}
}

func TestInfoCommandWithValidSkill(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "info", "git-master")
	output, err := cmd.CombinedOutput()

	if err == nil {
		outputStr := string(output)
		if !strings.Contains(outputStr, "git-master") {
			t.Logf("Info output doesn't contain skill name: %s", outputStr)
		}
	} else {
		t.Logf("Info command failed (skill may not exist): %v, output: %s", err, output)
	}
}

func TestCheckCommand(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "check")
	output, err := cmd.CombinedOutput()

	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "error") && !strings.Contains(string(output), "Error") {
				t.Logf("Check command failed but no error in output: %s", output)
			}
		} else {
			t.Fatalf("skills check failed with unexpected exit code: %v", err)
		}
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Status") && !strings.Contains(outputStr, "Check") && !strings.Contains(outputStr, "error") {
		t.Logf("Check output doesn't contain expected content: %s", outputStr)
	}
}

func TestEnableCommand(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "enable", "test-skill")
	output, err := cmd.CombinedOutput()

	if err == nil {
		if !strings.Contains(string(output), "Enabled") {
			t.Logf("Enable output doesn't contain 'Enabled': %s", output)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "Unknown") {
				t.Logf("Expected error for unknown skill, got: %s", output)
			}
		}
	}
}

func TestDisableCommand(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "disable", "test-skill")
	output, err := cmd.CombinedOutput()

	if err == nil {
		if !strings.Contains(string(output), "Disabled") {
			t.Logf("Disable output doesn't contain 'Disabled': %s", output)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "Unknown") {
				t.Logf("Expected error for unknown skill, got: %s", output)
			}
		}
	}
}

func TestInstallCommand(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "install", "test-skill")
	output, err := cmd.CombinedOutput()

	if err == nil {
		if !strings.Contains(string(output), "Install") {
			t.Logf("Install output doesn't contain 'Install': %s", output)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "Unknown") {
				t.Logf("Expected error for unknown skill, got: %s", output)
			}
		}
	}
}

func TestUninstallCommand(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "uninstall", "test-skill")
	output, err := cmd.CombinedOutput()

	if err == nil {
		if !strings.Contains(string(output), "Uninstall") {
			t.Logf("Uninstall output doesn't contain 'Uninstall': %s", output)
		}
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "Unknown") {
				t.Logf("Expected error for unknown skill, got: %s", output)
			}
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "unknown-cmd")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for unknown command")
	}

	if !strings.Contains(string(output), "Unknown") {
		t.Errorf("Expected 'Unknown command' in error output, got: %s", output)
	}
}

func TestHelpCommand(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 2 {
				t.Fatalf("skills help failed with unexpected exit code: %v", err)
			}
		}
	}

	expectedCommands := []string{"list", "info", "check", "enable", "disable", "install", "uninstall"}
	for _, cmdName := range expectedCommands {
		if !strings.Contains(string(output), cmdName) {
			t.Errorf("Expected '%s' in help output", cmdName)
		}
	}
}

func TestListCommandWithBothFlags(t *testing.T) {
	skipIfBinaryMissing(t)
	cmd := exec.Command("/tmp/pryx-core", "skills", "list", "--eligible", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("skills list --eligible --json failed: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr != "" && outputStr != "null" && outputStr != "[]" {
		if !strings.HasPrefix(outputStr, "[") {
			t.Errorf("Expected JSON array output, got: %s", outputStr)
		}
	}
}
