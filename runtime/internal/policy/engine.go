package policy

import (
	"regexp"
	"strings"
	"sync"
)

// Engine evaluates tool calls against the active policy
type Engine struct {
	mu     sync.RWMutex
	policy *Policy
}

func NewEngine(p *Policy) *Engine {
	if p == nil {
		p = NewDefaultPolicy()
	}
	return &Engine{
		policy: p,
	}
}

// Evaluate checks if a tool call is allowed
func (e *Engine) Evaluate(toolName string, args map[string]interface{}) Result {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// 1. Check Rules
	for _, rule := range e.policy.Rules {
		if matchTool(rule.Tool, toolName) {
			// TODO: Add more complex scope/arg matching here
			return Result{Decision: rule.Decision, Reason: rule.Description}
		}
	}

	// 2. Fallback to Default
	return Result{Decision: e.policy.Default, Reason: "Default policy"}
}

func matchTool(pattern, toolName string) bool {
	if pattern == "*" || pattern == toolName {
		return true
	}
	// Simple shell-style wildcard matching could go here, or regex
	// For now, strict match or simple prefix
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(toolName, prefix)
	}
	// Try regex if it looks like one? Or keep it simple.
	// Let's assume regex for now if it contains special chars
	matched, _ := regexp.MatchString(pattern, toolName)
	return matched
}
