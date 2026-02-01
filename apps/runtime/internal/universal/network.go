package universal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
)

// NetworkAgent represents an agent accessed via network URL
type NetworkAgent struct {
	AgentInfo
	EndpointURL    string         `json:"endpoint_url"`
	RegistryURL    string         `json:"registry_url"`
	AuthType       string         `json:"auth_type"` // "claim", "oauth2", "api_key"
	SocialFeatures SocialFeatures `json:"social_features"`
	CacheTTL       time.Duration  `json:"cache_ttl"`
	CachedAt       time.Time      `json:"cached_at"`
}

// SocialFeatures represents social networking capabilities
type SocialFeatures struct {
	Posts      bool   `json:"posts"`
	Voting     bool   `json:"voting"`
	Follows    bool   `json:"follows"`
	Feeds      bool   `json:"feeds"`
	Reputation bool   `json:"reputation"`
	Endpoint   string `json:"endpoint"`
}

// ClaimLink represents a claim authorization link
type ClaimLink struct {
	Token      string    `json:"token"`
	AgentID    string    `json:"agent_id"`
	NetworkURL string    `json:"network_url"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Status     string    `json:"status"` // pending, completed, expired
}

// NetworkConfig contains configuration for network agents
type NetworkConfig struct {
	BaseURL         string            `json:"base_url"`
	RegistryPath    string            `json:"registry_path"`
	AuthType        string            `json:"auth_type"`
	ClientID        string            `json:"client_id"`
	ClientSecret    string            `json:"client_secret,omitempty"`
	Scopes          []string          `json:"scopes"`
	Timeout         time.Duration     `json:"timeout"`
	CacheTTL        time.Duration     `json:"cache_ttl"`
	FollowRedirects bool              `json:"follow_redirects"`
	Headers         map[string]string `json:"headers"`
}

// NetworkAdapter manages network-based agents
type NetworkAdapter struct {
	config     NetworkConfig
	httpClient *http.Client
	mu         sync.RWMutex
	running    bool
	stopCh     chan struct{}
	cache      map[string]*NetworkAgent
}

// NewNetworkAdapter creates a new network adapter
func NewNetworkAdapter(config NetworkConfig) *NetworkAdapter {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 5 * time.Minute
	}

	return &NetworkAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if !config.FollowRedirects {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
		stopCh: make(chan struct{}),
		cache:  make(map[string]*NetworkAgent),
	}
}

// Protocol returns the protocol name
func (a *NetworkAdapter) Protocol() string {
	return "network"
}

// Name returns the adapter name
func (a *NetworkAdapter) Name() string {
	return "network-remote"
}

// Version returns the adapter version
func (a *NetworkAdapter) Version() string {
	return "1.0.0"
}

// Detect discovers agents from the network registry
func (a *NetworkAdapter) Detect(ctx context.Context) ([]DetectedAgent, error) {
	registryURL := a.getRegistryURL()

	req, err := http.NewRequestWithContext(ctx, "GET", registryURL+"/discover", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range a.config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to discover agents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status: %d", resp.StatusCode)
	}

	var agents struct {
		Agents []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			URL      string `json:"url"`
			Version  string `json:"version"`
			Protocol string `json:"protocol"`
		} `json:"agents"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	detected := make([]DetectedAgent, 0, len(agents.Agents))
	for _, agent := range agents.Agents {
		detected = append(detected, DetectedAgent{
			AgentInfo: AgentInfo{
				Identity: AgentIdentity{
					ID:      agent.ID,
					Name:    agent.Name,
					Version: agent.Version,
				},
				Protocol: agent.Protocol,
				Endpoint: EndpointInfo{
					Type: "http",
					URL:  agent.URL,
				},
				Capabilities: []string{"messaging", "social"},
				HealthStatus: "unknown",
			},
			DetectionMethod: "network_registry",
			Confidence:      0.9,
		})
	}

	return detected, nil
}

// Connect establishes a connection to a network agent
func (a *NetworkAdapter) Connect(ctx context.Context, info AgentInfo, config AgentConfig) (*AgentConnection, error) {
	agentURL := info.Endpoint.URL
	if agentURL == "" {
		agentURL = a.getAgentURL(info.Identity.ID)
	}

	// Verify agent is accessible
	req, err := http.NewRequestWithContext(ctx, "GET", agentURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent returned status: %d", resp.StatusCode)
	}

	now := time.Now()
	return &AgentConnection{
		ID: uuid.New().String(),
		AgentInfo: AgentInfo{
			Identity: AgentIdentity{
				ID:      info.Identity.ID,
				Name:    info.Identity.Name,
				Version: info.Identity.Version,
			},
			Protocol: "network",
			Endpoint: EndpointInfo{
				Type: "http",
				URL:  agentURL,
			},
			Capabilities: info.Capabilities,
			HealthStatus: "connected",
		},
		State:        ConnectionStateConnected,
		Protocol:     "network",
		Adapter:      a,
		LastActivity: now,
		CreatedAt:    now,
		ConnectedAt:  &now,
	}, nil
}

