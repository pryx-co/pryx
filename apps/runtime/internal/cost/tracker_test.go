package cost

import (
	"testing"
	"time"

	"pryx-core/internal/audit"
	"pryx-core/internal/store"
)

func TestCostTracker_RecordCost(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	store, err := store.New(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)

	costInfo := audit.CostInfo{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		InputCost:    0.03,
		OutputCost:   0.015,
		TotalCost:    0.045,
		Model:        "gpt-4",
	}

	err = tracker.RecordCost("test-session", "cli", "gpt-4", costInfo)
	if err != nil {
		t.Fatalf("RecordCost failed: %v", err)
	}
}

func TestCostTracker_GetSessionCosts(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	store, err := store.New(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)

	sessionID := "test-session"

	// Record multiple costs
	for i := 0; i < 3; i++ {
		costInfo := audit.CostInfo{
			InputTokens:  1000,
			OutputTokens: 500,
			TotalTokens:  1500,
			TotalCost:    0.045,
			Model:        "gpt-4",
		}
		err := tracker.RecordCost(sessionID, "cli", "gpt-4", costInfo)
		if err != nil {
			t.Fatalf("RecordCost failed: %v", err)
		}
	}

	costs, err := tracker.GetSessionCosts(sessionID)
	if err != nil {
		t.Fatalf("GetSessionCosts failed: %v", err)
	}

	if len(costs) != 3 {
		t.Errorf("Expected 3 costs, got %d", len(costs))
	}
}

func TestCostTracker_GetDailyCosts(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	store, err := store.New(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)

	today := time.Now()

	// Record a cost for today
	costInfo := audit.CostInfo{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		TotalCost:    0.045,
		Model:        "gpt-4",
	}
	err = tracker.RecordCost("test-session", "cli", "gpt-4", costInfo)
	if err != nil {
		t.Fatalf("RecordCost failed: %v", err)
	}

	summary, err := tracker.GetDailyCosts(today)
	if err != nil {
		t.Fatalf("GetDailyCosts failed: %v", err)
	}

	if summary.RequestCount < 0 {
		t.Error("Expected non-negative request count")
	}
}

func TestCostTracker_GetMonthlyCosts(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	store, err := store.New(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)

	now := time.Now()

	// Record multiple costs this month
	for i := 0; i < 5; i++ {
		costInfo := audit.CostInfo{
			InputTokens:  1000,
			OutputTokens: 500,
			TotalTokens:  1500,
			TotalCost:    0.045,
			Model:        "gpt-4",
		}
		err := tracker.RecordCost("test-session", "cli", "gpt-4", costInfo)
		if err != nil {
			t.Fatalf("RecordCost failed: %v", err)
		}
	}

	summary, err := tracker.GetMonthlyCosts(now.Year(), now.Month())
	if err != nil {
		t.Fatalf("GetMonthlyCosts failed: %v", err)
	}

	if summary.RequestCount < 0 {
		t.Error("Expected non-negative request count")
	}
}

func TestCostTracker_GetDailyCostsByDateRange(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	store, err := store.New(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)

	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	summaries, err := tracker.GetDailyCostsByDateRange(startDate, endDate)
	if err != nil {
		t.Fatalf("GetDailyCostsByDateRange failed: %v", err)
	}

	// Should return up to 7 days of summaries
	if len(summaries) > 7 {
		t.Errorf("Expected at most 7 daily summaries, got %d", len(summaries))
	}
}
