package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"pryx-core/internal/agentbus"
	"pryx-core/internal/bus"
	"pryx-core/internal/channels"
	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/llm"
	"pryx-core/internal/llm/factory"
	"pryx-core/internal/mcp"
	"pryx-core/internal/memory"
	"pryx-core/internal/models"
	"pryx-core/internal/prompt"
	"pryx-core/internal/skills"
)

// Agent orchestrates the interaction between the user, LLM, and tools.
type Agent struct {
	cfg           *config.Config
	bus           *bus.Bus
	agentbus      *agentbus.Service
	provider      llm.Provider
	promptBuilder *prompt.Builder
	version       string
	skills        *skills.Registry
	mcp           *mcp.Manager
	ragMemory     *memory.RAGManager
}

// New creates a new Agent instance with the provided configuration and dependencies.
func New(cfg *config.Config, eventBus *bus.Bus, kc *keychain.Keychain, catalog *models.Catalog, skillsRegistry *skills.Registry, mcpManager *mcp.Manager, agentbusService *agentbus.Service, ragMemory *memory.RAGManager) (*Agent, error) {
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

	var provider llm.Provider
	var err error

	// Try to use catalog-aware factory if catalog is available
	if catalog != nil {
		providerFactory := factory.NewProviderFactory(catalog, kc)
		provider, err = providerFactory.CreateProvider(cfg.ModelProvider, cfg.ModelName, apiKey)
		if err != nil {
			log.Printf("Warning: Failed to create provider from catalog: %v, using fallback", err)
			// Fallback to low-level factory
			provider, err = factory.NewProvider(cfg.ModelProvider, apiKey, baseURL)
			if err != nil {
				return nil, fmt.Errorf("failed to create LLM provider: %w", err)
			}
		}
	} else {
		// Fallback to low-level factory without catalog
		provider, err = factory.NewProvider(cfg.ModelProvider, apiKey, baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create LLM provider: %w", err)
		}
	}

	promptBuilder := prompt.NewBuilder(prompt.DefaultPryxDir(), prompt.ModeFull)
	if err := promptBuilder.EnsureTemplates(); err != nil {
		log.Printf("Warning: Failed to ensure prompt templates: %v", err)
	}

	return &Agent{
		cfg:           cfg,
		bus:           eventBus,
		agentbus:      agentbusService,
		provider:      provider,
		promptBuilder: promptBuilder,
		version:       "dev",
		skills:        skillsRegistry,
		mcp:           mcpManager,
		ragMemory:     ragMemory,
	}, nil
}

// Run starts the agent's main event loop, listening for chat requests and channel messages.
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
			// Handle event in goroutine with panic recovery
			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Agent: Recovered from panic in event handler: %v", r)
						a.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, evt.SessionID, map[string]interface{}{
							"kind":  "agent.handler.panic",
							"error": fmt.Sprintf("%v", r),
							"event": evt.Event,
						}))
					}
				}()
				a.handleEvent(ctx, evt)
			}()
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
	// Panic recovery at handler level
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Agent: Recovered from panic in handleChatRequest: %v", r)
			a.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, evt.SessionID, map[string]interface{}{
				"kind":  "agent.chat_request.panic",
				"error": fmt.Sprintf("%v", r),
			}))
		}
	}()

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
		a.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, sessionID, map[string]interface{}{
			"kind":  "agent.llm_error",
			"error": err.Error(),
		}))
		return
	}

	var fullResponse strings.Builder
	for chunk := range stream {
		if chunk.Err != nil {
			log.Printf("Agent: Stream error: %v", chunk.Err)
			a.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, sessionID, map[string]interface{}{
				"kind":  "agent.stream_error",
				"error": chunk.Err.Error(),
			}))
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
	// Panic recovery at handler level
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Agent: Recovered from panic in handleChannelMessage: %v", r)
			a.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, evt.SessionID, map[string]interface{}{
				"kind":  "agent.channel_message.panic",
				"error": fmt.Sprintf("%v", r),
			}))
		}
	}()

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
		a.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, "", map[string]interface{}{
			"kind":  "agent.channel.llm_error",
			"error": err.Error(),
		}))
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

	memoryContext := ""
	if a.cfg.MemoryEnabled && a.ragMemory != nil && a.ragMemory.Enabled() {
		if a.cfg.MemoryAutoFlush && a.ragMemory.AutoFlush() != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			memContext, _ := a.ragMemory.AutoFlush().GetMemoryContextForAgent("current context", 5)
			if memContext != "" {
				memoryContext = memContext
			}
			_ = ctx
		}
	}

	metadata := prompt.Metadata{
		CurrentTime:     time.Now(),
		Version:         a.version,
		SessionID:       sessionID,
		AvailableTools:  a.getAvailableTools(),
		AvailableSkills: a.getAvailableSkills(),
		MemoryContext:   memoryContext,
	}

	return a.promptBuilder.Build(metadata)
}

func (a *Agent) getAvailableTools() []string {
	if a.mcp == nil {
		return []string{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tools, err := a.mcp.ListToolsFlat(ctx, false)
	if err != nil {
		log.Printf("Agent: Failed to list MCP tools: %v", err)
		return []string{}
	}

	var result []string
	for _, tool := range tools {
		result = append(result, tool.Name)
	}
	return result
}

func (a *Agent) getAvailableSkills() []string {
	if a.skills == nil {
		return []string{}
	}

	var result []string
	for _, skill := range a.skills.List() {
		result = append(result, skill.ID)
	}
	return result
}
