package agentbus

import (
	"context"
	"sync"
	"time"

	"pryx-core/internal/bus"
)

// RegistryManager manages agent registration and discovery
type RegistryManager struct {
	mu       sync.RWMutex
	bus      *bus.Bus
	logger   *StructuredLogger
	agents   map[string]*AgentInfo
	byName   map[string][]string // name -> agent IDs
	byTag    map[string][]string // tag -> agent IDs
	byProto  map[string][]string // protocol -> agent IDs
	byNS     map[string][]string // namespace -> agent IDs
	listener chan *AgentInfo
	running  bool
	stopCh   chan struct{}
}

// NewRegistryManager creates a new registry manager
func NewRegistryManager(b *bus.Bus) *RegistryManager {
	return &RegistryManager{
		bus:      b,
		logger:   NewStructuredLogger("registry", "info"),
		agents:   make(map[string]*AgentInfo),
		byName:   make(map[string][]string),
		byTag:    make(map[string][]string),
		byProto:  make(map[string][]string),
		byNS:     make(map[string][]string),
		listener: make(chan *AgentInfo, 100),
		stopCh:   make(chan struct{}),
	}
}

// Start initializes the registry
func (r *RegistryManager) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return nil
	}
	r.running = true
	r.mu.Unlock()

	r.logger.Info("registry manager started", nil)

	// Publish event
	r.bus.Publish(bus.NewEvent("agentbus.registry.started", "", nil))

	return nil
}

// Stop gracefully shuts down the registry
func (r *RegistryManager) Stop(ctx context.Context) error {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return nil
	}
	r.running = false
	r.mu.Unlock()

	// Close stop channel
	close(r.stopCh)

	r.logger.Info("registry manager stopped", nil)

	r.bus.Publish(bus.NewEvent("agentbus.registry.stopped", "", nil))

	return nil
}

// Register adds an agent to the registry
func (r *RegistryManager) Register(ctx context.Context, agent *AgentInfo) (*AgentInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already registered
	if existing, ok := r.agents[agent.Identity.ID]; ok {
		// Update last seen
		existing.LastSeen = time.Now().UTC()
		existing.HealthStatus = agent.HealthStatus
		r.logger.Debug("updated existing agent", map[string]interface{}{
			"agent_id": agent.Identity.ID,
		})
		return existing, nil
	}

	// Set registration time
	now := time.Now().UTC()
	agent.LastSeen = now

	// Store agent
	r.agents[agent.Identity.ID] = agent

	// Index by name
	r.byName[agent.Identity.Name] = append(r.byName[agent.Identity.Name], agent.Identity.ID)

	// Index by tags
	for _, tag := range agent.Identity.Tags {
		r.byTag[tag] = append(r.byTag[tag], agent.Identity.ID)
	}

	// Index by protocol
	r.byProto[agent.Protocol] = append(r.byProto[agent.Protocol], agent.Identity.ID)

	// Index by namespace
	if agent.Identity.Namespace != "" {
		r.byNS[agent.Identity.Namespace] = append(r.byNS[agent.Identity.Namespace], agent.Identity.ID)
	}

	r.logger.Info("registered agent", map[string]interface{}{
		"agent_id":   agent.Identity.ID,
		"agent_name": agent.Identity.Name,
		"protocol":   agent.Protocol,
	})

	// Publish event
	r.bus.Publish(bus.NewEvent("agentbus.agent.registered", "", map[string]interface{}{
		"agent_id":   agent.Identity.ID,
		"agent_name": agent.Identity.Name,
		"protocol":   agent.Protocol,
		"endpoint":   agent.Endpoint.URL,
	}))

	return agent, nil
}

