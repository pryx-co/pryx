package doctor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/mcp"
	"pryx-core/internal/policy"
	"pryx-core/internal/store"
)

type Status string

const (
	StatusOK   Status = "ok"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

type Check struct {
	Name       string `json:"name"`
	Status     Status `json:"status"`
	Detail     string `json:"detail,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

type Report struct {
	Checks []Check `json:"checks"`
}

func (r *Report) Add(c Check) {
	r.Checks = append(r.Checks, c)
}

func Run(ctx context.Context, cfg *config.Config, kc *keychain.Keychain) (Report, int) {
	rep := Report{}

	rep.Add(checkInstallation())
	rep.Add(checkDependencies())
	rep.Add(checkRuntimeHealth(ctx, cfg))

	dbCheck, dbConn := checkDatabase(cfg)
	rep.Add(dbCheck)
	if dbConn != nil {
		defer dbConn.Close()
	}

	rep.Add(checkMCP(ctx, kc))
	rep.Add(checkChannels())

	exitCode := 0
	for _, c := range rep.Checks {
		if c.Status == StatusFail {
			exitCode = 1
			break
		}
	}
	return rep, exitCode
}

func checkInstallation() Check {
	exe, err := os.Executable()
	if err != nil || strings.TrimSpace(exe) == "" {
		return Check{Name: "installation", Status: StatusFail, Detail: "cannot resolve executable path", Suggestion: "reinstall pryx-core"}
	}
	if _, err := os.Stat(exe); err != nil {
		return Check{Name: "installation", Status: StatusFail, Detail: err.Error(), Suggestion: "reinstall pryx-core"}
	}
	return Check{Name: "installation", Status: StatusOK, Detail: exe}
}

func checkDependencies() Check {
	if runtime.GOOS != "windows" {
		if _, err := exec.LookPath("sh"); err != nil {
			return Check{Name: "dependencies", Status: StatusWarn, Detail: "sh not found", Suggestion: "ensure a POSIX shell is installed"}
		}
	}
	return Check{Name: "dependencies", Status: StatusOK}
}

func checkRuntimeHealth(ctx context.Context, cfg *config.Config) Check {
	addr := strings.TrimSpace(cfg.ListenAddr)
	if addr == "" {
		return Check{Name: "pryx-core health", Status: StatusWarn, Detail: "missing listen addr", Suggestion: "set PRYX_LISTEN_ADDR"}
	}

	url := healthURL(addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Check{Name: "pryx-core health", Status: StatusWarn, Detail: err.Error(), Suggestion: "start pryx-core and retry"}
	}
	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return Check{Name: "pryx-core health", Status: StatusWarn, Detail: err.Error(), Suggestion: "start pryx-core and retry"}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Check{Name: "pryx-core health", Status: StatusWarn, Detail: fmt.Sprintf("status %d", resp.StatusCode), Suggestion: "check runtime logs"}
	}
	return Check{Name: "pryx-core health", Status: StatusOK, Detail: url}
}

func checkDatabase(cfg *config.Config) (Check, *sql.DB) {
	path := strings.TrimSpace(cfg.DatabasePath)
	if path == "" {
		return Check{Name: "sqlite", Status: StatusFail, Detail: "missing database path", Suggestion: "set PRYX_DB_PATH"}, nil
	}
	s, err := store.New(path)
	if err != nil {
		return Check{Name: "sqlite", Status: StatusFail, Detail: err.Error(), Suggestion: "check file permissions or PRYX_DB_PATH"}, nil
	}
	return Check{Name: "sqlite", Status: StatusOK, Detail: filepath.Clean(path)}, s.DB
}

func checkMCP(ctx context.Context, kc *keychain.Keychain) Check {
	p := policy.NewEngine(nil)
	mgr := mcp.NewManager(nil, p, kc)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := mgr.LoadAndConnect(ctx)
	if err != nil {
		return Check{Name: "mcp", Status: StatusFail, Detail: err.Error(), Suggestion: "check ~/.pryx/mcp/servers.json or remove it to use bundled defaults"}
	}
	tools, err := mgr.ListToolsFlat(ctx, false)
	if err != nil {
		return Check{Name: "mcp", Status: StatusFail, Detail: err.Error(), Suggestion: "check MCP server configuration"}
	}
	if len(tools) == 0 {
		return Check{Name: "mcp", Status: StatusWarn, Detail: "no tools returned", Suggestion: "check MCP server configuration"}
	}
	return Check{Name: "mcp", Status: StatusOK, Detail: fmt.Sprintf("%d tools", len(tools))}
}

func checkChannels() Check {
	pryxDir := filepath.Dir(config.DefaultPath())
	path := filepath.Join(pryxDir, "channels.json")
	if _, err := os.Stat(path); err == nil {
		return Check{Name: "channels", Status: StatusOK, Detail: path}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Check{Name: "channels", Status: StatusWarn, Detail: err.Error(), Suggestion: "check channel config permissions"}
	}
	return Check{Name: "channels", Status: StatusWarn, Detail: "no channel configuration found", Suggestion: "create .pryx/channels.json to enable channels"}
}

func healthURL(listenAddr string) string {
	addr := strings.TrimSpace(listenAddr)
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return strings.TrimRight(addr, "/") + "/health"
	}
	if strings.HasPrefix(addr, ":") {
		return "http://127.0.0.1" + addr + "/health"
	}
	return "http://" + addr + "/health"
}
