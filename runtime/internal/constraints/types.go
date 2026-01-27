package constraints

type Action string

const (
	ActionAllow    Action = "allow"
	ActionDeny     Action = "deny"
	ActionFallback Action = "fallback"
	ActionAsk      Action = "ask"
)

type Resolution struct {
	Action      Action `json:"action"`
	TargetModel string `json:"target_model,omitempty"` // If fallback
	Reason      string `json:"reason,omitempty"`
}

type Request struct {
	Model        string
	PromptTokens int
	OutputTokens int // Estimated or MaxTokens
	Tools        []string
	Images       bool
}
