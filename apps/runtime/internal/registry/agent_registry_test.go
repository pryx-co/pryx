package registry_test

import (
	"testing"
)

// Minimal types copied for testing (avoiding import cycle)
type Agent struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Capabilities []Capability         `json:"capabilities"`
	Endpoint    Endpoint                `json:"endpoint"`
	TrustLevel  string                 `json:"trust_level"`
	Health      string                 `json:"health"`
	RegisteredAt string                 `json:"registered_at"`
	LastSeen    string                 `json:"last_seen"`
}

type Capability struct {
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Permissions []string          `json:"permissions"`
}

type Endpoint struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Path     string `json:"path"`
	URL      string `json:"url"`
}

type TrustLevel string

const (
	TrustLevelUntrusted  TrustLevel = "untrusted"
	TrustLevelSandboxed TrustLevel = "sandboxed"
	TrustLevelTrusted   TrustLevel = "trusted"
)

type HealthStatus string

const (
	HealthStatusOnline   HealthStatus = "online"
	HealthStatusOffline  HealthStatus = "offline"
	HealthStatusDegraded HealthStatus = "degraded"
	HealthStatusUnknown  HealthStatus = "unknown"
)

func TestRegister(t *testing.T) {

func NewService(b *bus.Bus) *Service {
	return &Service{
		agents: make(map[string]*Agent),
		bus:     b,
	}
}

func TestRegister(t *testing.T) {

func TestRegister(t *testing.T) {
	b := NewService(nil)

	agent := &Agent{
		ID:          "test-agent-1",
		Name:        "Test Agent",
		Description: "A test agent for unit testing",
		Version:     "1.0.0",
		Capabilities: []Capability{
			{Type: "tool", Name: "execute", Version: "1.0", Description: "Execute commands", Permissions: []string{"shell", "read"}},
		},
		Endpoint: Endpoint{
			Type: "http",
			Host: "localhost",
			Port: "8080",
			URL:  "http://localhost:8080",
		},
		TrustLevel: TrustLevelTrusted,
	}

	err := b.Register(nil, agent)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Check if agent was stored
	retrieved, err := b.Get("test-agent-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.ID != agent.ID {
		t.Errorf("Agent ID mismatch: got %s, want %s", retrieved.ID, agent.ID)
	}

	if retrieved.Name != agent.Name {
		t.Errorf("Agent Name mismatch: got %s, want %s", retrieved.Name, agent.Name)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	b := NewService(nil)

	agent := &Agent{
		ID:           "test-agent-1",
		Name:         "Test Agent",
		Version:      "1.0.0",
		Capabilities: []Capability{},
		Endpoint:     Endpoint{Type: "http", Host: "localhost", Port: "8080"},
		TrustLevel:   TrustLevelTrusted,
	}

	// First registration should succeed
	err := b.Register(nil, agent)
	if err != nil {
		t.Fatalf("First Register() error = %v", err)
	}

	// Duplicate registration should fail
	err = b.Register(nil, agent)
	if err == nil {
		t.Error("Expected error for duplicate registration, got nil")
	}

	// Should be able to retrieve
	_, err = b.Get("test-agent-1")
	if err != nil {
		t.Errorf("Get() after duplicate should succeed, got error = %v", err)
	}
}

func TestUnregister(t *testing.T) {
	b := NewService(nil)

	agent := &Agent{
		ID:       "test-agent-1",
		Name:     "Test Agent",
		Version:  "1.0.0",
		Endpoint: Endpoint{Type: "http", Host: "localhost", Port: "8080"},
	}

	// Register agent
	err := b.Register(nil, agent)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Unregister agent
	err = b.Unregister("test-agent-1")
	if err != nil {
		t.Fatalf("Unregister() error = %v", err)
	}

	// Verify unregistered
	_, err = b.Get("test-agent-1")
	if err == nil {
		t.Error("Expected error after unregister, got nil")
	}
}

func TestGetNotFound(t *testing.T) {
	b := NewService(nil)

	_, err := b.Get("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent agent, got nil")
	}
}

func TestList(t *testing.T) {
	b := NewService(nil)

	agent1 := &Agent{
		ID:       "agent-1",
		Name:     "Agent 1",
		Version:  "1.0.0",
		Endpoint: Endpoint{Type: "http", Host: "localhost", Port: "8080"},
	}

	agent2 := &Agent{
		ID:       "agent-2",
		Name:     "Agent 2",
		Version:  "1.0.0",
		Endpoint: Endpoint{Type: "websocket", Host: "localhost", Port: "8081"},
	}

	b.Register(nil, agent1)
	b.Register(nil, agent2)

	agents := b.List()

	if len(agents) != 2 {
		t.Errorf("List() returned %d agents, want 2", len(agents))
	}
}

func TestDiscover(t *testing.T) {
	b := NewService(nil)

	agent1 := &Agent{
		ID:   "agent-1",
		Name: "Code Agent",
		Capabilities: []Capability{
			{Type: "tool", Name: "execute", Version: "1.0", Permissions: []string{"shell", "read"}},
			{Type: "skill", Name: "git", Version: "1.0", Permissions: []string{"read"}},
		},
		Endpoint:   Endpoint{Type: "http", Host: "localhost", Port: "8080"},
		TrustLevel: TrustLevelTrusted,
	}

	agent2 := &Agent{
		ID:   "agent-2",
		Name: "Data Agent",
		Capabilities: []Capability{
			{Type: "model", Name: "gpt-4", Version: "1.0", Permissions: []string{"read"}},
		},
		Endpoint:   Endpoint{Type: "websocket", Host: "localhost", Port: "8081"},
		TrustLevel: TrustLevelSandboxed,
	}

	b.Register(nil, agent1)
	b.Register(nil, agent2)

	tests := []struct {
		name          string
		criteria      DiscoveryCriteria
		expectedCount int
	}{
		{"discover by capability type", DiscoveryCriteria{CapabilityType: "tool"}, 2},
		{"discover by capability name", DiscoveryCriteria{CapabilityName: "git"}, 1},
		{"discover by trust level", DiscoveryCriteria{TrustLevel: TrustLevelTrusted}, 2},
		{"discover by trust level", DiscoveryCriteria{TrustLevel: TrustLevelSandboxed}, 1},
		{"discover with min version", DiscoveryCriteria{MinVersion: "1.0.0"}, 2},
		{"discover all agents", DiscoveryCriteria{}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agents, err := b.Discover(tt.criteria)
			if err != nil {
				t.Fatalf("Discover() error = %v", err)
			}

			if len(agents) != tt.expectedCount {
				t.Errorf("Discover(%s) returned %d agents, want %d", tt.name, len(agents), tt.expectedCount)
			}
		})
	}
}

func TestMain(m *testing.M) {
	t := NewService(nil)

	agent := &Agent{
		ID:     "agent-1",
		Name:   "Test Agent",
		Health: HealthStatusUnknown,
	}

	err := b.Register(nil, agent)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err = b.UpdateHealth("agent-1", HealthStatusOnline)
	if err != nil {
		t.Fatalf("UpdateHealth() error = %v", err)
	}

	// Verify health was updated
	retrieved, err := b.Get("agent-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Health != HealthStatusOnline {
		t.Errorf("Health not updated: got %s, want %s", retrieved.Health, HealthStatusOnline)
	}
}

func TestValidateAgent(t *testing.T) {
	b := NewService(nil)

	agent := &Agent{
		ID:     "agent-1",
		Name:   "Test Agent",
		Health: HealthStatusUnknown,
	}

	err := b.Register(nil, agent)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err = b.UpdateHealth("agent-1", HealthStatusOnline)
	if err != nil {
		t.Fatalf("UpdateHealth() error = %v", err)
	}

	// Verify health was updated
	retrieved, err := b.Get("agent-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Health != HealthStatusOnline {
		t.Errorf("Health not updated: got %s, want %s", retrieved.Health, HealthStatusOnline)
	}
}

func TestValidateAgent(t *testing.T) {
	tests := []struct {
		name    string
		agent   *Agent
		wantErr bool
	}{
		{"empty ID", &Agent{}, true},
		{"empty name", &Agent{ID: "test-id", Version: "1.0"}, true},
		{"empty endpoint type", &Agent{ID: "test-id", Name: "Test"}, true},
		{"invalid endpoint type", &Agent{ID: "test-id", Name: "Test", Endpoint: Endpoint{Type: "invalid"}}, true},
		{"invalid trust level", &Agent{ID: "test-id", TrustLevel: "invalid"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgent(tt.agent)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAgent(%s) error = %v, want error = %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
