//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"pryx-core/internal/config"
)

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

func writeModelsCache(t *testing.T, home string, providers map[string]map[string]any) {
	t.Helper()

	cacheFile := filepath.Join(home, ".pryx", "cache", "models.json")
	if err := os.MkdirAll(filepath.Dir(cacheFile), 0o755); err != nil {
		t.Fatalf("mkdir models cache dir: %v", err)
	}

	now := time.Now().Format(time.RFC3339Nano)
	payload := map[string]any{
		"models":     map[string]any{},
		"providers":  providers,
		"fetched_at": now,
		"cached_at":  now,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal models cache: %v", err)
	}
	if err := os.WriteFile(cacheFile, data, 0o644); err != nil {
		t.Fatalf("write models cache: %v", err)
	}
}

func readKeychainFile(t *testing.T, path string) map[string]string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read keychain file: %v", err)
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse keychain file: %v", err)
	}
	return m
}

func runPryxCoreWithEnvInput(t *testing.T, home string, extraEnv map[string]string, input string, args ...string) (string, int) {
	t.Helper()

	bin := buildPryxCore(t)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = strings.NewReader(input)
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

func TestCLI_SkillsListJSON_IncludesBundledSkills(t *testing.T) {
	home := t.TempDir()
	bundled := filepath.Join(home, "bundled-skills")
	for _, id := range []string{"docker-manager", "git-tool", "cloud-deploy"} {
		writeSkill(t, bundled, id, true)
	}

	out, code := runPryxCoreWithEnv(t, home, map[string]string{
		"PRYX_BUNDLED_SKILLS_DIR": bundled,
		"PRYX_MANAGED_SKILLS_DIR": filepath.Join(home, "managed-skills"),
		"PRYX_WORKSPACE_ROOT":     filepath.Join(home, "workspace"),
	}, "skills", "list", "--json")
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
	bundled := filepath.Join(home, "bundled-skills")
	writeSkill(t, bundled, "docker-manager", true)

	out, code := runPryxCoreWithEnv(t, home, map[string]string{
		"PRYX_BUNDLED_SKILLS_DIR": bundled,
		"PRYX_MANAGED_SKILLS_DIR": filepath.Join(home, "managed-skills"),
		"PRYX_WORKSPACE_ROOT":     filepath.Join(home, "workspace"),
	}, "skills", "info", "docker-manager")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "Skill: docker-manager") {
		t.Fatalf("expected output to contain skill header, got:\n%s", out)
	}
}

func TestCLI_MCPConfig_RoundTrip(t *testing.T) {
	home := t.TempDir()

	out, code := runPryxCoreWithEnv(t, home, nil, "mcp", "list", "--json")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}

	out, code = runPryxCoreWithEnv(t, home, nil, "mcp", "add", "test-server", "--url", "https://example.com")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}

	out, code = runPryxCoreWithEnv(t, home, nil, "mcp", "list", "--json")
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

	out, code = runPryxCoreWithEnv(t, home, nil, "mcp", "remove", "test-server")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}
}

func TestCLI_Config_SetThenGet(t *testing.T) {
	home := t.TempDir()

	out, code := runPryxCoreWithEnv(t, home, nil, "config", "set", "listen_addr", ":12345")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}

	out, code = runPryxCoreWithEnv(t, home, nil, "config", "get", "listen_addr")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}
	if strings.TrimSpace(out) != ":12345" {
		t.Fatalf("expected listen_addr to be updated, got: %q", strings.TrimSpace(out))
	}
}

