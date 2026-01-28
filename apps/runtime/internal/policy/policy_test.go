package policy

import (
	"testing"
)

func TestDefaultPolicy(t *testing.T) {
	engine := NewEngine(nil)

	// Test default behavior (should be Ask)
	res := engine.Evaluate("some.tool", nil)
	if res.Decision != DecisionAsk {
		t.Errorf("Expected default decision Ask, got %v", res.Decision)
	}
}

func TestExplicitAllow(t *testing.T) {
	p := NewDefaultPolicy()
	p.Rules = append(p.Rules, Rule{
		Tool:        "safe.*",
		Decision:    DecisionAllow,
		Description: "Allow safe tools",
	})
	engine := NewEngine(p)

	res := engine.Evaluate("safe.read_file", nil)
	if res.Decision != DecisionAllow {
		t.Errorf("Expected Allow for safe.read_file, got %v", res.Decision)
	}

	res = engine.Evaluate("dangerous.exec", nil)
	if res.Decision != DecisionAsk {
		t.Errorf("Expected Ask for dangerous.exec, got %v", res.Decision)
	}
}

func TestExactMatch(t *testing.T) {
	p := NewDefaultPolicy()
	p.Rules = append(p.Rules, Rule{
		Tool:     "exact.match",
		Decision: DecisionDeny,
	})
	engine := NewEngine(p)

	res := engine.Evaluate("exact.match", nil)
	if res.Decision != DecisionDeny {
		t.Errorf("Expected Deny, got %v", res.Decision)
	}
}

func TestScopeMatching(t *testing.T) {
	p := NewDefaultPolicy()
	p.Rules = append(p.Rules, Rule{
		Tool:        "workspace.*",
		Scope:       ScopeWorkspace,
		Decision:    DecisionAllow,
		Description: "Allow workspace tools",
	})
	engine := NewEngine(p)

	// Test with workspace scope
	res := engine.Evaluate("workspace.read", map[string]interface{}{"_scope": "workspace"})
	if res.Decision != DecisionAllow {
		t.Errorf("Expected Allow for workspace scope, got %v", res.Decision)
	}

	// Test with global scope (should fall back to default)
	res = engine.Evaluate("workspace.read", map[string]interface{}{"_scope": "global"})
	if res.Decision != DecisionAsk {
		t.Errorf("Expected Ask for global scope with workspace rule, got %v", res.Decision)
	}

	// Test with no scope (defaults to global)
	res = engine.Evaluate("workspace.read", nil)
	if res.Decision != DecisionAsk {
		t.Errorf("Expected Ask for no scope with workspace rule, got %v", res.Decision)
	}
}

func TestArgMatching(t *testing.T) {
	p := NewDefaultPolicy()
	p.Rules = append(p.Rules, Rule{
		Tool:        "file.read",
		Args:        []ArgMatcher{{Key: "path", Operator: "eq", Value: "/safe/path"}},
		Decision:    DecisionAllow,
		Description: "Allow safe file reads",
	})
	engine := NewEngine(p)

	// Test matching arg
	res := engine.Evaluate("file.read", map[string]interface{}{"path": "/safe/path"})
	if res.Decision != DecisionAllow {
		t.Errorf("Expected Allow for matching arg, got %v", res.Decision)
	}

	// Test non-matching arg
	res = engine.Evaluate("file.read", map[string]interface{}{"path": "/dangerous/path"})
	if res.Decision != DecisionAsk {
		t.Errorf("Expected Ask for non-matching arg, got %v", res.Decision)
	}
}

func TestArgExistsMatching(t *testing.T) {
	p := NewDefaultPolicy()
	p.Rules = append(p.Rules, Rule{
		Tool:        "network.request",
		Args:        []ArgMatcher{{Key: "url", Operator: "exists"}},
		Decision:    DecisionAllow,
		Description: "Allow network requests with URL",
	})
	engine := NewEngine(p)

	// Test with URL present
	res := engine.Evaluate("network.request", map[string]interface{}{"url": "https://example.com"})
	if res.Decision != DecisionAllow {
		t.Errorf("Expected Allow when URL exists, got %v", res.Decision)
	}

	// Test without URL
	res = engine.Evaluate("network.request", map[string]interface{}{})
	if res.Decision != DecisionAsk {
		t.Errorf("Expected Ask when URL missing, got %v", res.Decision)
	}
}

func TestCombinedScopeAndArgMatching(t *testing.T) {
	p := NewDefaultPolicy()
	p.Rules = append(p.Rules, Rule{
		Tool:        "data.export",
		Scope:       ScopeWorkspace,
		Args:        []ArgMatcher{{Key: "format", Operator: "eq", Value: "json"}},
		Decision:    DecisionAllow,
		Description: "Allow JSON exports in workspace",
	})
	engine := NewEngine(p)

	// Test matching both scope and arg
	res := engine.Evaluate("data.export", map[string]interface{}{"_scope": "workspace", "format": "json"})
	if res.Decision != DecisionAllow {
		t.Errorf("Expected Allow for matching scope and arg, got %v", res.Decision)
	}

	// Test matching scope but not arg
	res = engine.Evaluate("data.export", map[string]interface{}{"_scope": "workspace", "format": "csv"})
	if res.Decision != DecisionAsk {
		t.Errorf("Expected Ask when arg doesn't match, got %v", res.Decision)
	}

	// Test matching arg but not scope
	res = engine.Evaluate("data.export", map[string]interface{}{"_scope": "network", "format": "json"})
	if res.Decision != DecisionAsk {
		t.Errorf("Expected Ask when scope doesn't match, got %v", res.Decision)
	}
}
