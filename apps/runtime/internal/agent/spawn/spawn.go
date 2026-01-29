package spawn

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/llm"
	"pryx-core/internal/llm/factory"
)

// SubAgent represents a spawned child agent running in its own goroutine
type SubAgent struct {
	ID        string
	SessionID string
	ParentID  string
	SystemCtx string
	Status    Status
	CreatedAt time.Time
	cancel    context.CancelFunc
	eventCh   chan bus.Event
	bus       *bus.Bus
	provider  llm.Provider
	maxTokens int
	maxTools  int
	toolCount int
	tokenUsed int
}

// Status represents the state of a sub-agent
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// Result contains the output of a sub-agent execution
type Result struct {
	AgentID   string
	Status    Status
	Output    string
	Error     string
	TokenUsed int
	ToolUsed  int
	Duration  time.Duration
}

// Spawner manages the lifecycle of sub-agents
type Spawner struct {
	cfg       *config.Config
	bus       *bus.Bus
	keychain  *keychain.Keychain
	mu        sync.RWMutex
	agents    map[string]*SubAgent
	maxAgents int
}

// NewSpawner creates a new agent spawner
func NewSpawner(cfg *config.Config, b *bus.Bus, kc *keychain.Keychain) *Spawner {
	return &Spawner{
		cfg:       cfg,
		bus:       b,
		keychain:  kc,
		agents:    make(map[string]*SubAgent),
		maxAgents: 10,
	}
}

// Spawn creates a new sub-agent with the given task
func (s *Spawner) Spawn(ctx context.Context, parentID, sessionID, task, systemContext string) (*SubAgent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.agents) >= s.maxAgents {
		return nil, fmt.Errorf("max sub-agents reached (%d)", s.maxAgents)
	}

	agentID := generateAgentID()

	provider, err := s.createProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	agentCtx, cancel := context.WithCancel(ctx)

	agent := &SubAgent{
		ID:        agentID,
		SessionID: sessionID,
		ParentID:  parentID,
		SystemCtx: systemContext,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		cancel:    cancel,
		eventCh:   make(chan bus.Event, 100),
		bus:       s.bus,
		provider:  provider,
		maxTokens: 100000,
		maxTools:  10,
	}

	s.agents[agentID] = agent

	// Start the agent in a goroutine
	go agent.run(agentCtx, task)

	log.Printf("Spawned sub-agent %s (parent: %s, session: %s)", agentID, parentID, sessionID)
	return agent, nil
}

// Fork creates a fork of an existing session (like OpenCode's session.fork)
func (s *Spawner) Fork(ctx context.Context, sourceSessionID string) (string, error) {
	// Generate new session ID
	newSessionID := generateSessionID()

	// TODO: Copy session state from source to new session
	// For now, we just create a new empty session reference

	log.Printf("Forked session %s to %s", sourceSessionID, newSessionID)
	return newSessionID, nil
}

// Get retrieves a sub-agent by ID
func (s *Spawner) Get(agentID string) (*SubAgent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	agent, ok := s.agents[agentID]
	return agent, ok
}

// List returns all active sub-agents
func (s *Spawner) List() []*SubAgent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]*SubAgent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}
	return agents
}

// Cancel stops a running sub-agent
func (s *Spawner) Cancel(agentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, ok := s.agents[agentID]
	if !ok {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	agent.cancel()
	agent.Status = StatusCancelled
	return nil
}

// Cleanup removes completed agents (call periodically)
func (s *Spawner) Cleanup(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for id, agent := range s.agents {
		if agent.Status == StatusCompleted || agent.Status == StatusFailed || agent.Status == StatusCancelled {
			if agent.CreatedAt.Before(cutoff) {
				delete(s.agents, id)
				log.Printf("Cleaned up sub-agent %s", id)
			}
		}
	}
}

// run executes the sub-agent's task
func (a *SubAgent) run(ctx context.Context, task string) {
	a.Status = StatusRunning
	startTime := time.Now()

	// Publish start event
	a.bus.Publish(bus.NewEvent(bus.EventTraceEvent, a.SessionID, map[string]interface{}{
		"kind":     "subagent.started",
		"agent_id": a.ID,
		"parent":   a.ParentID,
		"task":     task,
	}))

	// Build prompt with context
	prompt := a.buildPrompt(task)

	// Execute LLM request
	req := llm.ChatRequest{
		Model: "llama3",
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: a.SystemCtx},
			{Role: llm.RoleUser, Content: prompt},
		},
		Stream: false,
	}

	resp, err := a.provider.Complete(ctx, req)

	duration := time.Since(startTime)

	if err != nil {
		a.Status = StatusFailed
		a.publishResult(Result{
			AgentID:  a.ID,
			Status:   StatusFailed,
			Error:    err.Error(),
			Duration: duration,
		})
		return
	}

	// Update usage stats
	a.tokenUsed = resp.Usage.TotalTokens
	a.Status = StatusCompleted

	// Publish completion
	a.publishResult(Result{
		AgentID:   a.ID,
		Status:    StatusCompleted,
		Output:    resp.Content,
		TokenUsed: resp.Usage.TotalTokens,
		Duration:  duration,
	})
}

func (a *SubAgent) buildPrompt(task string) string {
	return fmt.Sprintf(`You are a specialized sub-agent working on a specific task.

Your task: %s

Guidelines:
- Focus only on the assigned task
- You have access to tools if needed (use sparingly)
- Return a clear, concise result
- If you encounter issues, report them clearly

Provide your response:`, task)
}

func (a *SubAgent) publishResult(result Result) {
	a.bus.Publish(bus.NewEvent(bus.EventTraceEvent, a.SessionID, map[string]interface{}{
		"kind":     "subagent.completed",
		"agent_id": result.AgentID,
		"status":   result.Status,
		"output":   result.Output,
		"error":    result.Error,
		"tokens":   result.TokenUsed,
		"duration": result.Duration.Milliseconds(),
	}))

	// Also notify parent session
	if a.ParentID != "" {
		a.bus.Publish(bus.NewEvent(bus.EventSessionMessage, a.ParentID, map[string]interface{}{
			"role":     "subagent",
			"agent_id": a.ID,
			"status":   result.Status,
			"content":  result.Output,
		}))
	}
}

func (s *Spawner) createProvider() (llm.Provider, error) {
	var apiKey string
	var baseURL string

	switch s.cfg.ModelProvider {
	case "openai", "anthropic", "openrouter", "together", "groq", "xai", "mistral", "cohere", "google", "glm":
		if s.keychain != nil {
			if key, err := s.keychain.GetProviderKey(s.cfg.ModelProvider); err == nil {
				apiKey = key
			}
		}
	case "ollama":
		baseURL = s.cfg.OllamaEndpoint
	default:
		return nil, fmt.Errorf("unsupported provider: %s", s.cfg.ModelProvider)
	}

	return factory.NewProvider(s.cfg.ModelProvider, apiKey, baseURL)
}

func generateAgentID() string {
	return fmt.Sprintf("agent-%d", time.Now().UnixNano())
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}