func TestCLI_Login_Success_WithPKCE(t *testing.T) {
	home := t.TempDir()
	keychainPath := filepath.Join(home, ".pryx", "keychain.json")

	deviceCode := "device-123"

	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device/code":
			var req map[string]any
			_ = json.NewDecoder(r.Body).Decode(&req)

			challenge, _ := req["code_challenge"].(string)
			method, _ := req["code_challenge_method"].(string)
			if challenge == "" || method != "S256" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "missing pkce params"})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":      deviceCode,
				"user_code":        "USER-CODE",
				"verification_uri": "https://example.com/verify",
				"expires_in":       600,
				"interval":         1,
			})
		case "/auth/device/token":
			var req map[string]string
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req["device_code"] != deviceCode || strings.TrimSpace(req["code_verifier"]) == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid token request"})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "token-123",
				"expires_in":   3600,
				"token_type":   "bearer",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer cloud.Close()

	out, code := runPryxCoreWithEnv(t, home, map[string]string{
		"PRYX_CLOUD_API_URL": cloud.URL,
		"PRYX_KEYCHAIN_FILE": keychainPath,
	}, "login")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "Verification URL:") || !strings.Contains(out, "User Code:") {
		t.Fatalf("expected login prompt output, got:\n%s", out)
	}
	if !strings.Contains(out, "Successfully logged in") {
		t.Fatalf("expected success output, got:\n%s", out)
	}

	kc := readKeychainFile(t, keychainPath)
	if got := kc["pryx:cloud_access_token"]; got != "token-123" {
		t.Fatalf("expected keychain to store access token, got %q", got)
	}
}

func TestCLI_Provider_AddUseRemove_UsesKeychain(t *testing.T) {
	home := t.TempDir()
	keychainPath := filepath.Join(home, ".pryx", "keychain.json")
	t.Setenv("HOME", home)
	cfg := config.Load()
	if err := cfg.Save(filepath.Join(home, ".pryx", "config.yaml")); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}
	writeModelsCache(t, home, map[string]map[string]any{
		"openai": {
			"name": "OpenAI",
			"env":  []string{"OPENAI_API_KEY"},
			"doc":  "https://example.com/openai",
		},
		"ollama": {
			"name": "Ollama",
			"env":  []string{},
			"doc":  "https://example.com/ollama",
		},
	})

	extraEnv := map[string]string{
		"PRYX_KEYCHAIN_FILE": keychainPath,
	}

	out, code := runPryxCoreWithEnvInput(t, home, extraEnv, "sk-test\n", "provider", "add", "openai")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}

	kc := readKeychainFile(t, keychainPath)
	if got := kc["pryx:provider:openai"]; got != "sk-test" {
		t.Fatalf("expected keychain to store provider key, got %q", got)
	}

	out, code = runPryxCoreWithEnv(t, home, extraEnv, "provider", "use", "openai")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}

	cfgPath := filepath.Join(home, ".pryx", "config.yaml")
	cfgBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config.yaml: %v", err)
	}
	cfgText := string(cfgBytes)
	if strings.Contains(cfgText, "sk-test") {
		t.Fatalf("expected config.yaml to not contain API key, got:\n%s", cfgText)
	}
	if !strings.Contains(cfgText, "model_provider: openai") {
		t.Fatalf("expected config.yaml to set model_provider, got:\n%s", cfgText)
	}

	out, code = runPryxCoreWithEnv(t, home, extraEnv, "provider", "list")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "OpenAI") || !strings.Contains(strings.ToLower(out), "keychain") {
		t.Fatalf("expected provider list output to mention OpenAI and keychain, got:\n%s", out)
	}

	out, code = runPryxCoreWithEnv(t, home, extraEnv, "provider", "remove", "openai")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\n%s", code, out)
	}

	kc = readKeychainFile(t, keychainPath)
	if _, ok := kc["pryx:provider:openai"]; ok {
		t.Fatalf("expected provider key to be removed from keychain")
	}
}

