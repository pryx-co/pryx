package spawn

import (
	"context"
	"errors"
	"testing"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/llm"
	"pryx-core/internal/store"
)

// MockProvider implements llm.Provider for testing
type MockProvider struct {
	CompleteFunc func(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error)
}

func (m *MockProvider) Complete(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.CompleteFunc != nil {
		return m.CompleteFunc(ctx, req)
	}
	return &llm.ChatResponse{
		Content: "mock response",
		Usage: llm.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func (m *MockProvider) Stream(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: "mock", Done: true}
	close(ch)
	return ch, nil
}

func TestNewSpawner(t *testing.T) {
	cfg := &config.Config{
		ModelProvider: "openai",
	}
	eventBus := bus.New()
	kc := keychain.New("test")
	s, _ := store.New(":memory:")

	spawner := NewSpawner(cfg, eventBus, kc, s)

	if spawner == nil {
		t.Fatal("NewSpawner() returned nil")
	}

	if spawner.cfg != cfg {
		t.Error("NewSpawner() spawner.cfg not set correctly")
	}

	if spawner.bus != eventBus {
		t.Error("NewSpawner() spawner.bus not set correctly")
	}

	if spawner.agents == nil {
		t.Error("NewSpawner() spawner.agents not initialized")
	}

	if spawner.maxAgents != 10 {
		t.Errorf("NewSpawner() maxAgents = %d, want 10", spawner.maxAgents)
	}
}

func TestSpawner_Spawn(t *testing.T) {
	tests := []struct {
		name          string
		maxAgents     int
		initialCount  int
		wantError     bool
		errMsg        string
		providerError bool
	}{
		{
			name:         "successful spawn",
			maxAgents:    10,
			initialCount: 0,
			wantError:    false,
		},
		{
			name:         "max agents reached",
			maxAgents:    2,
			initialCount: 2,
			wantError:    true,
			errMsg:       "max sub-agents reached (2)",
		},
		{
			name:          "provider creation error",
			maxAgents:     10,
			initialCount:  0,
			wantError:     true,
			providerError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ModelProvider: "openai",
			}

			if tt.providerError {
				cfg.ModelProvider = "unsupported"
			}

			eventBus := bus.New()
			kc := keychain.New("test")
			s, _ := store.New(":memory:")
			spawner := NewSpawner(cfg, eventBus, kc, s)
			spawner.maxAgents = tt.maxAgents

			// Pre-populate agents if needed
			for i := 0; i < tt.initialCount; i++ {
				spawner.agents[generateAgentID()] = &SubAgent{
					ID:     generateAgentID(),
					Status: StatusRunning,
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			agent, err := spawner.Spawn(ctx, "parent-1", "session-1", "test task", "test context")

			if tt.wantError {
				if err == nil {
					t.Error("Spawn() expected error but got none")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Spawn() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Spawn() unexpected error = %v", err)
				}
				if agent == nil {
					t.Error("Spawn() returned nil agent")
				} else {
					if agent.ParentID != "parent-1" {
						t.Errorf("Spawn() agent.ParentID = %v, want parent-1", agent.ParentID)
					}
					if agent.SessionID != "session-1" {
						t.Errorf("Spawn() agent.SessionID = %v, want session-1", agent.SessionID)
					}
					if agent.SystemCtx != "test context" {
						t.Errorf("Spawn() agent.SystemCtx = %v, want test context", agent.SystemCtx)
					}
					// Status may be pending or running depending on timing
					agent.mu.RLock()
					status := agent.Status
					agent.mu.RUnlock()
					if status != StatusPending && status != StatusRunning {
						t.Errorf("Spawn() agent.Status = %v, want pending or running", status)
					}
					if agent.ID == "" {
						t.Error("Spawn() agent.ID is empty")
					}

					// Cancel the agent to clean up
					agent.cancel()
				}
			}
		})
	}
}

func TestSpawner_Get(t *testing.T) {
	cfg := &config.Config{
		ModelProvider: "openai",
	}
	eventBus := bus.New()
	kc := keychain.New("test")
	s, _ := store.New(":memory:")
	spawner := NewSpawner(cfg, eventBus, kc, s)

	// Test getting non-existent agent
	agent, ok := spawner.Get("non-existent")
	if ok {
		t.Error("Get() returned ok=true for non-existent agent")
	}
	if agent != nil {
		t.Error("Get() returned non-nil agent for non-existent id")
	}

	// Add an agent and test getting it
	testAgent := &SubAgent{
		ID:     "test-agent-1",
		Status: StatusRunning,
	}
	spawner.agents["test-agent-1"] = testAgent

	agent, ok = spawner.Get("test-agent-1")
	if !ok {
		t.Error("Get() returned ok=false for existing agent")
	}
	if agent != testAgent {
		t.Error("Get() returned wrong agent")
	}
}

func TestSpawner_List(t *testing.T) {
	cfg := &config.Config{
		ModelProvider: "openai",
	}
	eventBus := bus.New()
	kc := keychain.New("test")
	s, _ := store.New(":memory:")
	spawner := NewSpawner(cfg, eventBus, kc, s)

	// Test empty list
	list := spawner.List()
	if len(list) != 0 {
		t.Errorf("List() returned %d agents, want 0", len(list))
	}

	// Add some agents
	agent1 := &SubAgent{ID: "agent-1", Status: StatusRunning}
	agent2 := &SubAgent{ID: "agent-2", Status: StatusPending}
	spawner.agents["agent-1"] = agent1
	spawner.agents["agent-2"] = agent2

	list = spawner.List()
	if len(list) != 2 {
		t.Errorf("List() returned %d agents, want 2", len(list))
	}

	// Verify agents are in list
	found := make(map[string]bool)
	for _, a := range list {
		found[a.ID] = true
	}
	if !found["agent-1"] || !found["agent-2"] {
		t.Error("List() missing expected agents")
	}
}

func TestSpawner_Cancel(t *testing.T) {
	cfg := &config.Config{
		ModelProvider: "openai",
	}
	eventBus := bus.New()
	kc := keychain.New("test")
	s, _ := store.New(":memory:")
	spawner := NewSpawner(cfg, eventBus, kc, s)

	// Test cancel non-existent agent
	err := spawner.Cancel("non-existent")
	if err == nil {
		t.Error("Cancel() expected error for non-existent agent")
	}

	// Add and cancel an agent
	ctx, cancel := context.WithCancel(context.Background())
	agent := &SubAgent{
		ID:     "test-agent",
		Status: StatusRunning,
		cancel: cancel,
	}
	spawner.agents["test-agent"] = agent

	err = spawner.Cancel("test-agent")
	if err != nil {
		t.Errorf("Cancel() unexpected error = %v", err)
	}

	if agent.Status != StatusCancelled {
		t.Errorf("Cancel() agent.Status = %v, want cancelled", agent.Status)
	}

	// Verify context was cancelled
	select {
	case <-ctx.Done():
		// Success - context was cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("Cancel() did not cancel agent context")
	}
}

func TestSpawner_Cleanup(t *testing.T) {
	cfg := &config.Config{
		ModelProvider: "openai",
	}
	eventBus := bus.New()
	kc := keychain.New("test")
	s, _ := store.New(":memory:")
	spawner := NewSpawner(cfg, eventBus, kc, s)

	now := time.Now()

	// Add agents with different statuses and ages
	spawner.agents["running"] = &SubAgent{
		ID:        "running",
		Status:    StatusRunning,
		CreatedAt: now.Add(-2 * time.Hour),
	}
	spawner.agents["completed-old"] = &SubAgent{
		ID:        "completed-old",
		Status:    StatusCompleted,
		CreatedAt: now.Add(-2 * time.Hour),
	}
	spawner.agents["completed-new"] = &SubAgent{
		ID:        "completed-new",
		Status:    StatusCompleted,
		CreatedAt: now.Add(-30 * time.Minute),
	}
	spawner.agents["failed-old"] = &SubAgent{
		ID:        "failed-old",
		Status:    StatusFailed,
		CreatedAt: now.Add(-2 * time.Hour),
	}
	spawner.agents["cancelled-old"] = &SubAgent{
		ID:        "cancelled-old",
		Status:    StatusCancelled,
		CreatedAt: now.Add(-2 * time.Hour),
	}

	spawner.Cleanup(1 * time.Hour)

	// Verify old completed/failed/cancelled agents were removed
	if _, ok := spawner.agents["completed-old"]; ok {
		t.Error("Cleanup() did not remove old completed agent")
	}
	if _, ok := spawner.agents["failed-old"]; ok {
		t.Error("Cleanup() did not remove old failed agent")
	}
	if _, ok := spawner.agents["cancelled-old"]; ok {
		t.Error("Cleanup() did not remove old cancelled agent")
	}

	// Verify running agent still exists
	if _, ok := spawner.agents["running"]; !ok {
		t.Error("Cleanup() removed running agent")
	}

	// Verify new completed agent still exists
	if _, ok := spawner.agents["completed-new"]; !ok {
		t.Error("Cleanup() removed new completed agent")
	}
}

func TestSpawner_Fork(t *testing.T) {
	cfg := &config.Config{
		ModelProvider: "openai",
	}
	eventBus := bus.New()
	kc := keychain.New("test")
	s, _ := store.New(":memory:")
	spawner := NewSpawner(cfg, eventBus, kc, s)

	ctx := context.Background()

	// Create the source session first - note: CreateSession generates a UUID, not using the title as ID
	sourceSession, err := s.CreateSession("test source session")
	if err != nil {
		t.Fatalf("Failed to create source session: %v", err)
	}

	// Fork using the actual session ID (not the title)
	newSessionID, err := spawner.Fork(ctx, sourceSession.ID)

	if err != nil {
		t.Errorf("Fork() unexpected error = %v", err)
	}

	if newSessionID == "" {
		t.Error("Fork() returned empty session ID")
	}

	if newSessionID == sourceSession.ID {
		t.Error("Fork() returned same session ID as source")
	}

	// Verify session ID format
	if len(newSessionID) < 10 {
		t.Errorf("Fork() returned suspiciously short session ID: %s", newSessionID)
	}
}

func TestSubAgent_run(t *testing.T) {
	tests := []struct {
		name           string
		providerError  error
		expectedStatus Status
		expectOutput   bool
	}{
		{
			name:           "successful execution",
			providerError:  nil,
			expectedStatus: StatusCompleted,
			expectOutput:   true,
		},
		{
			name:           "provider error",
			providerError:  errors.New("provider failed"),
			expectedStatus: StatusFailed,
			expectOutput:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventBus := bus.New()

			// Subscribe to trace events
			events, cancel := eventBus.Subscribe(bus.EventTraceEvent)
			defer cancel()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			agentCtx, agentCancel := context.WithCancel(ctx)
			defer agentCancel()

			agent := &SubAgent{
				ID:        "test-agent",
				SessionID: "test-session",
				ParentID:  "parent-1",
				SystemCtx: "test system context",
				Status:    StatusPending,
				cancel:    agentCancel,
				eventCh:   make(chan bus.Event, 10),
				bus:       eventBus,
				provider: &MockProvider{
					CompleteFunc: func(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
						if tt.providerError != nil {
							return nil, tt.providerError
						}
						return &llm.ChatResponse{
							Content: "test output",
							Usage: llm.Usage{
								TotalTokens: 20,
							},
						}, nil
					},
				},
				maxTokens: 100000,
				maxTools:  10,
			}

			// Run the agent
			go agent.run(agentCtx, "test task")

			// Wait for completion
			time.Sleep(200 * time.Millisecond)

			agent.mu.RLock()
			status := agent.Status
			agent.mu.RUnlock()
			if status != tt.expectedStatus {
				t.Errorf("run() agent.Status = %v, want %v", status, tt.expectedStatus)
			}

			// Check events
			eventCount := 0
			timeout := time.After(500 * time.Millisecond)
		eventLoop:
			for {
				select {
				case <-events:
					eventCount++
				case <-timeout:
					break eventLoop
				}
			}

			// We expect at least start and completion events
			if eventCount < 1 {
				t.Error("run() did not publish expected events")
			}
		})
	}
}

func TestSubAgent_buildPrompt(t *testing.T) {
	agent := &SubAgent{
		SystemCtx: "test system context",
	}

	task := "do something important"
	prompt := agent.buildPrompt(task)

	if prompt == "" {
		t.Error("buildPrompt() returned empty string")
	}

	if !contains(prompt, task) {
		t.Errorf("buildPrompt() does not contain task: %s", task)
	}

	if !contains(prompt, "specialized sub-agent") {
		t.Error("buildPrompt() does not contain expected guidelines")
	}
}

func TestGenerateAgentID(t *testing.T) {
	id1 := generateAgentID()
	id2 := generateAgentID()

	if id1 == "" {
		t.Error("generateAgentID() returned empty string")
	}

	if id1 == id2 {
		t.Error("generateAgentID() returned duplicate IDs")
	}

	if !contains(id1, "agent-") {
		t.Errorf("generateAgentID() = %s, want prefix 'agent-'", id1)
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	if id1 == "" {
		t.Error("generateSessionID() returned empty string")
	}

	if id1 == id2 {
		t.Error("generateSessionID() returned duplicate IDs")
	}

	if !contains(id1, "session-") {
		t.Errorf("generateSessionID() = %s, want prefix 'session-'", id1)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
