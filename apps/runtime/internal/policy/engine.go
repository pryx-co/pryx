package policy

import (
	"fmt"
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
			// Check scope matching if specified
			if rule.Scope != "" && !matchScope(rule.Scope, args) {
				continue
			}
			// Check argument matching if specified
			if len(rule.Args) > 0 && !matchArgs(rule.Args, args) {
				continue
			}
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

// matchScope checks if the current execution context matches the required scope
func matchScope(scope ScopeType, args map[string]interface{}) bool {
	// Extract current scope from args (context should provide this)
	currentScope := ""
	if args != nil {
		if s, ok := args["_scope"].(string); ok {
			currentScope = s
		}
	}

	// If no scope specified in context, default to global
	if currentScope == "" {
		currentScope = string(ScopeGlobal)
	}

	switch scope {
	case ScopeGlobal:
		// Global scope: only operations explicitly marked as global
		return currentScope == string(ScopeGlobal)
	case ScopeWorkspace:
		// Workspace scope: allows operations in workspace context only
		return currentScope == string(ScopeWorkspace)
	case ScopeNetwork:
		// Network scope: allows operations in network context only
		return currentScope == string(ScopeNetwork)
	default:
		return false
	}
}

// matchArgs checks if the provided arguments match the required pattern
func matchArgs(matchers []ArgMatcher, args map[string]interface{}) bool {
	if args == nil {
		args = make(map[string]interface{})
	}

	for _, matcher := range matchers {
		if !matchArg(matcher, args) {
			return false
		}
	}
	return true
}

// matchArg checks if a single argument matches the required pattern
func matchArg(matcher ArgMatcher, args map[string]interface{}) bool {
	actualValue, exists := args[matcher.Key]

	switch matcher.Operator {
	case "exists":
		return exists
	case "not_exists":
		return !exists
	case "eq":
		if !exists {
			return false
		}
		return compareValues(actualValue, matcher.Value)
	case "neq":
		if !exists {
			return true
		}
		return !compareValues(actualValue, matcher.Value)
	case "contains":
		if !exists {
			return false
		}
		return containsValue(actualValue, matcher.Value)
	case "regex":
		if !exists {
			return false
		}
		return matchRegex(actualValue, matcher.Value)
	default:
		return false
	}
}

// compareValues compares an actual value with an expected string value
func compareValues(actual interface{}, expected string) bool {
	switch v := actual.(type) {
	case string:
		return v == expected
	case int, int32, int64:
		return fmt.Sprintf("%d", v) == expected
	case float32, float64:
		return fmt.Sprintf("%f", v) == expected
	case bool:
		return fmt.Sprintf("%t", v) == expected
	default:
		return fmt.Sprintf("%v", v) == expected
	}
}

// containsValue checks if the actual value contains the expected value
func containsValue(actual interface{}, expected string) bool {
	switch v := actual.(type) {
	case string:
		return strings.Contains(v, expected)
	case []interface{}:
		for _, item := range v {
			if compareValues(item, expected) {
				return true
			}
		}
		return false
	case map[string]interface{}:
		_, exists := v[expected]
		return exists
	default:
		return false
	}
}

// matchRegex checks if the actual value matches the expected regex pattern
func matchRegex(actual interface{}, pattern string) bool {
	switch v := actual.(type) {
	case string:
		matched, _ := regexp.MatchString(pattern, v)
		return matched
	default:
		return false
	}
}
