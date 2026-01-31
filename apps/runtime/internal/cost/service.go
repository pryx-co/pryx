package cost

import (
	"time"

	"pryx-core/internal/store"
)

// CostService provides cost awareness features
type CostService struct {
	tracker      *CostTracker
	calculator   *CostCalculator
	pricingMgr   *PricingManager
	sessionStore *store.Store
	budgets      map[string]BudgetConfig
}

// NewCostService creates a new cost service
func NewCostService(tracker *CostTracker, calculator *CostCalculator, pricingMgr *PricingManager, sessionStore *store.Store) *CostService {
	return &CostService{
		tracker:      tracker,
		calculator:   calculator,
		pricingMgr:   pricingMgr,
		sessionStore: sessionStore,
		budgets:      make(map[string]BudgetConfig),
	}
}

// SetBudget sets a budget for a user
func (s *CostService) SetBudget(userID string, config BudgetConfig) {
	s.budgets[userID] = config
}

// GetBudgetStatus returns current budget usage for a user
func (s *CostService) GetBudgetStatus(userID string) BudgetStatus {
	config, exists := s.budgets[userID]
	if !exists {
		return BudgetStatus{
			DailySpent:       0,
			DailyRemaining:   0,
			DailyPercent:     0,
			MonthlySpent:     0,
			MonthlyRemaining: 0,
			MonthlyPercent:   0,
			IsOverBudget:     false,
			Warnings:         []string{},
		}
	}

	today := time.Now().UTC()
	dailyCosts, _ := s.tracker.GetDailyCosts(today)
	monthlyCosts, _ := s.tracker.GetMonthlyCosts(today.Year(), today.Month())

	dailyPercent := 0.0
	if config.DailyBudget > 0 {
		dailyPercent = (dailyCosts.TotalCost / config.DailyBudget) * 100
	}

	monthlyPercent := 0.0
	if config.MonthlyBudget > 0 {
		monthlyPercent = (monthlyCosts.TotalCost / config.MonthlyBudget) * 100
	}

	status := BudgetStatus{
		DailySpent:       dailyCosts.TotalCost,
		DailyRemaining:   config.DailyBudget - dailyCosts.TotalCost,
		DailyPercent:     dailyPercent,
		MonthlySpent:     monthlyCosts.TotalCost,
		MonthlyRemaining: config.MonthlyBudget - monthlyCosts.TotalCost,
		MonthlyPercent:   monthlyPercent,
		IsOverBudget:     dailyCosts.TotalCost > config.DailyBudget || monthlyCosts.TotalCost > config.MonthlyBudget,
		Warnings:         s.generateBudgetWarnings(config, dailyCosts.TotalCost, monthlyCosts.TotalCost, dailyPercent, monthlyPercent),
	}

	return status
}

// generateBudgetWarnings creates warning messages based on budget usage
func (s *CostService) generateBudgetWarnings(config BudgetConfig, dailySpent, monthlySpent, dailyPercent, monthlyPercent float64) []string {
	var warnings []string

	// Warning threshold warnings
	if config.WarningThreshold > 0 {
		if dailyPercent >= config.WarningThreshold {
			warnings = append(warnings, "Daily spending has reached warning threshold")
		}
		if monthlyPercent >= config.WarningThreshold {
			warnings = append(warnings, "Monthly spending has reached warning threshold")
		}
	}

	// Over budget warnings
	if dailySpent > config.DailyBudget {
		warnings = append(warnings, "Daily budget exceeded")
	}
	if monthlySpent > config.MonthlyBudget {
		warnings = append(warnings, "Monthly budget exceeded")
	}

	return warnings
}

// GetOptimizationSuggestions returns cost optimization suggestions
func (s *CostService) GetOptimizationSuggestions() []CostOptimization {
	var suggestions []CostOptimization

	// Get current month costs to analyze
	today := time.Now().UTC()
	monthlyCosts, _ := s.tracker.GetMonthlyCosts(today.Year(), today.Month())

	// Suggest switching to cheaper models if using expensive ones
	if monthlyCosts.TotalCost > 10.0 {
		suggestions = append(suggestions, CostOptimization{
			Type:            "model_switch",
			SavingsEstimate: monthlyCosts.TotalCost * 0.3, // Estimate 30% savings
			Description:     "Consider using gpt-4o-mini or claude-haiku-3 for non-complex tasks to reduce costs by up to 70%",
			Priority:        1,
		})
	}

	// Suggest context optimization if token counts are high
	if monthlyCosts.TotalTokens > 100000 {
		suggestions = append(suggestions, CostOptimization{
			Type:            "context_reduction",
			SavingsEstimate: monthlyCosts.TotalCost * 0.2, // Estimate 20% savings
			Description:     "Your token usage is high. Consider summarizing conversation history or using shorter context windows",
			Priority:        2,
		})
	}

	// Suggest caching if many similar requests
	if monthlyCosts.RequestCount > 50 {
		suggestions = append(suggestions, CostOptimization{
			Type:            "caching",
			SavingsEstimate: monthlyCosts.TotalCost * 0.15, // Estimate 15% savings
			Description:     "Enable response caching for repeated queries to avoid redundant API calls",
			Priority:        3,
		})
	}

	// Add budget-specific suggestions
	budgetStatus := s.GetBudgetStatus("default")
	if budgetStatus.IsOverBudget {
		suggestions = append(suggestions, CostOptimization{
			Type:            "budget_alert",
			SavingsEstimate: 0,
			Description:     "Budget exceeded! Consider reducing usage or increasing budget limits",
			Priority:        1,
		})
	}

	return suggestions
}

// GetCurrentSessionCost returns the current cost for active sessions
func (s *CostService) GetCurrentSessionCost() (CostSummary, error) {
	sessions, err := s.sessionStore.ListSessions()
	if err != nil {
		return CostSummary{}, err
	}

	var totalCost float64
	var totalTokens int64
	var requestCount int

	for _, session := range sessions {
		costs, err := s.tracker.GetSessionCosts(session.ID)
		if err != nil {
			continue
		}

		for _, cost := range costs {
			totalCost += cost.TotalCost
			totalTokens += cost.TotalTokens
			requestCount++
		}
	}

	return CostSummary{
		TotalCost:    totalCost,
		TotalTokens:  totalTokens,
		RequestCount: requestCount,
		AverageCostPerReq: func() float64 {
			if requestCount == 0 {
				return 0
			}
			return totalCost / float64(requestCount)
		}(),
		PeriodStart: time.Now().Add(-24 * time.Hour),
		PeriodEnd:   time.Now(),
	}, nil
}

// GetAllModelPricing returns pricing for all available models
func (s *CostService) GetAllModelPricing() []ModelPricing {
	return s.pricingMgr.AllPricing()
}

// GetModelPricing returns pricing for a specific model
func (s *CostService) GetModelPricing(modelID string) (ModelPricing, bool) {
	return s.pricingMgr.GetPricing(modelID)
}
