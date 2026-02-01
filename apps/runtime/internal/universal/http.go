package universal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// HTTPAdapter implements the AgentAdapter for HTTP/REST agents
type HTTPAdapter struct {
	baseURL string
	client  *http.Client
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
}

// NewHTTPAdapter creates a new HTTP adapter
func NewHTTPAdapter(baseURL string) *HTTPAdapter {
	return &HTTPAdapter{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		stopCh: make(chan struct{}),
	}
}

// Protocol returns the protocol name
func (a *HTTPAdapter) Protocol() string {
	return "http"
}

// Name returns the adapter name
func (a *HTTPAdapter) Name() string {
	return "http-rest"
}

// Version returns the adapter version
func (a *HTTPAdapter) Version() string {
	return "1.0.0"
}

// Connect establishes an HTTP connection to an agent
func (a *HTTPAdapter) Connect(ctx context.Context, info AgentInfo, config AgentConfig) (*AgentConnection, error) {
	// Use endpoint URL if provided
	baseURL := a.baseURL
	if info.Endpoint.URL != "" {
		baseURL = info.Endpoint.URL
	}

	// Test connection
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent returned status: %d", resp.StatusCode)
	}

	now := time.Now()
	return &AgentConnection{
		ID: fmt.Sprintf("http-%d", time.Now().UnixNano()),
		AgentInfo: AgentInfo{
			Identity: AgentIdentity{
				ID:      info.Identity.ID,
				Name:    info.Identity.Name,
				Version: info.Identity.Version,
			},
			Protocol:     "http",
			Endpoint:     info.Endpoint,
			Capabilities: info.Capabilities,
			HealthStatus: "healthy",
		},
		State:        ConnectionStateConnected,
		Protocol:     "http",
		Adapter:      a,
		LastActivity: now,
		CreatedAt:    now,
		ConnectedAt:  &now,
	}, nil
}

// Send transmits a message to the agent via HTTP POST
func (a *HTTPAdapter) Send(ctx context.Context, conn *AgentConnection, msg *UniversalMessage) error {
	translator := NewTranslator()
	httpMsg := translator.ToOpenClaw(msg)

	data, err := json.Marshal(httpMsg)
	if err != nil {
		return err
	}

	url := conn.AgentInfo.Endpoint.URL + "/message"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("agent returned error: %s", string(body))
	}

	return nil
}

// Receive waits for a message from the agent via polling
func (a *HTTPAdapter) Receive(ctx context.Context, conn *AgentConnection) (*UniversalMessage, error) {
	url := conn.AgentInfo.Endpoint.URL + "/message"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		// No message available
		return nil, fmt.Errorf("no message available")
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("agent returned error: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var openclawMsg OpenClawMessage
	if err := json.Unmarshal(data, &openclawMsg); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	translator := NewTranslator()
	return translator.ToUniversal(&openclawMsg), nil
}

// Disconnect closes the HTTP connection
func (a *HTTPAdapter) Disconnect(ctx context.Context, conn *AgentConnection) error {
	a.mu.Lock()
	a.running = false
	a.mu.Unlock()

	close(a.stopCh)
	a.stopCh = make(chan struct{})

	return nil
}

// HealthCheck checks the agent health via HTTP
func (a *HTTPAdapter) HealthCheck(ctx context.Context, conn *AgentConnection) error {
	url := conn.AgentInfo.Endpoint.URL + "/health"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// Detect discovers HTTP agents by scanning endpoints
func (a *HTTPAdapter) Detect(ctx context.Context) ([]DetectedAgent, error) {
	detector := NewDetector([]int{8080, 3000})
	return detector.DetectByProtocol(ctx, "http")
}

// Install installs an HTTP agent package (not implemented for HTTP)
func (a *HTTPAdapter) Install(ctx context.Context, ref string, config AgentConfig) error {
	return fmt.Errorf("HTTP adapter does not support installation")
}

// Uninstall removes an HTTP agent package (not implemented for HTTP)
func (a *HTTPAdapter) Uninstall(ctx context.Context, ref string) error {
	return fmt.Errorf("HTTP adapter does not support uninstallation")
}

// HTTPPackage returns the HTTP adapter package definition
func HTTPPackage() AgentPackage {
	return AgentPackage{
		Name:        "http-agent",
		Version:     "1.0.0",
		Description: "Generic HTTP/REST agent adapter",
		Protocols:   []string{"http"},
		Endpoints: []EndpointInfo{
			{
				Type: "http",
				URL:  "http://localhost:8080",
				Host: "localhost",
				Port: 8080,
			},
		},
		Capabilities: []string{"http", "rest"},
		Install: InstallConfig{
			Type:   "local",
			Source: "",
		},
		Permissions:  []string{},
		Dependencies: map[string]string{},
	}
}
