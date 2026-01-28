package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type ShellProvider struct {
	root string
}

func NewShellProvider() *ShellProvider {
	root := strings.TrimSpace(os.Getenv("PRYX_WORKSPACE_ROOT"))
	if root == "" {
		if cwd, err := os.Getwd(); err == nil {
			root = cwd
		}
	}
	if root != "" {
		if abs, err := filepath.Abs(root); err == nil {
			root = abs
		}
	}
	return &ShellProvider{root: root}
}

func (p *ShellProvider) ServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":    "pryx-core/shell",
		"title":   "Pryx Shell (Bundled)",
		"version": "dev",
	}
}

func (p *ShellProvider) ListTools(ctx context.Context) ([]Tool, error) {
	_ = ctx
	return []Tool{
		{Name: "exec", Title: "Execute Command", InputSchema: schemaRaw(`{"type":"object","properties":{"command":{"type":"string"},"args":{"type":"array","items":{"type":"string"}},"cwd":{"type":"string"},"timeout_ms":{"type":"integer"},"env":{"type":"object","additionalProperties":{"type":"string"}}},"additionalProperties":false}`)},
	}, nil
}

func (p *ShellProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (ToolResult, error) {
	switch name {
	case "exec":
		return p.exec(ctx, arguments)
	default:
		return ToolResult{}, errors.New("unknown tool")
	}
}

func (p *ShellProvider) exec(ctx context.Context, arguments map[string]interface{}) (ToolResult, error) {
	timeoutMS := argInt(arguments, "timeout_ms")
	if timeoutMS <= 0 {
		timeoutMS = 30_000
	}
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()

	cwd := strings.TrimSpace(argString(arguments, "cwd"))
	if cwd != "" {
		var err error
		cwd, err = p.resolveCwd(cwd)
		if err != nil {
			return ToolResult{}, err
		}
	} else {
		cwd = p.root
	}

	env := map[string]string{}
	if raw, ok := arguments["env"]; ok {
		if m, ok := raw.(map[string]interface{}); ok {
			for k, v := range m {
				env[k] = argString(map[string]interface{}{"v": v}, "v")
			}
		}
	}

	var cmd *exec.Cmd
	args := argStringSlice(arguments, "args")
	command := strings.TrimSpace(argString(arguments, "command"))

	if len(args) > 0 && command == "" {
		command = args[0]
		args = args[1:]
	}
	if command == "" {
		return ToolResult{}, errors.New("missing command")
	}

	if len(args) == 0 && strings.Contains(command, " ") {
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(execCtx, "cmd.exe", "/C", command)
		} else {
			cmd = exec.CommandContext(execCtx, "sh", "-lc", command)
		}
	} else {
		cmd = exec.CommandContext(execCtx, command, args...)
	}
	cmd.Dir = cwd

	if len(env) > 0 {
		cmd.Env = append(os.Environ(), flattenEnv(env)...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
		}
	}

	out := map[string]interface{}{
		"cwd":       cwd,
		"command":   command,
		"args":      args,
		"exit_code": exitCode,
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
	}

	if err != nil && exitCode == -1 {
		out["error"] = err.Error()
	}

	return ToolResult{
		Content:           []ToolContent{{Type: "text", Text: "OK"}},
		StructuredContent: jsonRaw(out),
	}, nil
}

func (p *ShellProvider) resolveCwd(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return p.root, nil
	}
	if strings.HasPrefix(raw, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			raw = filepath.Join(home, strings.TrimPrefix(raw, "~"))
		}
	}
	abs := raw
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(p.root, abs)
	}
	abs = filepath.Clean(abs)
	abs, err := filepath.Abs(abs)
	if err != nil {
		return "", err
	}
	root := p.root
	if root == "" {
		return abs, nil
	}
	root = filepath.Clean(root)
	if !strings.HasPrefix(abs, root+string(filepath.Separator)) && abs != root {
		return "", errors.New("cwd escapes workspace root")
	}
	return abs, nil
}

func argInt(args map[string]interface{}, key string, def ...int) int {
	v, ok := args[key]
	if !ok || v == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case string:
		t = strings.TrimSpace(t)
		if t == "" {
			if len(def) > 0 {
				return def[0]
			}
			return 0
		}
		var n int
		_ = json.Unmarshal([]byte(t), &n)
		return n
	default:
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
}

func argStringSlice(args map[string]interface{}, key string) []string {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	if arr, ok := v.([]interface{}); ok {
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			} else {
				b, _ := json.Marshal(item)
				out = append(out, strings.Trim(string(b), `"`))
			}
		}
		return out
	}
	if arr, ok := v.([]string); ok {
		return arr
	}
	return nil
}

func flattenEnv(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		if strings.TrimSpace(k) == "" {
			continue
		}
		out = append(out, k+"="+v)
	}
	return out
}
