package constraints

type Action string

const (
	ActionAllow    Action = "allow"
	ActionDeny     Action = "deny"
	ActionFallback Action = "fallback"
	ActionAsk      Action = "ask"
)

type Resolution struct {
	Action           Action  `json:"action"`
	TargetModel      string  `json:"target_model,omitempty"` // If fallback
	Reason           string  `json:"reason,omitempty"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd,omitempty"` // Cost for the resolved model
}

type Request struct {
	Model          string
	ProviderID     string
	PromptTokens   int
	OutputTokens   int
	ThinkingTokens int
	Tools          []string
	Images         bool
	MaxCostUSD     float64 `json:"max_cost_usd,omitempty"` // Optional cost constraint
}

type CostEstimate struct {
	InputTokensCost    float64
	OutputTokensCost   float64
	ThinkingTokensCost float64
	FixedCost          float64
	TotalUSD           float64
}
