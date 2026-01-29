package cost

import "time"

// ModelPricing defines pricing for a specific model
type ModelPricing struct {
	ModelID          string    `json:"model_id"`
	InputPricePer1K  float64   `json:"input_price_per_1k"`  // Price per 1K input tokens
	OutputPricePer1K float64   `json:"output_price_per_1k"` // Price per 1K output tokens
	Provider         string    `json:"provider"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CostSummary represents aggregated cost information
type CostSummary struct {
	TotalInputTokens  int64     `json:"total_input_tokens"`
	TotalOutputTokens int64     `json:"total_output_tokens"`
	TotalTokens       int64     `json:"total_tokens"`
	TotalInputCost    float64   `json:"total_input_cost"`
	TotalOutputCost   float64   `json:"total_output_cost"`
	TotalCost         float64   `json:"total_cost"`
	RequestCount      int       `json:"request_count"`
	AverageCostPerReq float64   `json:"average_cost_per_request"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
}

// SessionCost represents cost for a specific session
type SessionCost struct {
	SessionID      string    `json:"session_id"`
	TotalCost      float64   `json:"total_cost"`
	TotalTokens    int64     `json:"total_tokens"`
	RequestCount   int       `json:"request_count"`
	LastActivityAt time.Time `json:"last_activity_at"`
}

// CostInfo represents cost tracking information for a single request
type CostInfo struct {
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	TotalTokens  int64   `json:"total_tokens"`
	InputCost    float64 `json:"input_cost"`
	OutputCost   float64 `json:"output_cost"`
	TotalCost    float64 `json:"total_cost"`
	Model        string  `json:"model,omitempty"`
}

// BudgetConfig defines spending limits
type BudgetConfig struct {
	DailyBudget      float64 `json:"daily_budget"`
	MonthlyBudget    float64 `json:"monthly_budget"`
	WarningThreshold float64 `json:"warning_threshold"` // Percentage of budget
}

// BudgetStatus represents current budget usage
type BudgetStatus struct {
	DailySpent       float64  `json:"daily_spent"`
	DailyRemaining   float64  `json:"daily_remaining"`
	DailyPercent     float64  `json:"daily_percent"`
	MonthlySpent     float64  `json:"monthly_spent"`
	MonthlyRemaining float64  `json:"monthly_remaining"`
	MonthlyPercent   float64  `json:"monthly_percent"`
	IsOverBudget     bool     `json:"is_over_budget"`
	Warnings         []string `json:"warnings"`
}

// CostQueryOptions defines filtering for cost queries
type CostQueryOptions struct {
	SessionID string
	StartTime *time.Time
	EndTime   *time.Time
	ModelID   string
	Provider  string
	GroupBy   string // "session", "day", "model", "provider"
	Limit     int
	Offset    int
}

// CostOptimization represents optimization suggestions
type CostOptimization struct {
	Type            string  `json:"type"` // "model_switch", "context_reduction", "caching"
	SavingsEstimate float64 `json:"savings_estimate"`
	Description     string  `json:"description"`
	Priority        int     `json:"priority"` // 1-5, 1 is highest priority
}
