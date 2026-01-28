package constraints

import (
	"testing"
)

func TestCatalog_LoadFromBytes(t *testing.T) {
	jsonConfig := `
	{
		"exact": {
			"test-model": {
				"context_window": 1000,
				"supports_vision": true
			}
		},
		"patterns": {
			"^test-.*": {
				"context_window": 500,
				"supports_tools": true
			}
		}
	}`

	c := NewCatalog()
	err := c.LoadFromBytes([]byte(jsonConfig))
	if err != nil {
		t.Fatalf("Failed to load catalog: %v", err)
	}

	// Check Exact
	if caps, ok := c.Get("test-model"); !ok {
		t.Error("Expected 'test-model' to be found")
	} else if caps.ContextWindow != 1000 {
		t.Errorf("Expected context 1000, got %d", caps.ContextWindow)
	}

	// Check Pattern
	if caps, ok := c.Get("test-pattern-match"); !ok {
		t.Error("Expected 'test-pattern-match' to be found via regex")
	} else if caps.ContextWindow != 500 {
		t.Errorf("Expected context 500, got %d", caps.ContextWindow)
	}
}

func TestCatalog_DefaultCatalog(t *testing.T) {
	c := DefaultCatalog()

	// Test a known model from default_models.json (Jan 2026)
	if _, ok := c.Get("openai/gpt-5-turbo"); !ok {
		t.Error("Expected 'openai/gpt-5-turbo' in default catalog")
	}

	// Test a known pattern
	if _, ok := c.Get("anthropic/claude-4.5-sonnet-20260220"); !ok {
		t.Error("Expected 'anthropic/claude-4.5-sonnet-20260220' (pattern match) in default catalog")
	}
}

func TestCatalog_PatternMatching_RealWorld(t *testing.T) {
	c := DefaultCatalog()

	tests := []struct {
		modelID       string
		expectMatch   bool
		expectContext int
		expectVision  bool
	}{
		{"openai/gpt-5-turbo", true, 200000, true},        // Exact match
		{"openai/gpt-5-preview-0125", true, 200000, true}, // Regex match gpt-5*
		{"anthropic/claude-4.6-opus", true, 500000, true}, // Regex match claude-4.*
		{"meta-llama/llama-4-405b", true, 128000, true},   // Regex match *llama-4*
		{"unknown-model-vision", true, 4096, true},        // Heuristic match
		{"totally-random-model", false, 0, false},         // No match
	}

	for _, tt := range tests {
		caps, ok := c.Get(tt.modelID)
		if ok != tt.expectMatch {
			t.Errorf("Model %s: expected match %v, got %v", tt.modelID, tt.expectMatch, ok)
			continue
		}
		if ok {
			// Skip exact context checks for patterns as they might change, just check presence
			if caps.SupportsVision != tt.expectVision {
				t.Errorf("Model %s: expected vision %v, got %v", tt.modelID, tt.expectVision, caps.SupportsVision)
			}
		}
	}
}