// Unregister removes an agent from the registry
func (r *RegistryManager) Unregister(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, ok := r.agents[agentID]
	if !ok {
		return nil // Already unregistered
	}

	// Remove from main registry
	delete(r.agents, agentID)

	// Remove from name index
	if ids, ok := r.byName[agent.Identity.Name]; ok {
		r.byName[agent.Identity.Name] = filterIDs(ids, agentID)
	}

	// Remove from tag indices
	for _, tag := range agent.Identity.Tags {
		if ids, ok := r.byTag[tag]; ok {
			r.byTag[tag] = filterIDs(ids, agentID)
		}
	}

	// Remove from protocol index
	if ids, ok := r.byProto[agent.Protocol]; ok {
		r.byProto[agent.Protocol] = filterIDs(ids, agentID)
	}

	// Remove from namespace index
	if agent.Identity.Namespace != "" {
		if ids, ok := r.byNS[agent.Identity.Namespace]; ok {
			r.byNS[agent.Identity.Namespace] = filterIDs(ids, agentID)
		}
	}

	r.logger.Info("unregistered agent", map[string]interface{}{
		"agent_id":   agentID,
		"agent_name": agent.Identity.Name,
	})

	// Publish event
	r.bus.Publish(bus.NewEvent("agentbus.agent.unregistered", "", map[string]interface{}{
		"agent_id":   agentID,
		"agent_name": agent.Identity.Name,
	}))

	return nil
}

// Get retrieves an agent by ID
func (r *RegistryManager) Get(ctx context.Context, agentID string) (*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[agentID]
	if !ok {
		return nil, nil
	}

	return agent, nil
}

// GetByName retrieves agents by name
func (r *RegistryManager) GetByName(ctx context.Context, name string) ([]*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentIDs := r.byName[name]
	agents := make([]*AgentInfo, 0, len(agentIDs))

	for _, id := range agentIDs {
		if agent, ok := r.agents[id]; ok {
			agents = append(agents, agent)
		}
	}

	return agents, nil
}

// List returns all registered agents
func (r *RegistryManager) List(ctx context.Context) ([]*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentInfo, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

// ListByProtocol returns agents filtered by protocol
func (r *RegistryManager) ListByProtocol(ctx context.Context, protocol string) ([]*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentIDs := r.byProto[protocol]
	agents := make([]*AgentInfo, 0, len(agentIDs))

	for _, id := range agentIDs {
		if agent, ok := r.agents[id]; ok {
			agents = append(agents, agent)
		}
	}

	return agents, nil
}

// ListByNamespace returns agents filtered by namespace
func (r *RegistryManager) ListByNamespace(ctx context.Context, namespace string) ([]*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentIDs := r.byNS[namespace]
	agents := make([]*AgentInfo, 0, len(agentIDs))

	for _, id := range agentIDs {
		if agent, ok := r.agents[id]; ok {
			agents = append(agents, agent)
		}
	}

	return agents, nil
}

// ListByTag returns agents filtered by tag
func (r *RegistryManager) ListByTag(ctx context.Context, tag string) ([]*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentIDs := r.byTag[tag]
	agents := make([]*AgentInfo, 0, len(agentIDs))

	for _, id := range agentIDs {
		if agent, ok := r.agents[id]; ok {
			agents = append(agents, agent)
		}
	}

	return agents, nil
}

// FindByCapabilities finds agents that have all required capabilities
func (r *RegistryManager) FindByCapabilities(ctx context.Context, required []string) ([]*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matching []*AgentInfo
	for _, agent := range r.agents {
		if r.hasCapabilities(agent.Capabilities, required) {
			matching = append(matching, agent)
		}
	}

	return matching, nil
}

// hasCapabilities checks if agent has all required capabilities
func (r *RegistryManager) hasCapabilities(agentCaps, required []string) bool {
	for _, req := range required {
		found := false
		for _, cap := range agentCaps {
			if cap == req {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Discover performs discovery and registers found agents
func (r *RegistryManager) Discover(ctx context.Context, agents []AgentInfo) error {
	for i := range agents {
		if _, err := r.Register(ctx, &agents[i]); err != nil {
			r.logger.Error("failed to register discovered agent", map[string]interface{}{
				"agent_name": agents[i].Identity.Name,
				"error":      err.Error(),
			})
		}
	}
	return nil
}

// Count returns the number of registered agents
func (r *RegistryManager) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// filterIDs removes an ID from a slice
func filterIDs(ids []string, toRemove string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != toRemove {
			result = append(result, id)
		}
	}
	return result
}
