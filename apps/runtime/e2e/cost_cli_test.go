package e2e

import (
	"strings"
	"testing"
)

func TestCostCLI_Summary(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "cost", "summary")
	if code != 0 {
		t.Fatalf("cost summary failed (code %d):\n%s", code, out)
	}

	if !strings.Contains(out, "Total Cost") {
		t.Errorf("Expected 'Total Cost' in output, got: %s", out)
	}

	if !strings.Contains(out, "Total Tokens") {
		t.Errorf("Expected 'Total Tokens' in output, got: %s", out)
	}
}

func TestCostCLI_Pricing(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "cost", "pricing")
	if code != 0 {
		t.Fatalf("cost pricing failed (code %d):\n%s", code, out)
	}

	if !strings.Contains(out, "Model") {
		t.Errorf("Expected 'Model' in output, got: %s", out)
	}

	if !strings.Contains(out, "Provider") {
		t.Errorf("Expected 'Provider' in output, got: %s", out)
	}

	if !strings.Contains(out, "Input") {
		t.Errorf("Expected 'Input' in output, got: %s", out)
	}
}

func TestCostCLI_Daily(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "cost", "daily")
	if code != 0 {
		t.Fatalf("cost daily failed (code %d):\n%s", code, out)
	}

	if strings.TrimSpace(out) == "" {
		t.Error("Expected daily breakdown output")
	}
}

func TestCostCLI_Monthly(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "cost", "monthly")
	if code != 0 {
		t.Fatalf("cost monthly failed (code %d):\n%s", code, out)
	}

	if strings.TrimSpace(out) == "" {
		t.Error("Expected monthly breakdown output")
	}
}

func TestCostCLI_BudgetStatus(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "cost", "budget")
	if code != 0 {
		t.Fatalf("cost budget status failed (code %d):\n%s", code, out)
	}

	if !strings.Contains(out, "Budget Status") {
		t.Errorf("Expected 'Budget Status' in output, got: %s", out)
	}
}

func TestCostCLI_BudgetSet(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "cost", "budget", "set", "--daily", "10.00", "--monthly", "100.00")
	if code != 0 {
		t.Fatalf("cost budget set failed (code %d):\n%s", code, out)
	}
	if !strings.Contains(out, "Budget set") {
		t.Fatalf("expected budget set confirmation, got:\n%s", out)
	}
}

func TestCostCLI_Optimize(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "cost", "optimize")
	if code != 0 {
		t.Fatalf("cost optimize failed (code %d):\n%s", code, out)
	}

	if strings.TrimSpace(out) == "" {
		t.Error("Expected optimize output")
	}
}

func TestCostCLI_InvalidCommand(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "cost", "invalid")
	if code == 0 {
		t.Error("Expected error for invalid cost command")
	}

	if !strings.Contains(out, "Unknown") {
		t.Logf("Output: %s", out)
	}
}

func TestCostCLI_NoSubcommand(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "cost")
	if code == 0 {
		t.Error("Expected error for missing subcommand")
	}

	if !strings.Contains(out, "Usage") && !strings.Contains(out, "Commands") {
		t.Logf("Output: %s", out)
	}
}
