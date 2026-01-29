package policy

// ScopeType definitions
type ScopeType string

const (
	ScopeGlobal    ScopeType = "global"
	ScopeWorkspace ScopeType = "workspace"
	ScopeNetwork   ScopeType = "network"
)

// ArgMatcher defines how to match arguments
type ArgMatcher struct {
	Key      string `json:"key"`             // Argument key to match
	Operator string `json:"operator"`        // Operator: eq, neq, exists, not_exists, contains, regex
	Value    string `json:"value,omitempty"` // Value for comparison (optional for exists/not_exists)
}

// Rule defines a single policy rule
type Rule struct {
	ID          string       `json:"id"`
	Description string       `json:"description"`
	Tool        string       `json:"tool"`           // exact match, prefix wildcard (foo*), "*" or regex
	Scope       ScopeType    `json:"scope"`          // scope requirement
	Args        []ArgMatcher `json:"args,omitempty"` // argument matchers
	Decision    Decision     `json:"decision"`
}

// Policy is a collection of rules
type Policy struct {
	Version string   `json:"version"`
	Rules   []Rule   `json:"rules"`
	Default Decision `json:"default"`
}

func NewDefaultPolicy() *Policy {
	// Secure by default: everything not explicitly allowed is "Ask" (or "Deny" if strictly locked down)
	// For MVP, "Ask" is a good balance.
	return &Policy{
		Version: "1.0",
		Rules:   []Rule{},
		Default: DecisionAsk,
	}
}
