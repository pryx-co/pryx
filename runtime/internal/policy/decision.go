package policy

// Decision represents the result of a policy evaluation
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
	DecisionAsk   Decision = "ask"
)

// Result contains the decision and any reasoning
type Result struct {
	Decision Decision `json:"decision"`
	Reason   string   `json:"reason,omitempty"`
}

func Allow(reason string) Result {
	return Result{Decision: DecisionAllow, Reason: reason}
}

func Deny(reason string) Result {
	return Result{Decision: DecisionDeny, Reason: reason}
}

func Ask(reason string) Result {
	return Result{Decision: DecisionAsk, Reason: reason}
}