// Send transmits a message to a network agent
func (a *NetworkAdapter) Send(ctx context.Context, conn *AgentConnection, msg *UniversalMessage) error {
	endpoint := conn.AgentInfo.Endpoint.URL + "/message"

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add auth if configured
	if a.config.ClientID != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.ClientID)
	}

	resp, err := a.httpClient.Do(req)
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

// Receive receives a message from a network agent
func (a *NetworkAdapter) Receive(ctx context.Context, conn *AgentConnection) (*UniversalMessage, error) {
	endpoint := conn.AgentInfo.Endpoint.URL + "/message"

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, fmt.Errorf("no message available")
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("agent returned error: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var msg UniversalMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}

	return &msg, nil
}

// Disconnect closes the connection to a network agent
func (a *NetworkAdapter) Disconnect(ctx context.Context, conn *AgentConnection) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		// Already stopped
		return nil
	}
	a.running = false

	close(a.stopCh)
	a.stopCh = make(chan struct{})

	return nil
}

// HealthCheck checks the health of a network agent
func (a *NetworkAdapter) HealthCheck(ctx context.Context, conn *AgentConnection) error {
	endpoint := conn.AgentInfo.Endpoint.URL + "/health"

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// Install installs a network agent from a URL
func (a *NetworkAdapter) Install(ctx context.Context, ref string, config AgentConfig) error {
	// ref is a URL like "https://moltbook.com/agent/xyz"
	manifestURL := ref
	if !hasExtension(manifestURL) {
		manifestURL = manifestURL + "/agent.json"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", manifestURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("manifest not found: %d", resp.StatusCode)
	}

	var pkg AgentPackage
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return fmt.Errorf("failed to decode manifest: %w", err)
	}

	// Cache the agent
	agent := &NetworkAgent{
		AgentInfo: AgentInfo{
			Identity: AgentIdentity{
				ID:      pkg.Name,
				Name:    pkg.Name,
				Version: pkg.Version,
			},
			Protocol:     "network",
			Capabilities: pkg.Capabilities,
		},
		EndpointURL: ref,
		CacheTTL:    a.config.CacheTTL,
		CachedAt:    time.Now(),
	}

	a.mu.Lock()
	a.cache[pkg.Name] = agent
	a.mu.Unlock()

	return nil
}

// Uninstall removes a network agent from cache
func (a *NetworkAdapter) Uninstall(ctx context.Context, ref string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.cache, ref)
	return nil
}

// FetchManifest fetches an agent manifest from a URL
func (a *NetworkAdapter) FetchManifest(ctx context.Context, manifestURL string) (*AgentPackage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", manifestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest not found: %d", resp.StatusCode)
	}

	var pkg AgentPackage
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return &pkg, nil
}

// SearchAgents searches the network registry for agents
func (a *NetworkAdapter) SearchAgents(ctx context.Context, query string, criteria SearchCriteria) ([]AgentInfo, error) {
	searchURL := a.getRegistryURL() + "/search"

	params := url.Values{}
	params.Set("q", query)
	if criteria.Capability != "" {
		params.Set("capability", criteria.Capability)
	}
	if criteria.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", criteria.Limit))
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %d", resp.StatusCode)
	}

	var results struct {
		Agents []AgentInfo `json:"agents"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results.Agents, nil
}

// Helper functions
func (a *NetworkAdapter) getRegistryURL() string {
	if a.config.RegistryPath != "" {
		return a.config.BaseURL + "/" + a.config.RegistryPath
	}
	return a.config.BaseURL + "/api/v1/agents"
}

func (a *NetworkAdapter) getAgentURL(agentID string) string {
	return a.config.BaseURL + "/agents/" + agentID
}

func hasExtension(s string) bool {
	extensions := []string{".json", ".yaml", ".yml", ".md"}
	for _, ext := range extensions {
		if len(s) >= len(ext) && s[len(s)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// SearchCriteria defines search parameters for agent discovery
type SearchCriteria struct {
	Capability string `json:"capability,omitempty"`
	Tags       string `json:"tags,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}
