package spawn

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pryx-core/internal/bus"
)

// SpawnTool provides the sessions_spawn functionality for LLM tool calling
type SpawnTool struct {
	spawner *Spawner
	bus     *bus.Bus
}

// NewSpawnTool creates a new spawn tool instance
func NewSpawnTool(spawner *Spawner, bus *bus.Bus) *SpawnTool {
	return &SpawnTool{
		spawner: spawner,
		bus:     bus,
	}
}

// Name returns the tool name
func (t *SpawnTool) Name() string {
	return "sessions_spawn"
}

// Description returns the tool description for LLM
func (t *SpawnTool) Description() string {
	return `Spawn a sub-agent to handle a specific task concurrently.

Use this when:
1. A task can be broken into independent sub-tasks
2. You need to research multiple topics in parallel
3. You want to delegate work to specialized agents
4. A task might take a long time and you want to continue other work

The sub-agent will run in parallel and return its result. You can check status and retrieve results later.`
}

// Schema returns the JSON schema for the tool parameters
func (t *SpawnTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task": map[string]interface{}{
				"type":        "string",
				"description": "The specific task for the sub-agent to complete",
			},
			"context": map[string]interface{}{
				"type":        "string",
				"description": "System context/persona for the sub-agent (e.g., 'You are a code reviewer focusing on security')",
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional: session ID to associate with this sub-agent (defaults to current session)",
			},
		},
		"required": []string{"task"},
	}
}

// Execute runs the spawn tool
func (t *SpawnTool) Execute(ctx context.Context, params json.RawMessage, parentID string) (interface{}, error) {
	var args struct {
		Task      string `json:"task"`
		Context   string `json:"context"`
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.Task == "" {
		return nil, fmt.Errorf("task is required")
	}

	// Use provided session ID or generate one
	sessionID := args.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	// Default context if not provided
	systemContext := args.Context
	if systemContext == "" {
		systemContext = "You are a helpful AI assistant working on a specific sub-task."
	}

	// Spawn the sub-agent
	agent, err := t.spawner.Spawn(ctx, parentID, sessionID, args.Task, systemContext)
	if err != nil {
		return nil, fmt.Errorf("failed to spawn agent: %w", err)
	}

	// Wait for completion with timeout
	result, err := t.waitForResult(ctx, agent.ID, 5*time.Minute)
	if err != nil {
		return map[string]interface{}{
			"agent_id":   agent.ID,
			"status":     agent.Status,
			"session_id": sessionID,
			"message":    "Agent spawned but did not complete in time. Check status later.",
		}, nil
	}

	return map[string]interface{}{
		"agent_id":    result.AgentID,
		"status":      result.Status,
		"output":      result.Output,
		"error":       result.Error,
		"tokens_used": result.TokenUsed,
		"duration_ms": result.Duration.Milliseconds(),
	}, nil
}

// waitForResult waits for a sub-agent to complete
func (t *SpawnTool) waitForResult(ctx context.Context, agentID string, timeout time.Duration) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			agent, ok := t.spawner.Get(agentID)
			if !ok {
				return nil, fmt.Errorf("agent not found")
			}

			switch agent.Status {
			case StatusCompleted, StatusFailed, StatusCancelled:
				return &Result{
					AgentID: agentID,
					Status:  agent.Status,
				}, nil
			}
		}
	}
}

// GetAgentStatus returns the status of a specific agent
func (t *SpawnTool) GetAgentStatus(agentID string) (map[string]interface{}, error) {
	agent, ok := t.spawner.Get(agentID)
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return map[string]interface{}{
		"agent_id":   agent.ID,
		"status":     agent.Status,
		"session_id": agent.SessionID,
		"parent_id":  agent.ParentID,
		"created_at": agent.CreatedAt,
		"token_used": agent.tokenUsed,
	}, nil
}

// ListAgents returns all active agents
func (t *SpawnTool) ListAgents() []map[string]interface{} {
	agents := t.spawner.List()
	result := make([]map[string]interface{}, len(agents))

	for i, agent := range agents {
		result[i] = map[string]interface{}{
			"agent_id":   agent.ID,
			"status":     agent.Status,
			"session_id": agent.SessionID,
			"parent_id":  agent.ParentID,
		}
	}

	return result
}

// ForkSession creates a fork of the current session
func (t *SpawnTool) ForkSession(sourceSessionID string) (string, error) {
	return t.spawner.Fork(context.Background(), sourceSessionID)
}

// RegisterHandlers registers bus event handlers for spawn-related events
func (t *SpawnTool) RegisterHandlers() {
	// Subscribe to spawn requests
	events, _ := t.bus.Subscribe(bus.EventTraceEvent)
	go func() {
		for evt := range events {
			if evt.Event == bus.EventTraceEvent {
				// Handle trace events related to sub-agents
				// This could be used for monitoring, logging, etc.
			}
		}
	}()
}
