package policy

// ScopeType definitions
type ScopeType string

const (
	ScopeGlobal    ScopeType = "global"
	ScopeWorkspace ScopeType = "workspace"
	ScopeNetwork   ScopeType = "network"
)

// Rule defines a single policy rule
type Rule struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Tool        string    `json:"tool"` // exact match or regex
	Scope       ScopeType `json:"scope"`
	Decision    Decision  `json:"decision"`
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
