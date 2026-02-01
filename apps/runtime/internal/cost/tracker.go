package cost

import (
	"fmt"
	"time"

	"pryx-core/internal/audit"
)

// CostTracker tracks and stores cost information
type CostTracker struct {
	auditRepo *audit.AuditRepository
	pricing   *PricingManager
}

// NewCostTracker creates a new cost tracker
func NewCostTracker(auditRepo *audit.AuditRepository, pricing *PricingManager) *CostTracker {
	return &CostTracker{
		auditRepo: auditRepo,
		pricing:   pricing,
	}
}

// RecordCost records a cost entry to the audit log
func (t *CostTracker) RecordCost(sessionID, surface, modelID string, usage audit.CostInfo) error {
	entry := &audit.AuditEntry{
		SessionID:   sessionID,
		Surface:     surface,
		Action:      audit.ActionMessageSend,
		Description: fmt.Sprintf("LLM request to %s", modelID),
		Cost:        &usage,
		Success:     true,
	}

	return t.auditRepo.Create(entry)
}

// GetSessionCosts retrieves all cost entries for a session
func (t *CostTracker) GetSessionCosts(sessionID string) ([]CostInfo, error) {
	opts := audit.QueryOptions{
		SessionID: sessionID,
		Limit:     1000,
	}

	entries, err := t.auditRepo.Query(opts)
	if err != nil {
		return nil, err
	}

	var costs []CostInfo
	for _, entry := range entries {
		if entry.Cost != nil {
			costs = append(costs, CostInfo{
				InputTokens:  entry.Cost.InputTokens,
				OutputTokens: entry.Cost.OutputTokens,
				TotalTokens:  entry.Cost.TotalTokens,
				InputCost:    entry.Cost.InputCost,
				OutputCost:   entry.Cost.OutputCost,
				TotalCost:    entry.Cost.TotalCost,
				Model:        entry.Cost.Model,
			})
		}
	}

	return costs, nil
}

// GetDailyCosts retrieves costs for a specific day
func (t *CostTracker) GetDailyCosts(date time.Time) (CostSummary, error) {
	utcDate := date.In(time.UTC)
	startOfDay := time.Date(utcDate.Year(), utcDate.Month(), utcDate.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	opts := audit.QueryOptions{
		StartTime: &startOfDay,
		EndTime:   &endOfDay,
		Limit:     10000,
	}

	return t.queryAggregatedCosts(opts)
}

// GetMonthlyCosts retrieves costs for a specific month
func (t *CostTracker) GetMonthlyCosts(year int, month time.Month) (CostSummary, error) {
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	opts := audit.QueryOptions{
		StartTime: &startOfMonth,
		EndTime:   &endOfMonth,
		Limit:     100000,
	}

	return t.queryAggregatedCosts(opts)
}

// queryAggregatedCosts queries and aggregates costs from audit entries
func (t *CostTracker) queryAggregatedCosts(opts audit.QueryOptions) (CostSummary, error) {
	entries, err := t.auditRepo.Query(opts)
	if err != nil {
		return CostSummary{}, err
	}

	var summary CostSummary
	summary.PeriodStart = time.Now()
	summary.PeriodEnd = time.Now()

	for _, entry := range entries {
		if entry.Cost != nil {
			summary.TotalInputTokens += entry.Cost.InputTokens
			summary.TotalOutputTokens += entry.Cost.OutputTokens
			summary.TotalTokens += entry.Cost.TotalTokens
			summary.TotalInputCost += entry.Cost.InputCost
			summary.TotalOutputCost += entry.Cost.OutputCost
			summary.TotalCost += entry.Cost.TotalCost
			summary.RequestCount++
		}
	}

	if summary.RequestCount > 0 {
		summary.AverageCostPerReq = summary.TotalCost / float64(summary.RequestCount)
	}

	return summary, nil
}

// GetDailyCostsByDateRange retrieves daily costs for a date range
func (t *CostTracker) GetDailyCostsByDateRange(startDate, endDate time.Time) ([]CostSummary, error) {
	var summaries []CostSummary

	currentDate := startDate
	for currentDate.Before(endDate) {
		dayCost, err := t.GetDailyCosts(currentDate)
		if err != nil {
			return nil, err
		}

		if dayCost.RequestCount > 0 {
			summaries = append(summaries, dayCost)
		}

		currentDate = currentDate.Add(24 * time.Hour)
	}

	return summaries, nil
}
