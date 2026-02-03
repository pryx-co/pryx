package e2e

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	pryxCoreBuildOnce sync.Once
	pryxCoreBuildErr  error
	pryxCorePath      string
)

func buildPryxCore(t *testing.T) string {
	t.Helper()

	pryxCoreBuildOnce.Do(func() {
		cwd, err := os.Getwd()
		if err != nil {
			pryxCoreBuildErr = err
			return
		}

		runtimeRoot := filepath.Clean(filepath.Join(cwd, ".."))
		outPath := filepath.Join(os.TempDir(), fmt.Sprintf("pryx-core-test-%d", time.Now().UnixNano()))

		cmd := exec.Command("go", "build", "-o", outPath, "./cmd/pryx-core")
		cmd.Dir = runtimeRoot
		cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			pryxCoreBuildErr = fmt.Errorf("build pryx-core failed: %w\n%s", err, string(output))
			return
		}

		pryxCorePath = outPath
	})

	if pryxCoreBuildErr != nil {
		t.Fatalf("%v", pryxCoreBuildErr)
	}
	return pryxCorePath
}

func makeEnv(home string, extraEnv map[string]string) []string {
	skip := map[string]struct{}{
		"HOME":         {},
		"PRYX_DB_PATH": {},
	}
	for k := range extraEnv {
		skip[k] = struct{}{}
	}

	var env []string
	for _, kv := range os.Environ() {
		key := strings.SplitN(kv, "=", 2)[0]
		if _, ok := skip[key]; ok {
			continue
		}
		env = append(env, kv)
	}

	env = append(env,
		"HOME="+home,
		"PRYX_DB_PATH="+filepath.Join(home, "pryx.db"),
	)

	keys := make([]string, 0, len(extraEnv))
	for k := range extraEnv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		env = append(env, k+"="+extraEnv[k])
	}

	return env
}

func runPryxCoreWithEnv(t *testing.T, home string, extraEnv map[string]string, args ...string) (string, int) {
	t.Helper()

	bin := buildPryxCore(t)
	cmd := exec.Command(bin, args...)
	if extraEnv == nil {
		extraEnv = map[string]string{}
	}
	cmd.Env = makeEnv(home, extraEnv)

	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(out), exitErr.ExitCode()
	}
	t.Fatalf("run pryx-core failed: %v\n%s", err, string(out))
	return "", 1
}

func writeSkill(t *testing.T, skillsRoot string, id string, enabled bool) string {
	t.Helper()

	dir := filepath.Join(skillsRoot, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	state := "enabled"
	if !enabled {
		state = "disabled"
	}
	skillPath := filepath.Join(dir, "SKILL.md")
	body := fmt.Sprintf(`---
name: %s
description: test skill %s
enabled: %v
metadata:
  pryx:
    emoji: "🔧"
---
# %s

This is a %s test skill.
`, id, id, enabled, id, state)
	if err := os.WriteFile(skillPath, []byte(body), 0644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}
	return skillPath
}

func TestSkillsCLI_List(t *testing.T) {
	home := t.TempDir()
	bundled := filepath.Join(home, "bundled-skills")
	writeSkill(t, bundled, "alpha-skill", true)

	out, code := runPryxCoreWithEnv(t, home, map[string]string{
		"PRYX_BUNDLED_SKILLS_DIR": bundled,
		"PRYX_MANAGED_SKILLS_DIR": filepath.Join(home, "managed-skills"),
		"PRYX_WORKSPACE_ROOT":     filepath.Join(home, "workspace"),
	}, "skills", "list")
	if code != 0 {
		t.Fatalf("skills list failed (code %d):\n%s", code, out)
	}
	if !strings.Contains(out, "Available Skills") {
		t.Fatalf("expected 'Available Skills' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "alpha-skill") {
		t.Fatalf("expected skill id in output, got:\n%s", out)
	}
}

func TestSkillsCLI_Help(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "skills")

	// Allow exit code 0 or 2 (some CLIs use 2 for help)
	if code != 0 && code != 2 {
		t.Fatalf("skills help failed with unexpected exit code: %d\n%s", code, out)
	}

	// Should show commands
	expectedCommands := []string{"list", "info", "check", "enable", "disable", "install", "uninstall"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(out, cmd) {
			t.Errorf("Expected '%s' in help output", cmd)
		}
	}
}

func TestSkillsCLI_ListJSON(t *testing.T) {
	home := t.TempDir()
	bundled := filepath.Join(home, "bundled-skills")
	writeSkill(t, bundled, "alpha-skill", true)

	out, code := runPryxCoreWithEnv(t, home, map[string]string{
		"PRYX_BUNDLED_SKILLS_DIR": bundled,
		"PRYX_MANAGED_SKILLS_DIR": filepath.Join(home, "managed-skills"),
		"PRYX_WORKSPACE_ROOT":     filepath.Join(home, "workspace"),
	}, "skills", "list", "--json")
	if code != 0 {
		t.Fatalf("skills list --json failed (code %d):\n%s", code, out)
	}
	outStr := strings.TrimSpace(out)
	if outStr == "" || outStr == "null" {
		t.Fatalf("expected json array output, got: %q", outStr)
	}
	if !strings.HasPrefix(outStr, "[") {
		t.Fatalf("expected json array output, got:\n%s", outStr)
	}
}

func TestSkillsCLI_Check(t *testing.T) {
	home := t.TempDir()
	bundled := filepath.Join(home, "bundled-skills")
	writeSkill(t, bundled, "alpha-skill", true)

	out, _ := runPryxCoreWithEnv(t, home, map[string]string{
		"PRYX_BUNDLED_SKILLS_DIR": bundled,
		"PRYX_MANAGED_SKILLS_DIR": filepath.Join(home, "managed-skills"),
		"PRYX_WORKSPACE_ROOT":     filepath.Join(home, "workspace"),
	}, "skills", "check")
	if strings.TrimSpace(out) == "" {
		t.Fatalf("expected check output, got empty")
	}
}

func TestSkillsCLI_InfoError(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, map[string]string{
		"PRYX_BUNDLED_SKILLS_DIR": filepath.Join(home, "bundled-skills"),
		"PRYX_MANAGED_SKILLS_DIR": filepath.Join(home, "managed-skills"),
		"PRYX_WORKSPACE_ROOT":     filepath.Join(home, "workspace"),
	}, "skills", "info", "nonexistent")
	if code == 0 {
		t.Error("Expected error for non-existent skill")
	}

	if !strings.Contains(strings.ToLower(out), "not found") {
		t.Errorf("Expected 'not found' in error output, got: %s", out)
	}
}

