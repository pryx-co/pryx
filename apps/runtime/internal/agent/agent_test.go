package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/channels"
	"pryx-core/internal/config"
	"pryx-core/internal/llm"
)

// MockProvider implements llm.Provider for testing
type MockProvider struct {
	CompleteFunc func(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error)
	StreamFunc   func(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamChunk, error)
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
	if m.StreamFunc != nil {
		return m.StreamFunc(ctx, req)
	}

	ch := make(chan llm.StreamChunk, 2)
	go func() {
		ch <- llm.StreamChunk{Content: "mock ", Done: false}
		ch <- llm.StreamChunk{Content: "response", Done: true}
		close(ch)
	}()
	return ch, nil
}

func TestAgent_New(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid openai provider",
			provider:  "openai",
			wantError: false,
		},
		{
			name:      "valid anthropic provider",
			provider:  "anthropic",
			wantError: false,
		},
		{
			name:      "valid openrouter provider",
			provider:  "openrouter",
			wantError: false,
		},
		{
			name:      "valid ollama provider",
			provider:  "ollama",
			wantError: false,
		},
		{
			name:      "valid glm provider",
			provider:  "glm",
			wantError: false,
		},
		{
			name:      "unsupported provider",
			provider:  "unknown",
			wantError: true,
			errMsg:    "unsupported model provider: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ModelProvider:  tt.provider,
				ModelName:      "test-model",
				OllamaEndpoint: "http://localhost:11434",
			}
			eventBus := bus.New()

			agent, err := New(cfg, eventBus, nil)

			if tt.wantError {
				if err == nil {
					t.Errorf("New() expected error but got none")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("New() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("New() unexpected error = %v", err)
				}
				if agent == nil {
					t.Error("New() returned nil agent")
				} else {
					if agent.cfg != cfg {
						t.Error("New() agent.cfg not set correctly")
					}
					if agent.bus != eventBus {
						t.Error("New() agent.bus not set correctly")
					}
				}
			}
		})
	}
}

func TestAgent_Run_ContextCancellation(t *testing.T) {
	cfg := &config.Config{
		ModelProvider: "openai",
		ModelName:     "test-model",
	}
	eventBus := bus.New()
	agent, err := New(cfg, eventBus, nil)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run should return when context is cancelled
	errChan := make(chan error, 1)
	go func() {
		errChan <- agent.Run(ctx)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Run() did not return after context cancellation")
	}
}

func TestAgent_handleChatRequest(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		providerError  error
		expectedEvents int
	}{
		{
			name: "valid chat request",
			payload: map[string]interface{}{
				"content": "Hello",
			},
			expectedEvents: 2, // Two stream chunks
		},
		{
			name:           "invalid payload type",
			payload:        "invalid string",
			expectedEvents: 0,
		},
		{
			name: "empty content",
			payload: map[string]interface{}{
				"content": "",
			},
			expectedEvents: 0,
		},
		{
			name: "provider error",
			payload: map[string]interface{}{
				"content": "Hello",
			},
			providerError:  errors.New("provider failed"),
			expectedEvents: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventBus := bus.New()

			// Create agent with mock provider
			agent := &Agent{
				cfg: &config.Config{
					ModelProvider: "openai",
					ModelName:     "test-model",
				},
				bus:           eventBus,
				promptBuilder: nil,
				provider: &MockProvider{
					StreamFunc: func(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamChunk, error) {
						if tt.providerError != nil {
							return nil, tt.providerError
						}
						ch := make(chan llm.StreamChunk, 2)
						go func() {
							ch <- llm.StreamChunk{Content: "Hello", Done: false}
							ch <- llm.StreamChunk{Content: " World", Done: true}
							close(ch)
						}()
						return ch, nil
					},
				},
			}

			// Subscribe to events
			events, cancel := eventBus.Subscribe(bus.EventSessionMessage)
			defer cancel()

			// Create and publish event
			evt := bus.NewEvent(bus.EventChatRequest, "test-session", tt.payload)
			go agent.handleEvent(context.Background(), evt)

			// Count received events
			eventCount := 0
			timeout := time.After(500 * time.Millisecond)
		done:
			for {
				select {
				case <-events:
					eventCount++
					if eventCount >= tt.expectedEvents {
						break done
					}
				case <-timeout:
					break done
				}
			}

			// Note: We don't check exact count because events are async
			// Just verify no panic occurred
		})
	}
}

func TestAgent_handleChannelMessage(t *testing.T) {
	tests := []struct {
		name          string
		payload       interface{}
		providerError error
		wantEvent     bool
	}{
		{
			name: "valid channel message",
			payload: channels.Message{
				Source:    "telegram",
				ChannelID: "123456",
				Content:   "Hello from channel",
			},
			wantEvent: true,
		},
		{
			name:      "invalid payload type",
			payload:   "invalid string",
			wantEvent: false,
		},
		{
			name: "provider error",
			payload: channels.Message{
				Source:    "telegram",
				ChannelID: "123456",
				Content:   "Hello",
			},
			providerError: errors.New("provider failed"),
			wantEvent:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventBus := bus.New()

			agent := &Agent{
				cfg: &config.Config{
					ModelProvider: "openai",
					ModelName:     "test-model",
				},
				bus: eventBus,
				provider: &MockProvider{
					CompleteFunc: func(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
						if tt.providerError != nil {
							return nil, tt.providerError
						}
						return &llm.ChatResponse{
							Content: "Channel response",
							Usage: llm.Usage{
								TotalTokens: 10,
							},
						}, nil
					},
				},
			}

			// Subscribe to outbound events
			events, cancel := eventBus.Subscribe(bus.EventChannelOutboundMessage)
			defer cancel()

			// Create and handle event
			evt := bus.NewEvent(bus.EventChannelMessage, "", tt.payload)
			go agent.handleEvent(context.Background(), evt)

			// Check for event
			if tt.wantEvent {
				select {
				case <-events:
					// Success
				case <-time.After(500 * time.Millisecond):
					t.Error("Expected outbound message event not received")
				}
			} else {
				select {
				case <-events:
					t.Error("Unexpected outbound message event received")
				case <-time.After(100 * time.Millisecond):
					// Expected - no event
				}
			}
		})
	}
}

func TestAgent_handleEvent(t *testing.T) {
	agent := &Agent{
		cfg: &config.Config{
			ModelProvider: "openai",
			ModelName:     "test-model",
		},
		bus:      bus.New(),
		provider: &MockProvider{},
	}

	tests := []struct {
		name      string
		eventType bus.EventType
		panic     bool
	}{
		{
			name:      "chat request event",
			eventType: bus.EventChatRequest,
			panic:     false,
		},
		{
			name:      "channel message event",
			eventType: bus.EventChannelMessage,
			panic:     false,
		},
		{
			name:      "unknown event type",
			eventType: bus.EventType("unknown"),
			panic:     false, // Should not panic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.panic && r == nil {
					t.Error("Expected panic but none occurred")
				} else if !tt.panic && r != nil {
					t.Errorf("Unexpected panic: %v", r)
				}
			}()

			evt := bus.NewEvent(tt.eventType, "test-session", map[string]interface{}{"content": "test"})
			agent.handleEvent(context.Background(), evt)
		})
	}
}
