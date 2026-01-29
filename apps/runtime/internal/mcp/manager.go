package mcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/hostrpc"
	"pryx-core/internal/keychain"
	"pryx-core/internal/policy"
)

type Manager struct {
	bus      *bus.Bus
	policy   *policy.Engine
	keychain *keychain.Keychain

	mu      sync.RWMutex
	clients map[string]*Client

	cacheMu sync.RWMutex
	cache   map[string]cachedTools

	approvalMu       sync.Mutex
	pendingApprovals map[string]pendingApproval
}

type cachedTools struct {
	fetchedAt time.Time
	tools     []Tool
}

type pendingApproval struct {
	ch        chan bool
	sessionID string
	tool      string
	reason    string
	args      map[string]interface{}
}

func NewManager(b *bus.Bus, p *policy.Engine, kc *keychain.Keychain) *Manager {
	if p == nil {
		p = policy.NewEngine(nil)
	}
	return &Manager{
		bus:              b,
		policy:           p,
		keychain:         kc,
		clients:          map[string]*Client{},
		cache:            map[string]cachedTools{},
		pendingApprovals: map[string]pendingApproval{},
	}
}

func (m *Manager) ResolveApproval(approvalID string, approved bool) bool {
	m.approvalMu.Lock()
	pa, ok := m.pendingApprovals[approvalID]
	if ok {
		delete(m.pendingApprovals, approvalID)
	}
	m.approvalMu.Unlock()

	if !ok {
		return false
	}

	select {
	case pa.ch <- approved:
	default:
	}

	if m.bus != nil {
		m.bus.Publish(bus.NewEvent(bus.EventApprovalResolved, pa.sessionID, map[string]interface{}{
			"approval_id": approvalID,
			"tool":        pa.tool,
			"approved":    approved,
		}))
	}

	return true
}

func (m *Manager) LoadAndConnect(ctx context.Context) (string, error) {
	cfg, path, err := LoadServersConfigFromFirstExisting(DefaultServersConfigPaths())
	if err != nil {
		return path, err
	}

	if path == "" && len(cfg.Servers) == 0 {
		cfg.Servers = map[string]ServerConfig{
			"filesystem": {Transport: "bundled"},
			"shell":      {Transport: "bundled"},
			"browser":    {Transport: "bundled"},
			"clipboard":  {Transport: "bundled"},
		}
	}

	clients := map[string]*Client{}
	for name, sc := range cfg.Servers {
		client, err := m.buildClient(name, sc)
		if err != nil {
			return path, fmt.Errorf("%s: %w", name, err)
		}
		clients[name] = client
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(clients))
	for name, c := range clients {
		wg.Add(1)
		go func(name string, c *Client) {
			defer wg.Done()
			if err := c.Initialize(ctx); err != nil {
				errs <- fmt.Errorf("%s: %w", name, err)
			}
		}(name, c)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		return path, err
	}

	m.mu.Lock()
	m.clients = clients
	m.mu.Unlock()

	return path, nil
}

func (m *Manager) ListTools(ctx context.Context, refresh bool) (map[string][]Tool, error) {
	m.mu.RLock()
	clients := make(map[string]*Client, len(m.clients))
	for k, v := range m.clients {
		clients[k] = v
	}
	m.mu.RUnlock()

	out := make(map[string][]Tool, len(clients))
	for name, c := range clients {
		tools, err := m.listToolsCached(ctx, name, c, refresh)
		if err != nil {
			return nil, err
		}
		out[name] = tools
	}
	return out, nil
}

func (m *Manager) ListToolsFlat(ctx context.Context, refresh bool) ([]Tool, error) {
	perServer, err := m.ListTools(ctx, refresh)
	if err != nil {
		return nil, err
	}

	var all []Tool
	for server, tools := range perServer {
		for _, t := range tools {
			all = append(all, Tool{
				Name:         fmt.Sprintf("%s:%s", server, t.Name),
				Title:        t.Title,
				Description:  t.Description,
				InputSchema:  t.InputSchema,
				OutputSchema: t.OutputSchema,
			})
		}
	}
	return all, nil
}

