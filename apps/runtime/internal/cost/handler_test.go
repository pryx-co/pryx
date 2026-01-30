package cost

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pryx-core/internal/audit"
	"pryx-core/internal/store"
)

func setupTestHandler(t *testing.T) (*Handler, *store.Store) {
	tmpDB := t.TempDir() + "/test.db"
	store, err := store.New(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	auditRepo := audit.NewAuditRepository(store.DB)
	pricingMgr := NewPricingManager()
	tracker := NewCostTracker(auditRepo, pricingMgr)
	calculator := NewCostCalculator(pricingMgr)
	service := NewCostService(tracker, calculator, pricingMgr, store)

	handler := NewHandler(service)
	return handler, store
}

func TestHandler_GetCostSummary(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/cost/summary", nil)
	rec := httptest.NewRecorder()

	handler.handleGetCostSummary(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestHandler_GetCostSummary_MethodNotAllowed(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/cost/summary", nil)
	rec := httptest.NewRecorder()

	handler.handleGetCostSummary(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rec.Code)
	}
}

func TestHandler_GetSessionCost(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/cost/session/test-session", nil)
	rec := httptest.NewRecorder()

	handler.handleGetSessionCost(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestHandler_GetSessionCost_MissingID(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/cost/session/", nil)
	rec := httptest.NewRecorder()

	handler.handleGetSessionCost(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestHandler_GetDailyCost(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/cost/daily", nil)
	rec := httptest.NewRecorder()

	handler.handleGetDailyCost(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestHandler_GetDailyCost_WithDate(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/cost/daily?date=2024-01-15", nil)
	rec := httptest.NewRecorder()

	handler.handleGetDailyCost(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestHandler_GetDailyCost_InvalidDate(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/cost/daily?date=invalid-date", nil)
	rec := httptest.NewRecorder()

	handler.handleGetDailyCost(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestHandler_GetMonthlyCost(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/cost/monthly", nil)
	rec := httptest.NewRecorder()

	handler.handleGetMonthlyCost(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestHandler_GetMonthlyCost_WithYearMonth(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/cost/monthly?year=2024&month=1", nil)
	rec := httptest.NewRecorder()

	handler.handleGetMonthlyCost(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestHandler_GetBudgetStatus(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/cost/budget", nil)
	rec := httptest.NewRecorder()

	handler.handleGetBudgetStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestHandler_GetBudgetStatus_WithUserID(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/cost/budget?user_id=testuser", nil)
	rec := httptest.NewRecorder()

	handler.handleGetBudgetStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestHandler_SetBudget(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	config := BudgetConfig{
		DailyBudget:      10.0,
		MonthlyBudget:    100.0,
		WarningThreshold: 80.0,
	}

	body, _ := json.Marshal(config)
	req := httptest.NewRequest(http.MethodPost, "/api/cost/budget", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.handleSetBudget(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestHandler_SetBudget_InvalidBody(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/cost/budget", bytes.NewReader([]byte("invalid json")))
	rec := httptest.NewRecorder()

	handler.handleSetBudget(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestHandler_SetBudget_MethodNotAllowed(t *testing.T) {
	handler, store := setupTestHandler(t)
	defer store.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/cost/budget", nil)
	rec := httptest.NewRecorder()

	handler.handleSetBudget(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rec.Code)
	}
}
