package social

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Registry manages social adapter registration and discovery
type Registry struct {
	adapters  map[string]func() SocialAdapter
	manifests map[string]*NetworkManifest
	mu        sync.RWMutex
}

// NewRegistry creates a new social adapter registry
func NewRegistry() *Registry {
	return &Registry{
		adapters:  make(map[string]func() SocialAdapter),
		manifests: make(map[string]*NetworkManifest),
	}
}

// Register registers a social adapter factory
func (r *Registry) Register(name string, factory func() SocialAdapter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("cannot register adapter with empty name")
	}

	if factory == nil {
		return fmt.Errorf("cannot register nil factory for adapter '%s'", name)
	}

	if _, exists := r.adapters[name]; exists {
		return fmt.Errorf("adapter '%s' already registered", name)
	}

	r.adapters[name] = factory

	// Create instance to get manifest
	adapter := factory()
	r.manifests[name] = adapter.GetManifest()

	return nil
}

// RegisterAdapter registers a pre-created adapter instance
func (r *Registry) RegisterAdapter(adapter SocialAdapter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := adapter.Name()
	if name == "" {
		return fmt.Errorf("cannot register adapter with empty name")
	}

	if _, exists := r.adapters[name]; exists {
		return fmt.Errorf("adapter '%s' already registered", name)
	}

	// Create factory that returns new instances
	r.adapters[name] = func() SocialAdapter {
		return adapter
	}
	r.manifests[name] = adapter.GetManifest()

	return nil
}

// Unregister removes an adapter from the registry
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.adapters, name)
	delete(r.manifests, name)
}

// Create creates a new adapter instance by name
func (r *Registry) Create(name string) (SocialAdapter, error) {
	r.mu.RLock()
	factory, exists := r.adapters[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("adapter '%s' not found in registry", name)
	}

	return factory(), nil
}

// GetManifest returns the manifest for an adapter
func (r *Registry) GetManifest(name string) (*NetworkManifest, error) {
	r.mu.RLock()
	manifest, exists := r.manifests[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("adapter '%s' not found in registry", name)
	}

	return manifest, nil
}

// List returns all registered adapter names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}

// Capabilities returns capabilities for an adapter
func (r *Registry) Capabilities(name string) (SocialCapabilities, error) {
	manifest, err := r.GetManifest(name)
	if err != nil {
		return SocialCapabilities{}, err
	}
	return manifest.SocialFeatures, nil
}

// Supports checks if an adapter supports a specific action
func (r *Registry) Supports(name, action string) bool {
	caps, err := r.Capabilities(name)
	if err != nil {
		return false
	}
	return caps.HasCapability(action)
}

// HasAdapter checks if an adapter is registered
func (r *Registry) HasAdapter(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.adapters[name]
	return exists
}

// Count returns the number of registered adapters
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.adapters)
}

// ToJSON returns registry state as JSON
func (r *Registry) ToJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	type registryState struct {
		Adapters  []string               `json:"adapters"`
		Manifests map[string]interface{} `json:"manifests"`
	}

	state := registryState{
		Adapters:  r.List(),
		Manifests: make(map[string]interface{}),
	}

	for name, manifest := range r.manifests {
		state.Manifests[name] = manifest
	}

	return json.MarshalIndent(state, "", "  ")
}

// DefaultRegistry is the global default registry
var DefaultRegistry = NewRegistry()

// RegisterDefault registers an adapter with the default registry
func RegisterDefault(name string, factory func() SocialAdapter) error {
	return DefaultRegistry.Register(name, factory)
}

// CreateDefault creates an adapter from the default registry
func CreateDefault(name string) (SocialAdapter, error) {
	return DefaultRegistry.Create(name)
}

// ListDefault lists all adapters in the default registry
func ListDefault() []string {
	return DefaultRegistry.List()
}

// InitializeDefaultRegistry sets up the default registry with built-in adapters
func InitializeDefaultRegistry() {
	// Register Moltbook adapter
	RegisterDefault("moltbook", func() SocialAdapter {
		return NewMoltbookAdapter("https://moltbook.com")
	})
}