func (m *Manager) CallTool(ctx context.Context, sessionID string, toolName string, args map[string]interface{}) (ToolResult, error) {
	server, name := splitToolName(toolName)
	if server == "" || name == "" {
		return ToolResult{}, errors.New("invalid tool name")
	}

	m.mu.RLock()
	client := m.clients[server]
	m.mu.RUnlock()
	if client == nil {
		return ToolResult{}, fmt.Errorf("unknown mcp server: %s", server)
	}

	fullName := fmt.Sprintf("mcp.%s.%s", server, name)
	decision := m.policy.Evaluate(fullName, args)
	if m.bus != nil {
		m.bus.Publish(bus.NewEvent(bus.EventToolRequest, sessionID, map[string]interface{}{
			"tool":     fullName,
			"args":     args,
			"decision": decision,
		}))
	}

	switch decision.Decision {
	case policy.DecisionAllow:
	case policy.DecisionAsk:
		if strings.TrimSpace(os.Getenv("PRYX_HOST_RPC")) == "1" {
			client := hostrpc.NewDefaultClient()
			approved, err := client.RequestPermission(hostrpc.PermissionRequest{
				Description: fmt.Sprintf("Allow tool call: %s", fullName),
				Intent:      decision.Reason,
			})
			if err == nil && approved {
				break
			}
			if err != nil {
				return ToolResult{}, fmt.Errorf("approval failed: %w", err)
			}
			return ToolResult{}, errors.New("denied by user")
		}
		approvalID := fmt.Sprintf("%s-%d", sessionID, time.Now().UnixNano())
		ch := make(chan bool, 1)

		m.approvalMu.Lock()
		m.pendingApprovals[approvalID] = pendingApproval{
			ch:        ch,
			sessionID: sessionID,
			tool:      fullName,
			reason:    decision.Reason,
			args:      args,
		}
		m.approvalMu.Unlock()

		if m.bus != nil {
			m.bus.Publish(bus.NewEvent(bus.EventApprovalNeeded, sessionID, map[string]interface{}{
				"approval_id": approvalID,
				"tool":        fullName,
				"args":        args,
				"reason":      decision.Reason,
			}))
		}

		waitCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		select {
		case approved := <-ch:
			if !approved {
				return ToolResult{}, errors.New("denied by user")
			}
		case <-waitCtx.Done():
			m.approvalMu.Lock()
			delete(m.pendingApprovals, approvalID)
			m.approvalMu.Unlock()
			return ToolResult{}, errors.New("approval timed out")
		}
	case policy.DecisionDeny:
		return ToolResult{}, errors.New("denied by policy")
	default:
		return ToolResult{}, errors.New("unknown policy decision")
	}

	if m.bus != nil {
		m.bus.Publish(bus.NewEvent(bus.EventToolExecuting, sessionID, map[string]interface{}{
			"tool": fullName,
		}))
	}

	res, err := client.CallTool(ctx, name, args)
	if err != nil {
		if m.bus != nil {
			m.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, sessionID, map[string]interface{}{
				"tool":  fullName,
				"error": err.Error(),
			}))
		}
		return ToolResult{}, err
	}

	if m.bus != nil {
		m.bus.Publish(bus.NewEvent(bus.EventToolComplete, sessionID, map[string]interface{}{
			"tool":   fullName,
			"result": res,
		}))
	}
	return res, nil
}

func (m *Manager) buildClient(name string, sc ServerConfig) (*Client, error) {
	proto := sc.ProtocolVersion
	switch strings.ToLower(strings.TrimSpace(sc.Transport)) {
	case "bundled":
		provider, err := BundledProvider(name)
		if err != nil {
			return nil, err
		}
		tr := NewBundledTransport(provider)
		return NewClient(tr, proto), nil
	case "stdio":
		tr := NewStdioTransport(sc.Command, sc.Cwd, sc.Env)
		return NewClient(tr, proto), nil
	case "http":
		headers := map[string]string{}
		for k, v := range sc.Headers {
			headers[k] = v
		}
		if sc.Auth != nil {
			if err := m.applyAuth(headers, *sc.Auth); err != nil {
				return nil, err
			}
		}
		tr := NewHTTPTransport(sc.URL, headers)
		return NewClient(tr, proto), nil
	default:
		return nil, fmt.Errorf("unsupported transport: %s", sc.Transport)
	}
}

func (m *Manager) applyAuth(headers map[string]string, ac AuthConfig) error {
	if strings.ToLower(strings.TrimSpace(ac.Type)) != "oauth" {
		return nil
	}

	ref := strings.TrimSpace(ac.TokenRef)
	if ref == "" {
		return errors.New("oauth requires token_ref")
	}
	if !strings.HasPrefix(ref, "keychain:") {
		return errors.New("unsupported token_ref")
	}
	if m.keychain == nil {
		return errors.New("keychain not available")
	}

	key := strings.TrimPrefix(ref, "keychain:")
	token, err := m.keychain.Get(key)
	if err != nil {
		return err
	}
	if strings.TrimSpace(token) == "" {
		return errors.New("empty token")
	}

	headers["Authorization"] = "Bearer " + token
	return nil
}

func (m *Manager) listToolsCached(ctx context.Context, name string, c *Client, refresh bool) ([]Tool, error) {
	if !refresh {
		m.cacheMu.RLock()
		item, ok := m.cache[name]
		m.cacheMu.RUnlock()
		if ok && time.Since(item.fetchedAt) < 30*time.Second {
			return item.tools, nil
		}
	}

	tools, err := c.ListTools(ctx)
	if err != nil {
		return nil, err
	}

	m.cacheMu.Lock()
	m.cache[name] = cachedTools{fetchedAt: time.Now().UTC(), tools: tools}
	m.cacheMu.Unlock()

	return tools, nil
}

func splitToolName(full string) (string, string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	if strings.Contains(full, ":") {
		parts := strings.SplitN(full, ":", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	if strings.Contains(full, "/") {
		parts := strings.SplitN(full, "/", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "", ""
}
