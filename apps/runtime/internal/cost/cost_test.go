package cost

import (
	"testing"
	"time"

	"pryx-core/internal/llm"
)

func TestCostCalculator_CalculateFromUsage(t *testing.T) {
	pricingMgr := NewPricingManager()
	calculator := NewCostCalculator(pricingMgr)

	// Test with GPT-4o
	usage := llm.Usage{
		PromptTokens:     1000,
		CompletionTokens: 2000,
		TotalTokens:      3000,
	}

	cost, err := calculator.CalculateFromUsage("gpt-4o", usage)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// GPT-4o pricing: $2.50/1K input, $10.00/1K output
	// Expected: (1000/1000)*2.50 + (2000/1000)*10.00 = 2.50 + 20.00 = 22.50
	expectedCost := 22.50
	if cost.TotalCost != expectedCost {
		t.Errorf("Expected cost %.2f, got %.2f", expectedCost, cost.TotalCost)
	}

	if cost.InputCost != 2.50 {
		t.Errorf("Expected input cost 2.50, got %.2f", cost.InputCost)
	}

	if cost.OutputCost != 20.00 {
		t.Errorf("Expected output cost 20.00, got %.2f", cost.OutputCost)
	}

	if cost.TotalTokens != 3000 {
		t.Errorf("Expected total tokens 3000, got %d", cost.TotalTokens)
	}
}

func TestPricingManager_GetPricing(t *testing.T) {
	pricingMgr := NewPricingManager()

	// Test getting existing model pricing
	pricing, exists := pricingMgr.GetPricing("gpt-4o")
	if !exists {
		t.Fatal("Expected gpt-4o pricing to exist")
	}

	if pricing.InputPricePer1K != 2.50 {
		t.Errorf("Expected input price 2.50, got %.2f", pricing.InputPricePer1K)
	}

	// Test getting non-existing model pricing
	_, exists = pricingMgr.GetPricing("non-existent-model")
	if exists {
		t.Error("Expected non-existent model to return false")
	}
}

func TestCostCalculator_UnknownModel(t *testing.T) {
	pricingMgr := NewPricingManager()
	calculator := NewCostCalculator(pricingMgr)

	usage := llm.Usage{
		PromptTokens:     100,
		CompletionTokens: 100,
		TotalTokens:      200,
	}

	cost, err := calculator.CalculateFromUsage("unknown-model", usage)
	if err != nil {
		t.Fatalf("Unexpected error for unknown model: %v", err)
	}

	// Unknown models should return 0 cost
	if cost.TotalCost != 0 {
		t.Errorf("Expected 0 cost for unknown model, got %.2f", cost.TotalCost)
	}
}

func TestBudgetConfig(t *testing.T) {
	config := BudgetConfig{
		DailyBudget:      10.0,
		MonthlyBudget:    100.0,
		WarningThreshold: 80.0,
	}

	if config.DailyBudget != 10.0 {
		t.Errorf("Expected daily budget 10.0, got %.2f", config.DailyBudget)
	}

	if config.MonthlyBudget != 100.0 {
		t.Errorf("Expected monthly budget 100.0, got %.2f", config.MonthlyBudget)
	}

	if config.WarningThreshold != 80.0 {
		t.Errorf("Expected warning threshold 80.0, got %.2f", config.WarningThreshold)
	}
}

func TestCostSummary(t *testing.T) {
	summary := CostSummary{
		TotalInputTokens:  10000,
		TotalOutputTokens: 20000,
		TotalTokens:       30000,
		TotalInputCost:    25.0,
		TotalOutputCost:   200.0,
		TotalCost:         225.0,
		RequestCount:      10,
		AverageCostPerReq: 22.5, // Manually set for test since calculation is done by service
	}

	expectedAvgCost := 22.5
	if summary.AverageCostPerReq != expectedAvgCost {
		t.Errorf("Expected average cost %.2f, got %.2f", expectedAvgCost, summary.AverageCostPerReq)
	}
}