func TestCLI_Provider_UseFailsWhenNotConfigured(t *testing.T) {
	home := t.TempDir()
	keychainPath := filepath.Join(home, ".pryx", "keychain.json")
	writeModelsCache(t, home, map[string]map[string]any{
		"openai": {
			"name": "OpenAI",
			"env":  []string{"OPENAI_API_KEY"},
			"doc":  "https://example.com/openai",
		},
	})

	out, code := runPryxCoreWithEnv(t, home, map[string]string{
		"PRYX_KEYCHAIN_FILE": keychainPath,
	}, "provider", "use", "openai")
	if code == 0 {
		t.Fatalf("expected non-zero exit code for unconfigured provider, got 0\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "not configured") {
		t.Fatalf("expected output to mention not configured, got:\n%s", out)
	}
}

func TestCLI_Provider_SetKeyRejectsEmptyKey(t *testing.T) {
	home := t.TempDir()
	keychainPath := filepath.Join(home, ".pryx", "keychain.json")
	writeModelsCache(t, home, map[string]map[string]any{
		"openai": {
			"name": "OpenAI",
			"env":  []string{"OPENAI_API_KEY"},
			"doc":  "https://example.com/openai",
		},
	})

	out, code := runPryxCoreWithEnvInput(t, home, map[string]string{
		"PRYX_KEYCHAIN_FILE": keychainPath,
	}, "\n", "provider", "set-key", "openai")
	if code == 0 {
		t.Fatalf("expected non-zero exit code for empty key, got 0\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "cannot be empty") {
		t.Fatalf("expected output to mention empty key, got:\n%s", out)
	}
}

func TestRuntime_HealthAndWebsocket(t *testing.T) {
	home := t.TempDir()
	bin := buildPryxCore(t)
	bundled := filepath.Join(home, "bundled-skills")
	writeSkill(t, bundled, "alpha-skill", true)

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"HOME="+home,
		"PRYX_DB_PATH="+filepath.Join(home, "pryx.db"),
		"PRYX_WORKSPACE_ROOT="+filepath.Join(home, "workspace"),
		"PRYX_BUNDLED_SKILLS_DIR="+bundled,
		"PRYX_KEYCHAIN_FILE="+filepath.Join(home, ".pryx", "keychain.json"),
		"PRYX_TELEMETRY_DISABLED=true",
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
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected /health response: %d %q", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("expected json /health output, got:\n%s\nerror: %v", strings.TrimSpace(string(body)), err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("expected status ok, got: %v (body: %s)", payload["status"], strings.TrimSpace(string(body)))
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

func runtimeRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for i := 0; i < 8; i++ {
		if filepath.Base(dir) == "runtime" && filepath.Base(filepath.Dir(dir)) == "apps" {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}

	t.Fatalf("could not locate apps/runtime from cwd")
	return ""
}

func repoRoot(t *testing.T) string {
	t.Helper()

	runtime := runtimeRoot(t)
	apps := filepath.Dir(runtime)
	return filepath.Dir(apps)
}

func startPryxCore(t *testing.T, bin string, home string) (port string, cancel context.CancelFunc) {
	t.Helper()

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"HOME="+home,
		"PRYX_DB_PATH="+filepath.Join(home, "pryx.db"),
		"PRYX_LISTEN_ADDR=:0",
		"PRYX_WORKSPACE_ROOT="+repoRoot(t),
		"PRYX_BUNDLED_SKILLS_DIR="+filepath.Join(runtimeRoot(t), "internal", "skills", "bundled"),
		"PRYX_KEYCHAIN_FILE="+filepath.Join(home, ".pryx", "keychain.json"),
		"PRYX_TELEMETRY_DISABLED=true",
	)

	if err := cmd.Start(); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-ctx.Done()
		_ = cmd.Process.Signal(os.Interrupt)
		select {
		case <-time.After(2 * time.Second):
			_ = cmd.Process.Kill()
		case <-func() chan struct{} {
			ch := make(chan struct{})
			go func() {
				_ = cmd.Wait()
				close(ch)
			}()
			return ch
		}():
		}
	}()

	portFile := filepath.Join(home, ".pryx", "runtime.port")
	if err := waitForFile(portFile, 5*time.Second); err != nil {
		cancel()
		t.Fatalf("%v", err)
	}

	portBytes, err := os.ReadFile(portFile)
	if err != nil {
		cancel()
		t.Fatalf("read port file: %v", err)
	}
	port = strings.TrimSpace(string(portBytes))
	if port == "" {
		cancel()
		t.Fatalf("empty port file")
	}

	return port, cancel
}

func waitForServer(t *testing.T, port string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://localhost:" + port + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for server to be ready")
}
