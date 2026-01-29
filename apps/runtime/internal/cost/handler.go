package cost

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler handles HTTP requests for cost tracking
type Handler struct {
	service *CostService
}

// NewHandler creates a new cost handler
func NewHandler(service *CostService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers cost tracking routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/cost/summary", h.handleGetCostSummary)
	mux.HandleFunc("/api/cost/session/", h.handleGetSessionCost)
	mux.HandleFunc("/api/cost/daily", h.handleGetDailyCost)
	mux.HandleFunc("/api/cost/monthly", h.handleGetMonthlyCost)
	mux.HandleFunc("/api/cost/budget", h.handleGetBudgetStatus)
	mux.HandleFunc("/api/cost/budget", h.handleSetBudget)
	mux.HandleFunc("/api/cost/optimizations", h.handleGetOptimizations)
	mux.HandleFunc("/api/cost/pricing", h.handleGetPricing)
}

// handleGetCostSummary returns overall cost summary
func (h *Handler) handleGetCostSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cost, err := h.service.GetCurrentSessionCost()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cost)
}

// handleGetSessionCost returns cost for a specific session
func (h *Handler) handleGetSessionCost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Path[len("/api/cost/session/"):]
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// This would need to be implemented in the service
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_id": sessionID,
		"status":     "implemented",
	})
}

// handleGetDailyCost returns daily cost
func (h *Handler) handleGetDailyCost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dateStr := r.URL.Query().Get("date")
	var date time.Time
	if dateStr == "" {
		date = time.Now()
	} else {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	cost, err := h.service.tracker.GetDailyCosts(date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cost)
}

// handleGetMonthlyCost returns monthly cost
func (h *Handler) handleGetMonthlyCost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	year := time.Now().Year()
	month := time.Now().Month()

	// Parse optional parameters
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		var err error
		year, err = parseYear(yearStr)
		if err != nil {
			http.Error(w, "Invalid year format", http.StatusBadRequest)
			return
		}
	}
	if monthStr := r.URL.Query().Get("month"); monthStr != "" {
		var err error
		month, err = parseMonth(monthStr)
		if err != nil {
			http.Error(w, "Invalid month format", http.StatusBadRequest)
			return
		}
	}

	cost, err := h.service.tracker.GetMonthlyCosts(year, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cost)
}

// handleGetBudgetStatus returns current budget status
func (h *Handler) handleGetBudgetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "default"
	}

	status := h.service.GetBudgetStatus(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleSetBudget sets a budget for a user
func (h *Handler) handleSetBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config BudgetConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "default"
	}

	h.service.SetBudget(userID, config)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "budget set",
	})
}

// handleGetOptimizations returns cost optimization suggestions
func (h *Handler) handleGetOptimizations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	suggestions := h.service.GetOptimizationSuggestions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

// handleGetPricing returns pricing for all models
func (h *Handler) handleGetPricing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pricing := h.service.GetAllModelPricing()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pricing)
}

// parseYear parses year string to int
func parseYear(s string) (int, error) {
	var year int
	err := json.Unmarshal([]byte(s), &year)
	if err != nil {
		return 0, err
	}
	return year, nil
}

// parseMonth parses month string to time.Month
func parseMonth(s string) (time.Month, error) {
	monthMap := map[string]time.Month{
		"1":  time.January,
		"2":  time.February,
		"3":  time.March,
		"4":  time.April,
		"5":  time.May,
		"6":  time.June,
		"7":  time.July,
		"8":  time.August,
		"9":  time.September,
		"10": time.October,
		"11": time.November,
		"12": time.December,
	}

	if month, ok := monthMap[s]; ok {
		return month, nil
	}

	return time.January, &parseError{"invalid month"}
}

// parseError represents a parsing error
type parseError struct {
	msg string
}

func (e *parseError) Error() string {
	return e.msg
}
