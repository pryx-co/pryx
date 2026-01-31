package registry_test

import (
	"pryx-core/internal/registry"
	"testing"
)

// TestAgent creates a test agent for testing
func TestAgent() *registry.Agent {
	return &registry.Agent{
		ID:          "test-agent-1",
		Name:        "Test Agent",
		Description: "A test agent for unit testing",
		Version:     "1.0.0",
		Capabilities: []registry.Capability{
			{Type: "tool", Name: "execute", Version: "1.0", Description: "Execute commands", Permissions: []string{"shell", "read"}},
		},
		Endpoint: registry.Endpoint{
			Type: "http",
			Host: "localhost",
			Port: "8080",
			URL:  "http://localhost:8080",
		},
		TrustLevel: registry.TrustLevelTrusted,
	}
}
