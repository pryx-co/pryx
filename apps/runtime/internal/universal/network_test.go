package universal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNetworkAdapter(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/agents/discover":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"agents": []map[string]interface{}{
					{
						"id":       "network-agent-1",
						"name":     "Network Agent 1",
						"url":      "http://example.com/agent/1",
						"version":  "1.0.0",
						"protocol": "network",
					},
				},
			})
		case "/api/v1/agents/agent-1/health":
			w.WriteHeader(http.StatusOK)
		case "/api/v1/agents/agent-1/message":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusOK)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := NetworkConfig{
		BaseURL:      server.URL,
		RegistryPath: "api/v1/agents",
		Timeout:      10 * time.Second,
	}

	adapter := NewNetworkAdapter(config)
	ctx := context.Background()

	// Test Detect
	agents, err := adapter.Detect(ctx)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}

	// Test Connect
	info := AgentInfo{
		Identity: AgentIdentity{
			ID:      "agent-1",
			Name:    "Test Agent",
			Version: "1.0.0",
		},
		Protocol: "network",
		Endpoint: EndpointInfo{
			Type: "http",
			URL:  server.URL + "/api/v1/agents/agent-1",
		},
	}

	conn, err := adapter.Connect(ctx, info, AgentConfig{})
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if conn.State != ConnectionStateConnected {
		t.Errorf("expected connected state, got %s", conn.State)
	}

	// Test Send
	msg := &UniversalMessage{
		ID:     "test-msg",
		From:   AgentIdentity{ID: "from-1"},
		To:     AgentIdentity{ID: "agent-1"},
		Action: "test",
	}
	err = adapter.Send(ctx, conn, msg)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	// Test HealthCheck
	err = adapter.HealthCheck(ctx, conn)
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}

	// Test Disconnect
	err = adapter.Disconnect(ctx, conn)
	if err != nil {
		t.Fatalf("disconnect failed: %v", err)
	}
}

func TestNetworkAdapterSearch(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/agents/search" && r.URL.Query().Get("q") == "test" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"agents": []map[string]interface{}{
					{
						"id":       "search-result-1",
						"name":     "Search Result 1",
						"version":  "1.0.0",
						"protocol": "network",
					},
				},
			})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := NetworkConfig{
		BaseURL:      server.URL,
		RegistryPath: "api/v1/agents",
	}

	adapter := NewNetworkAdapter(config)
	ctx := context.Background()

	// Test SearchAgents
	criteria := SearchCriteria{
		Capability: "messaging",
		Limit:      10,
	}

	agents, err := adapter.SearchAgents(ctx, "test", criteria)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("expected 1 search result, got %d", len(agents))
	}
}

func TestNetworkAdapterFetchManifest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/agent/xyz/agent.json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(AgentPackage{
				Name:         "test-agent",
				Version:      "1.0.0",
				Description:  "Test agent",
				Protocols:    []string{"network"},
				Capabilities: []string{"messaging"},
			})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := NetworkConfig{
		BaseURL: server.URL,
	}

	adapter := NewNetworkAdapter(config)
	ctx := context.Background()

	// Test FetchManifest
	pkg, err := adapter.FetchManifest(ctx, server.URL+"/agent/xyz/agent.json")
	if err != nil {
		t.Fatalf("fetch manifest failed: %v", err)
	}
	if pkg.Name != "test-agent" {
		t.Errorf("expected agent name 'test-agent', got '%s'", pkg.Name)
	}
}

func TestNetworkAgentCache(t *testing.T) {
	// Create test server that handles agent manifest requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/agent/test-agent/agent.json":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(AgentPackage{
				Name:        "test-agent",
				Version:     "1.0.0",
				Capabilities: []string{"messaging", "social"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := NetworkConfig{
		BaseURL:  server.URL,
		CacheTTL: 1 * time.Minute,
	}

	adapter := NewNetworkAdapter(config)
	ctx := context.Background()

	// Install an agent
	err := adapter.Install(ctx, server.URL+"/agent/test-agent", AgentConfig{})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Check cache
	adapter.mu.RLock()
	cached, exists := adapter.cache["test-agent"]
	adapter.mu.RUnlock()

	if !exists {
		t.Fatal("expected agent to be cached")
	}

	if cached.AgentInfo.Identity.Name != "test-agent" {
		t.Errorf("expected cached name 'test-agent', got '%s'", cached.AgentInfo.Identity.Name)
	}

	// Uninstall
	err = adapter.Uninstall(ctx, "test-agent")
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	// Check cache is empty
	adapter.mu.RLock()
	_, exists = adapter.cache["test-agent"]
	adapter.mu.RUnlock()

	if exists {
		t.Fatal("expected agent to be removed from cache")
	}
}


