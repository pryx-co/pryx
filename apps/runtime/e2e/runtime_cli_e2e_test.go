//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func runtimeRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(cwd, ".."))
}

func repoRoot(t *testing.T) string {
	t.Helper()
	return filepath.Clean(filepath.Join(runtimeRoot(t), "..", ".."))
}

func runPryxCore(t *testing.T, home string, args ...string) (string, int) {
	t.Helper()

	bin := buildPryxCore(t)

	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(),
		"HOME="+home,
		"PRYX_DB_PATH="+filepath.Join(home, "pryx.db"),
		"PRYX_WORKSPACE_ROOT="+repoRoot(t),
		"PRYX_BUNDLED_SKILLS_DIR="+filepath.Join(runtimeRoot(t), "internal", "skills", "bundled"),
	)

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

func waitForFile(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for file: %s", path)
}

func TestCLI_SkillsListJSON_IncludesBundledSkills(t *testing.T) {
	home := t.TempDir()

	out, code := runPryxCore(t, home, "skills", "list", "--json")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}

	var skills []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(out), &skills); err != nil {
		t.Fatalf("expected json output, got:\n%s\nerror: %v", out, err)
	}

	got := map[string]bool{}
	for _, s := range skills {
		got[s.ID] = true
	}

	for _, id := range []string{"docker-manager", "git-tool", "cloud-deploy"} {
		if !got[id] {
			t.Fatalf("expected bundled skill %q in list, got ids: %v", id, keys(got))
		}
	}
}

func TestCLI_SkillsInfo_WorksForBundledSkill(t *testing.T) {
	home := t.TempDir()

	out, code := runPryxCore(t, home, "skills", "info", "docker-manager")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "Skill: docker-manager") {
		t.Fatalf("expected output to contain skill header, got:\n%s", out)
	}
}

func TestCLI_MCPConfig_RoundTrip(t *testing.T) {
	home := t.TempDir()

	out, code := runPryxCore(t, home, "mcp", "list", "--json")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}

	out, code = runPryxCore(t, home, "mcp", "add", "test-server", "--url", "https://example.com")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}

	out, code = runPryxCore(t, home, "mcp", "list", "--json")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}

	var servers map[string]any
	if strings.TrimSpace(out) != "" && strings.TrimSpace(out) != "null" {
		if err := json.Unmarshal([]byte(out), &servers); err != nil {
			t.Fatalf("expected json output, got:\n%s\nerror: %v", out, err)
		}
	}
	if servers == nil || servers["test-server"] == nil {
		t.Fatalf("expected test-server in config, got:\n%s", out)
	}

	out, code = runPryxCore(t, home, "mcp", "remove", "test-server")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}
}

func TestCLI_Config_SetThenGet(t *testing.T) {
	home := t.TempDir()

	out, code := runPryxCore(t, home, "config", "set", "listen_addr", ":12345")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}

	out, code = runPryxCore(t, home, "config", "get", "listen_addr")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}
	if strings.TrimSpace(out) != ":12345" {
		t.Fatalf("expected listen_addr to be updated, got: %q", strings.TrimSpace(out))
	}
}

func TestRuntime_HealthAndWebsocket(t *testing.T) {
	home := t.TempDir()
	bin := buildPryxCore(t)

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"HOME="+home,
		"PRYX_DB_PATH="+filepath.Join(home, "pryx.db"),
		"PRYX_WORKSPACE_ROOT="+repoRoot(t),
		"PRYX_BUNDLED_SKILLS_DIR="+filepath.Join(runtimeRoot(t), "internal", "skills", "bundled"),
		"PRYX_LISTEN_ADDR=:0",
	)

	if err := cmd.Start(); err != nil {
		t.Fatalf("start runtime: %v", err)
	}
	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()
	t.Cleanup(func() {
		_ = cmd.Process.Signal(os.Interrupt)
		select {
		case <-time.After(2 * time.Second):
			_ = cmd.Process.Kill()
			<-waitCh
		case <-waitCh:
		}
	})

	portFile := filepath.Join(home, ".pryx", "runtime.port")
	if err := waitForFile(portFile, 5*time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	portBytes, err := os.ReadFile(portFile)
	if err != nil {
		t.Fatalf("read port file: %v", err)
	}
	port := strings.TrimSpace(string(portBytes))
	if port == "" {
		t.Fatalf("empty port file")
	}

	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", port))
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "OK" {
		t.Fatalf("unexpected /health response: %d %q", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	wsURL := fmt.Sprintf("ws://localhost:%s/ws?surface=e2e&event=trace.event", port)
	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	_ = c.Close(websocket.StatusNormalClosure, "")
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
