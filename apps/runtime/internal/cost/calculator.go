package cost

import (
	"pryx-core/internal/llm"
)

// CostCalculator handles cost calculations from token usage
type CostCalculator struct {
	pricingManager *PricingManager
}

// NewCostCalculator creates a new cost calculator
func NewCostCalculator(pricingManager *PricingManager) *CostCalculator {
	return &CostCalculator{
		pricingManager: pricingManager,
	}
}

// CalculateFromUsage calculates cost from LLM usage data
func (c *CostCalculator) CalculateFromUsage(modelID string, usage llm.Usage) (CostInfo, error) {
	inputCostVal, outputCostVal := c.calculateBreakdown(modelID, usage)

	return CostInfo{
		InputTokens:  int64(usage.PromptTokens),
		OutputTokens: int64(usage.CompletionTokens),
		TotalTokens:  int64(usage.TotalTokens),
		InputCost:    inputCostVal,
		OutputCost:   outputCostVal,
		TotalCost:    inputCostVal + outputCostVal,
		Model:        modelID,
	}, nil
}

// calculateBreakdown calculates individual input/output costs
func (c *CostCalculator) calculateBreakdown(modelID string, usage llm.Usage) (float64, float64) {
	pricing, exists := c.pricingManager.GetPricing(modelID)
	if !exists {
		return 0, 0
	}

	inputCost := float64(usage.PromptTokens) * pricing.InputPricePer1K / PricePer1K
	outputCost := float64(usage.CompletionTokens) * pricing.OutputPricePer1K / PricePer1K

	return inputCost, outputCost
}

// CalculateSessionCost calculates total cost for a session from multiple requests
func (c *CostCalculator) CalculateSessionCost(requests []CostInfo) CostSummary {
	var summary CostSummary

	for _, req := range requests {
		summary.TotalInputTokens += req.InputTokens
		summary.TotalOutputTokens += req.OutputTokens
		summary.TotalTokens += req.TotalTokens
		summary.TotalInputCost += req.InputCost
		summary.TotalOutputCost += req.OutputCost
		summary.TotalCost += req.TotalCost
		summary.RequestCount++
	}

	if summary.RequestCount > 0 {
		summary.AverageCostPerReq = summary.TotalCost / float64(summary.RequestCount)
	}

	return summary
}
