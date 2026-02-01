package universal

import (
	"context"
	"sync"
)

// Registry manages agent registration and discovery
type Registry struct {
	mu      sync.RWMutex
	agents  map[string]*AgentInfo
	running bool
	stopCh  chan struct{}
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]*AgentInfo),
		stopCh: make(chan struct{}),
	}
}

// Start initializes the registry
func (r *Registry) Start(ctx context.Context) {
	r.running = true
}

// Stop gracefully shuts down the registry
func (r *Registry) Stop(ctx context.Context) {
	r.running = false
	close(r.stopCh)
}

// Register adds an agent to the registry
func (r *Registry) Register(agent *AgentInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[agent.Identity.ID] = agent
}

// Unregister removes an agent from the registry
func (r *Registry) Unregister(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agents, agentID)
}

// Get retrieves an agent by ID
func (r *Registry) Get(agentID string) (*AgentInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agent, exists := r.agents[agentID]
	return agent, exists
}

// List returns all registered agents
func (r *Registry) List() []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentInfo, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents
}

// ListByProtocol returns agents filtered by protocol
func (r *Registry) ListByProtocol(protocol string) []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var agents []*AgentInfo
	for _, agent := range r.agents {
		if agent.Protocol == protocol {
			agents = append(agents, agent)
		}
	}
	return agents
}

// Count returns the number of registered agents
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}
