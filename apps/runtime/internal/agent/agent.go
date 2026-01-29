package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/channels"
	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/llm"
	"pryx-core/internal/llm/factory"
	"pryx-core/internal/prompt"
)

type Agent struct {
	cfg           *config.Config
	bus           *bus.Bus
	provider      llm.Provider
	promptBuilder *prompt.Builder
	version       string
}

func New(cfg *config.Config, eventBus *bus.Bus, kc *keychain.Keychain) (*Agent, error) {
	var apiKey string
	var baseURL string

	switch strings.ToLower(cfg.ModelProvider) {
	case "openai", "anthropic", "openrouter", "together", "groq", "xai", "mistral", "cohere", "google", "glm":
		if kc != nil {
			if key, err := kc.GetProviderKey(cfg.ModelProvider); err == nil {
				apiKey = key
			}
		}
	case "ollama":
		baseURL = cfg.OllamaEndpoint
	default:
		return nil, fmt.Errorf("unsupported model provider: %s", cfg.ModelProvider)
	}

	provider, err := factory.NewProvider(cfg.ModelProvider, apiKey, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	promptBuilder := prompt.NewBuilder(prompt.DefaultPryxDir(), prompt.ModeFull)
	if err := promptBuilder.EnsureTemplates(); err != nil {
		log.Printf("Warning: Failed to ensure prompt templates: %v", err)
	}

	return &Agent{
		cfg:           cfg,
		bus:           eventBus,
		provider:      provider,
		promptBuilder: promptBuilder,
		version:       "dev",
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

	systemPrompt, err := a.buildSystemPrompt(sessionID)
	if err != nil {
		log.Printf("Agent: Failed to build system prompt: %v", err)
		systemPrompt = "You are Pryx, a helpful AI assistant."
	}

	req := llm.ChatRequest{
		Model: a.cfg.ModelName,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: systemPrompt},
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
	msg, ok := evt.Payload.(channels.Message)
	if !ok {
		log.Println("Agent: Invalid channel message payload")
		return
	}

	log.Printf("Agent: Processing channel message from %s (chat: %s): %s", msg.Source, msg.ChannelID, msg.Content)

	systemPrompt, err := a.buildSystemPrompt("")
	if err != nil {
		log.Printf("Agent: Failed to build system prompt: %v", err)
		systemPrompt = "You are Pryx, a helpful AI assistant."
	}

	req := llm.ChatRequest{
		Model: a.cfg.ModelName,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: systemPrompt},
			{Role: llm.RoleUser, Content: msg.Content},
		},
		Stream: false,
	}

	resp, err := a.provider.Complete(ctx, req)
	if err != nil {
		log.Printf("Agent: LLM error: %v", err)
		return
	}

	log.Printf("Agent: Sending channel response (%d chars)", len(resp.Content))

	a.bus.Publish(bus.NewEvent(bus.EventChannelOutboundMessage, "", map[string]interface{}{
		"source":     msg.Source,
		"channel_id": msg.ChannelID,
		"content":    resp.Content,
	}))
}

func (a *Agent) buildSystemPrompt(sessionID string) (string, error) {
	if a.promptBuilder == nil {
		return "You are Pryx, a helpful AI assistant.", nil
	}

	metadata := prompt.Metadata{
		CurrentTime:     time.Now(),
		Version:         a.version,
		SessionID:       sessionID,
		AvailableTools:  a.getAvailableTools(),
		AvailableSkills: a.getAvailableSkills(),
	}

	return a.promptBuilder.Build(metadata)
}

func (a *Agent) getAvailableTools() []string {
	return []string{
		"filesystem",
		"shell",
		"browser",
		"clipboard",
	}
}

func (a *Agent) getAvailableSkills() []string {
	return []string{}
}
