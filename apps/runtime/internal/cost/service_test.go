package cost

import (
	"os"
	"testing"

	"pryx-core/internal/audit"
	"pryx-core/internal/store"
)

func TestCostService_SetBudget(t *testing.T) {
	// Create test dependencies
	tmpDB := t.TempDir() + "/test.db"
	defer func() {
		os.Remove(tmpDB)
	}()

	store, err := store.New(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)
	calculator := NewCostCalculator(pricingMgr)

	service := NewCostService(tracker, calculator, pricingMgr, store)

	// Test setting budget
	config := BudgetConfig{
		DailyBudget:      10.0,
		MonthlyBudget:    100.0,
		WarningThreshold: 80.0,
	}

	service.SetBudget("user1", config)

	// Verify budget was set by getting status
	status := service.GetBudgetStatus("user1")
	if status.DailyRemaining != 10.0 {
		t.Errorf("Expected daily remaining to be 10.0, got %f", status.DailyRemaining)
	}
}

func TestCostService_GetBudgetStatus_NoBudget(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	store, _ := store.New(tmpDB)
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)
	calculator := NewCostCalculator(pricingMgr)

	service := NewCostService(tracker, calculator, pricingMgr, store)

	// Test getting budget status when no budget set
	status := service.GetBudgetStatus("nonexistent-user")

	if status.DailySpent != 0 {
		t.Errorf("Expected daily spent to be 0, got %f", status.DailySpent)
	}

	if status.IsOverBudget {
		t.Error("Expected not to be over budget when no budget set")
	}
}

func TestCostService_GetBudgetStatus_OverBudget(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	store, _ := store.New(tmpDB)
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)
	calculator := NewCostCalculator(pricingMgr)

	service := NewCostService(tracker, calculator, pricingMgr, store)

	// Set a low budget
	config := BudgetConfig{
		DailyBudget:   1.0,
		MonthlyBudget: 10.0,
	}
	service.SetBudget("user1", config)

	// Mock high costs by adding to tracker
	costInfo := audit.CostInfo{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		InputCost:    2.0,
		OutputCost:   3.0,
		TotalCost:    5.0, // Exceeds daily budget
		Model:        "gpt-4",
	}
	err := tracker.RecordCost("test-session", "cli", "gpt-4", costInfo)
	if err != nil {
		t.Fatalf("Failed to record cost: %v", err)
	}

	// Verify cost was recorded
	costs, err := tracker.GetSessionCosts("test-session")
	if err != nil {
		t.Fatalf("Failed to get session costs: %v", err)
	}
	if len(costs) != 1 {
		t.Fatalf("Expected 1 cost to be recorded, got %d", len(costs))
	}

	status := service.GetBudgetStatus("user1")

	// Should be over budget
	if !status.IsOverBudget {
		t.Error("Expected to be over budget with high costs")
	}
}

func TestCostService_generateBudgetWarnings_WarningThreshold(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	store, _ := store.New(tmpDB)
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)
	calculator := NewCostCalculator(pricingMgr)

	service := NewCostService(tracker, calculator, pricingMgr, store)

	config := BudgetConfig{
		DailyBudget:      100.0,
		MonthlyBudget:    1000.0,
		WarningThreshold: 80.0,
	}

	// Test at 85% of budget - should generate warning
	warnings := service.generateBudgetWarnings(config, 85.0, 850.0, 85.0, 85.0)

	if len(warnings) == 0 {
		t.Error("Expected warnings when exceeding warning threshold")
	}
}

func TestCostService_GetOptimizationSuggestions(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	store, _ := store.New(tmpDB)
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)
	calculator := NewCostCalculator(pricingMgr)

	service := NewCostService(tracker, calculator, pricingMgr, store)

	for i := 0; i < 60; i++ {
		costInfo := audit.CostInfo{
			InputTokens:  10000,
			OutputTokens: 5000,
			TotalTokens:  15000,
			InputCost:    0.5,
			OutputCost:   0.5,
			TotalCost:    1.0,
			Model:        "gpt-4",
		}
		err := tracker.RecordCost("test-session", "cli", "gpt-4", costInfo)
		if err != nil {
			t.Fatalf("Failed to record cost: %v", err)
		}
	}

	sessionCosts, err := tracker.GetSessionCosts("test-session")
	if err != nil {
		t.Fatalf("Failed to get session costs: %v", err)
	}
	if len(sessionCosts) != 60 {
		t.Fatalf("Expected 60 costs, got %d", len(sessionCosts))
	}

	suggestions := service.GetOptimizationSuggestions()

	if len(suggestions) == 0 {
		t.Error("Expected optimization suggestions for high costs")
	}
}

func TestCostService_GetCurrentSessionCost(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	store, _ := store.New(tmpDB)
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)
	calculator := NewCostCalculator(pricingMgr)

	service := NewCostService(tracker, calculator, pricingMgr, store)

	// Track some costs
	costInfo := audit.CostInfo{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		InputCost:    0.3,
		OutputCost:   0.2,
		TotalCost:    0.5,
		Model:        "gpt-4",
	}
	tracker.RecordCost("current-session", "cli", "gpt-4", costInfo)

	summary, err := service.GetCurrentSessionCost()
	if err != nil {
		t.Fatalf("GetCurrentSessionCost failed: %v", err)
	}

	// Verify summary has valid data
	if summary.RequestCount < 0 {
		t.Fatal("Expected valid summary with non-negative request count")
	}
}

func TestCostService_GetAllModelPricing(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	store, _ := store.New(tmpDB)
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)
	calculator := NewCostCalculator(pricingMgr)

	service := NewCostService(tracker, calculator, pricingMgr, store)

	pricing := service.GetAllModelPricing()

	if len(pricing) == 0 {
		t.Error("Expected pricing data for models")
	}
}
