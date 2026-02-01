package social

import (
	"context"
	"fmt"
	"sync"
)

// Hub manages social adapters and provides unified access to social features
type Hub struct {
	adapters map[string]SocialAdapter
	mu       sync.RWMutex
}

// NewHub creates a new social hub
func NewHub() *Hub {
	return &Hub{
		adapters: make(map[string]SocialAdapter),
	}
}

// RegisterAdapter registers a social adapter
func (h *Hub) RegisterAdapter(adapter SocialAdapter) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if adapter == nil {
		return fmt.Errorf("cannot register nil adapter")
	}

	name := adapter.Name()
	if name == "" {
		return fmt.Errorf("adapter has empty name")
	}

	if _, exists := h.adapters[name]; exists {
		return fmt.Errorf("adapter '%s' already registered", name)
	}

	h.adapters[name] = adapter
	return nil
}

// UnregisterAdapter removes an adapter from the hub
func (h *Hub) UnregisterAdapter(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.adapters, name)
}

// GetAdapter retrieves an adapter by name
func (h *Hub) GetAdapter(name string) (SocialAdapter, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	adapter, ok := h.adapters[name]
	return adapter, ok
}

// ListAdapters returns all registered adapter names
func (h *Hub) ListAdapters() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	names := make([]string, 0, len(h.adapters))
	for name := range h.adapters {
		names = append(names, name)
	}
	return names
}

// Capabilities returns the capabilities of an adapter
func (h *Hub) Capabilities(adapterName string) (SocialCapabilities, error) {
	adapter, ok := h.GetAdapter(adapterName)
	if !ok {
		return SocialCapabilities{}, fmt.Errorf("adapter '%s' not found", adapterName)
	}
	return adapter.Capabilities(), nil
}

// Supports checks if an adapter supports a specific action
func (h *Hub) Supports(adapterName, action string) bool {
	adapter, ok := h.GetAdapter(adapterName)
	if !ok {
		return false
	}
	return adapter.Capabilities().HasCapability(action)
}

// HasAdapter checks if an adapter is registered
func (h *Hub) HasAdapter(name string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.adapters[name]
	return ok
}

// Post creates a post on the specified network
func (h *Hub) Post(ctx context.Context, network string, content PostContent) (interface{}, error) {
	adapter, ok := h.GetAdapter(network)
	if !ok {
		return nil, fmt.Errorf("network '%s' not found", network)
	}
	return CreatePost(ctx, adapter, content)
}

// Vote casts a vote on the specified network
func (h *Hub) Vote(ctx context.Context, network string, content VoteContent) (interface{}, error) {
	adapter, ok := h.GetAdapter(network)
	if !ok {
		return nil, fmt.Errorf("network '%s' not found", network)
	}
	return Vote(ctx, adapter, content)
}

// Follow follows an agent on the specified network
func (h *Hub) Follow(ctx context.Context, network string, content FollowContent) (interface{}, error) {
	adapter, ok := h.GetAdapter(network)
	if !ok {
		return nil, fmt.Errorf("network '%s' not found", network)
	}
	return Follow(ctx, adapter, content)
}

// GetFeed retrieves a feed from the specified network
func (h *Hub) GetFeed(ctx context.Context, network string, request FeedRequest) ([]FeedItem, error) {
	adapter, ok := h.GetAdapter(network)
	if !ok {
		return nil, fmt.Errorf("network '%s' not found", network)
	}
	return GetFeed(ctx, adapter, request)
}

// GetNotifications retrieves notifications from the specified network
func (h *Hub) GetNotifications(ctx context.Context, network string, limit int) ([]Notification, error) {
	adapter, ok := h.GetAdapter(network)
	if !ok {
		return nil, fmt.Errorf("network '%s' not found", network)
	}
	return GetNotifications(ctx, adapter, limit)
}

// Call executes a dynamic action on the specified network
func (h *Hub) Call(ctx context.Context, network, action string, params map[string]interface{}) (interface{}, error) {
	adapter, ok := h.GetAdapter(network)
	if !ok {
		return nil, fmt.Errorf("network '%s' not found", network)
	}
	return Call(ctx, adapter, action, params)
}

// Authenticate authenticates with a network
func (h *Hub) Authenticate(ctx context.Context, network, token string) error {
	adapter, ok := h.GetAdapter(network)
	if !ok {
		return fmt.Errorf("network '%s' not found", network)
	}
	return adapter.Authenticate(ctx, token)
}

// IsAuthenticated checks if authenticated with a network
func (h *Hub) IsAuthenticated(network string) bool {
	adapter, ok := h.GetAdapter(network)
	if !ok {
		return false
	}
	return adapter.IsAuthenticated()
}

// HealthCheck checks the health of all adapters
func (h *Hub) HealthCheck(ctx context.Context) map[string]error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make(map[string]error)
	for name, adapter := range h.adapters {
		if err := adapter.HealthCheck(ctx); err != nil {
			results[name] = err
		}
	}
	return results
}

// GetManifest returns the manifest for a network
func (h *Hub) GetManifest(network string) (*NetworkManifest, error) {
	adapter, ok := h.GetAdapter(network)
	if !ok {
		return nil, fmt.Errorf("network '%s' not found", network)
	}
	return adapter.GetManifest(), nil
}

// Broadcast sends a message to all registered adapters
func (h *Hub) Broadcast(ctx context.Context, action string, params map[string]interface{}) map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make(map[string]interface{})
	for name, adapter := range h.adapters {
		if adapter.Capabilities().HasCapability(action) {
			result, _ := adapter.Call(ctx, action, params)
			results[name] = result
		}
	}
	return results
}

// AdapterCount returns the number of registered adapters
func (h *Hub) AdapterCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.adapters)
}