func TestSkillsCLI_UnknownCommand(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "skills", "unknown")
	if code == 0 {
		t.Error("Expected error for unknown command")
	}

	if !strings.Contains(out, "Unknown command") {
		t.Errorf("Expected 'Unknown command' in error output, got: %s", out)
	}
}

func TestMCPCLI_List(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "mcp", "list")
	if code != 0 {
		t.Fatalf("mcp list failed (code %d):\n%s", code, out)
	}
	if !strings.Contains(out, "MCP Servers") {
		t.Errorf("Expected 'MCP Servers' in output, got: %s", out)
	}
}

func TestMCPCLI_Help(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "mcp")

	// Allow exit code 0 or 2 (some CLIs use 2 for help)
	if code != 0 && code != 2 {
		t.Fatalf("mcp help failed with unexpected exit code: %d\n%s", code, out)
	}

	// Should show commands
	expectedCommands := []string{"list", "add", "remove", "test", "auth"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(out, cmd) {
			t.Errorf("Expected '%s' in help output", cmd)
		}
	}
}

func TestMCPCLI_AddRemove(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "mcp", "add", "test-server", "--url", "https://example.com")
	if code != 0 {
		t.Fatalf("mcp add failed (code %d):\n%s", code, out)
	}

	out, code = runPryxCoreWithEnv(t, home, nil, "mcp", "remove", "test-server")
	if code != 0 {
		t.Fatalf("mcp remove failed (code %d):\n%s", code, out)
	}
}

func TestMCPCLI_ListJSON(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "mcp", "list", "--json")
	if code != 0 {
		t.Fatalf("mcp list --json failed (code %d):\n%s", code, out)
	}
	outStr := strings.TrimSpace(out)
	if outStr == "" || outStr == "null" {
		t.Fatalf("expected json object output, got: %q", outStr)
	}
	if !strings.HasPrefix(outStr, "{") {
		t.Fatalf("expected json object output, got:\n%s", outStr)
	}
}

func TestMCPCLI_TestError(t *testing.T) {
	home := t.TempDir()
	out, code := runPryxCoreWithEnv(t, home, nil, "mcp", "test", "nonexistent")
	if code == 0 {
		t.Error("Expected error for non-existent server")
	}

	if !strings.Contains(strings.ToLower(out), "not found") {
		t.Errorf("Expected 'not found' in error output, got: %s", out)
	}
}
