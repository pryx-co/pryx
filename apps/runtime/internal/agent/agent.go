package agent

import (
	"context"
	"fmt"
	"log"
	"strings"

	"pryx-core/internal/bus"
	"pryx-core/internal/channels"
	"pryx-core/internal/config"
	"pryx-core/internal/llm"
	"pryx-core/internal/llm/factory"
)

type Agent struct {
	cfg      *config.Config
	bus      *bus.Bus
	provider llm.Provider
}

func New(cfg *config.Config, eventBus *bus.Bus) (*Agent, error) {
	// Initialize LLM Provider based on Config
	var providerType factory.ProviderType
	var apiKey string
	var baseURL string

	switch strings.ToLower(cfg.ModelProvider) {
	case "openai":
		providerType = factory.ProviderOpenAI
		apiKey = cfg.OpenAIKey
	case "anthropic":
		providerType = factory.ProviderAnthropic
		apiKey = cfg.AnthropicKey
	case "openrouter":
		providerType = factory.ProviderOpenRouter
		apiKey = cfg.OpenAIKey // OpenRouter uses OpenAI key config
	case "ollama":
		providerType = factory.ProviderOllama
		baseURL = cfg.OllamaEndpoint
	case "glm":
		providerType = factory.ProviderGLM
		apiKey = cfg.GLMKey
	default:
		return nil, fmt.Errorf("unsupported model provider: %s", cfg.ModelProvider)
	}

	provider, err := factory.NewProvider(providerType, apiKey, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	return &Agent{
		cfg:      cfg,
		bus:      eventBus,
		provider: provider,
	}, nil
}

func (a *Agent) Run(ctx context.Context) error {
	// Subscribe to incoming messages
	events, cancel := a.bus.Subscribe(bus.EventChatRequest, bus.EventChannelMessage)
	defer cancel()

	log.Println("Agent: Started listening for messages...")

	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-events:
			if !ok {
				return nil
			}
			go a.handleEvent(ctx, evt)
		}
	}
}

func (a *Agent) handleEvent(ctx context.Context, evt bus.Event) {
	switch evt.Event {
	case bus.EventChatRequest:
		a.handleChatRequest(ctx, evt)
	case bus.EventChannelMessage:
		a.handleChannelMessage(ctx, evt)
	}
}

func (a *Agent) handleChatRequest(ctx context.Context, evt bus.Event) {
	// Parse TUI chat request
	payload, ok := evt.Payload.(map[string]interface{})
	if !ok {
		log.Println("Agent: Invalid chat request payload")
		return
	}

	content, _ := payload["content"].(string)
	sessionID := evt.SessionID

	if content == "" {
		return
	}

	log.Printf("Agent: Processing TUI message: %s (session: %s)", content, sessionID)

	// Build request
	req := llm.ChatRequest{
		Model: a.cfg.ModelName,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "You are a helpful AI assistant."},
			{Role: llm.RoleUser, Content: content},
		},
		Stream: true,
	}

	// Stream response
	stream, err := a.provider.Stream(ctx, req)
	if err != nil {
		log.Printf("Agent: LLM error: %v", err)
		return
	}

	var fullResponse strings.Builder
	for chunk := range stream {
		if chunk.Err != nil {
			log.Printf("Agent: Stream error: %v", chunk.Err)
			break
		}
		fullResponse.WriteString(chunk.Content)

		// Publish delta to TUI
		a.bus.Publish(bus.NewEvent(bus.EventSessionMessage, sessionID, map[string]interface{}{
			"content": chunk.Content,
			"done":    chunk.Done,
		}))

		if chunk.Done {
			break
		}
	}

	log.Printf("Agent: Completed TUI response (%d chars)", fullResponse.Len())
}

func (a *Agent) handleChannelMessage(ctx context.Context, evt bus.Event) {
	// Parse channel message
	msg, ok := evt.Payload.(channels.Message)
	if !ok {
		log.Println("Agent: Invalid channel message payload")
		return
	}

	log.Printf("Agent: Processing channel message from %s (chat: %s): %s", msg.Source, msg.ChannelID, msg.Content)

	// Build request
	req := llm.ChatRequest{
		Model: a.cfg.ModelName,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "You are a helpful AI assistant."},
			{Role: llm.RoleUser, Content: msg.Content},
		},
		Stream: false, // Use non-streaming for channels for simplicity
	}

	// Get response
	resp, err := a.provider.Complete(ctx, req)
	if err != nil {
		log.Printf("Agent: LLM error: %v", err)
		return
	}

	log.Printf("Agent: Sending channel response (%d chars)", len(resp.Content))

	// Publish outbound message
	a.bus.Publish(bus.NewEvent(bus.EventChannelOutboundMessage, "", map[string]interface{}{
		"source":     msg.Source,    // Route back to same channel instance
		"channel_id": msg.ChannelID, // Reply to same chat
		"content":    resp.Content,
	}))
}