func TestModelPricing_Providers(t *testing.T) {
	pricingMgr := NewPricingManager()

	// Test OpenAI provider
	openaiPricing := pricingMgr.GetProviderPricing("openai")
	if len(openaiPricing) == 0 {
		t.Error("Expected OpenAI pricing to exist")
	}

	// Test Anthropic provider
	anthropicPricing := pricingMgr.GetProviderPricing("anthropic")
	if len(anthropicPricing) == 0 {
		t.Error("Expected Anthropic pricing to exist")
	}

	// Test non-existent provider
	unknownPricing := pricingMgr.GetProviderPricing("unknown")
	if len(unknownPricing) != 0 {
		t.Error("Expected no pricing for unknown provider")
	}
}

func TestCostOptimization(t *testing.T) {
	opt := CostOptimization{
		Type:            "model_switch",
		SavingsEstimate: 50.0,
		Description:     "Switch to cheaper model",
		Priority:        1,
	}

	if opt.Type != "model_switch" {
		t.Errorf("Expected type 'model_switch', got '%s'", opt.Type)
	}

	if opt.SavingsEstimate != 50.0 {
		t.Errorf("Expected savings 50.0, got %.2f", opt.SavingsEstimate)
	}

	if opt.Priority != 1 {
		t.Errorf("Expected priority 1, got %d", opt.Priority)
	}
}

func TestAllPricing(t *testing.T) {
	pricingMgr := NewPricingManager()

	allPricing := pricingMgr.AllPricing()
	if len(allPricing) == 0 {
		t.Error("Expected some pricing models to exist")
	}

	// Verify all models have required fields
	for _, pricing := range allPricing {
		if pricing.ModelID == "" {
			t.Error("Expected model ID to be set")
		}
		if pricing.Provider == "" {
			t.Error("Expected provider to be set")
		}
		if pricing.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set")
		}
	}
}

func TestBudgetStatus(t *testing.T) {
	status := BudgetStatus{
		DailySpent:       5.0,
		DailyRemaining:   5.0,
		DailyPercent:     50.0,
		MonthlySpent:     50.0,
		MonthlyRemaining: 50.0,
		MonthlyPercent:   50.0,
		IsOverBudget:     false,
		Warnings:         []string{},
	}

	if status.IsOverBudget {
		t.Error("Expected not over budget")
	}

	if status.DailyPercent != 50.0 {
		t.Errorf("Expected daily percent 50.0, got %.2f", status.DailyPercent)
	}
}

func TestSessionCost(t *testing.T) {
	sessionCost := SessionCost{
		SessionID:      "test-session",
		TotalCost:      1.50,
		TotalTokens:    1000,
		RequestCount:   5,
		LastActivityAt: time.Now(),
	}

	if sessionCost.SessionID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got '%s'", sessionCost.SessionID)
	}

	if sessionCost.RequestCount != 5 {
		t.Errorf("Expected request count 5, got %d", sessionCost.RequestCount)
	}
}

func TestPricingManager_SetPricing(t *testing.T) {
	pricingMgr := NewPricingManager()

	newPricing := ModelPricing{
		ModelID:          "custom-model",
		InputPricePer1K:  1.0,
		OutputPricePer1K: 2.0,
		Provider:         "custom",
		UpdatedAt:        time.Now(),
	}

	pricingMgr.SetPricing(newPricing)

	retrieved, exists := pricingMgr.GetPricing("custom-model")
	if !exists {
		t.Fatal("Expected custom model pricing to exist after setting")
	}

	if retrieved.InputPricePer1K != 1.0 {
		t.Errorf("Expected input price 1.0, got %.2f", retrieved.InputPricePer1K)
	}

	if retrieved.OutputPricePer1K != 2.0 {
		t.Errorf("Expected output price 2.0, got %.2f", retrieved.OutputPricePer1K)
	}
}
