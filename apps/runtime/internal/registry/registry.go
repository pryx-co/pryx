package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"pryx-core/internal/bus"
)

// Agent represents an AI agent that can register with Pryx interoperability system
type Agent struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Version      string       `json:"version"`
	Capabilities []Capability `json:"capabilities"`
	Endpoint     Endpoint     `json:"endpoint"`
	TrustLevel   TrustLevel   `json:"trust_level"`
	Health       HealthStatus `json:"health"`
	RegisteredAt time.Time    `json:"registered_at"`
	LastSeen     time.Time    `json:"last_seen"`
}

// Capability represents an agent's advertised capability
type Capability struct {
	Type        string   `json:"type"` // "tool", "skill", "model"
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// Endpoint represents where an agent can be contacted
type Endpoint struct {
	Type string `json:"type"` // "http", "websocket"
	Host string `json:"host"`
	Port string `json:"port"`
	Path string `json:"path"` // optional path
	URL  string `json:"url"`  // full URL (constructed from host+port+path)
}

// TrustLevel represents the trust level for an agent
type TrustLevel string

const (
	TrustLevelUntrusted TrustLevel = "untrusted"
	TrustLevelSandboxed TrustLevel = "sandboxed"
	TrustLevelTrusted   TrustLevel = "trusted"
)

// HealthStatus represents the health status of an agent
type HealthStatus string

const (
	HealthStatusOnline   HealthStatus = "online"
	HealthStatusOffline  HealthStatus = "offline"
	HealthStatusDegraded HealthStatus = "degraded"
	HealthStatusUnknown  HealthStatus = "unknown"
)

// Service manages the agent registry
type Service struct {
	agents map[string]*Agent
	mu     sync.RWMutex
	bus    *bus.Bus
}

// NewService creates a new agent registry service
func NewService(b *bus.Bus) *Service {
	return &Service{
		agents: make(map[string]*Agent),
		bus:    b,
	}
}

// Register registers a new agent with the registry
func (s *Service) Register(ctx context.Context, agent *Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate agent ID
	if agent.ID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	// Check for duplicate registration
	if _, exists := s.agents[Agent.ID]; exists {
		return fmt.Errorf("agent with ID %s already registered", Agent.ID)
	}

	// Set registration time
	Agent.RegisteredAt = time.Now().UTC()
	Agent.LastSeen = Agent.RegisteredAt

	// Store agent
	s.agents[Agent.ID] = Agent

	// Publish event
	s.bus.Publish(bus.NewEvent("agent.registered", "", map[string]interface{}{
		"agent_id": Agent.ID,
		"name":     Agent.Name,
		"version":  Agent.Version,
	}))

	return nil
}

// Unregister removes an agent from the registry
func (s *Service) Unregister(agentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.agents[agentID]; !exists {
		return fmt.Errorf("agent with ID %s not found", agentID)
	}

	delete(s.agents, agentID)

	// Publish event
	s.bus.Publish(bus.NewEvent("agent.unregistered", "", map[string]interface{}{
		"agent_id": agentID,
	}))

	return nil
}

// Get retrieves an agent by ID
func (s *Service) Get(agentID string) (*Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, exists := s.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}

	return agent, nil
}

// List returns all registered agents
func (s *Service) List() []*Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]*Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}

	return agents
}

// Discover finds agents matching criteria
func (s *Service) Discover(criteria DiscoveryCriteria) ([]*Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*Agent, 0)

	for _, agent := range s.agents {
		if s.matchesCriteria(agent, criteria) {
			results = append(results, agent)
		}
	}

	return results, nil
}

// UpdateHealth updates an agent's health status
func (s *Service) UpdateHealth(agentID string, status HealthStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[agentID]
	if !exists {
		return fmt.Errorf("agent with ID %s not found", agentID)
	}

	agent.Health = status
	agent.LastSeen = time.Now().UTC()

	s.bus.Publish(bus.NewEvent("agent.health_updated", "", map[string]interface{}{
		"agent_id": agentID,
		"status":   status,
	}))

	return nil
}

// DiscoveryCriteria specifies search criteria for agent discovery
type DiscoveryCriteria struct {
	CapabilityType string `json:"capability_type"` // filter by capability type
	CapabilityName string `json:"capability_name"` // filter by capability name
	TrustLevel     string `json:"trust_level"`     // filter by trust level
	MinVersion     string `json:"min_version"`     // minimum version requirement
	MaxVersion     string `json:"max_version"`     // maximum version constraint
}

// matchesCriteria checks if an agent matches discovery criteria
func (s *Service) matchesCriteria(agent *Agent, criteria DiscoveryCriteria) bool {
	// Check trust level
	if criteria.TrustLevel != "" && agent.TrustLevel != criteria.TrustLevel {
		return false
	}

	// Check capability type
	if criteria.CapabilityType != "" {
		hasType := false
		for _, cap := range agent.Capabilities {
			if cap.Type == criteria.CapabilityType {
				hasType = true
				break
			}
		}
		if !hasType {
			return false
		}
	}

	// Check capability name
	if criteria.CapabilityName != "" {
		hasName := false
		for _, cap := range agent.Capabilities {
			if cap.Name == criteria.CapabilityName {
				hasName = true
				break
			}
		}
		if !hasName {
			return false
		}
	}

	return true
}

// ValidateAgent checks if an agent registration is valid
func ValidateAgent(agent *Agent) error {
	if agent.ID == "" {
		return fmt.Errorf("agent ID is required")
	}

	if agent.Name == "" {
		return fmt.Errorf("agent name is required")
	}

	if agent.Endpoint.Type == "" {
		return fmt.Errorf("endpoint type is required")
	}

	if agent.Endpoint.Type != "http" && agent.Endpoint.Type != "websocket" {
		return fmt.Errorf("endpoint type must be 'http' or 'websocket'")
	}

	if agent.TrustLevel != TrustLevelUntrusted &&
		agent.TrustLevel != TrustLevelSandboxed &&
		agent.TrustLevel != TrustLevelTrusted {
		return fmt.Errorf("invalid trust level: must be 'untrusted', 'sandboxed', or 'trusted'")
	}

	return nil
}
