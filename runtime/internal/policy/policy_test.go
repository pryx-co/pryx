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
