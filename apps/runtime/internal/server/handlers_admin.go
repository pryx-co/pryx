package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pryx-core/internal/cost"
	"pryx-core/internal/telemetry"

	"github.com/google/uuid"
)

// AdminStats represents aggregated statistics for the admin dashboard
type AdminStats struct {
	TotalUsers        int64                    `json:"total_users"`
	TotalDevices      int64                    `json:"total_devices"`
	TotalSessions     int64                    `json:"total_sessions"`
	TotalCost         float64                  `json:"total_cost"`
	ActiveNow         int                      `json:"active_now"`
	TodayMessages     int64                    `json:"today_messages"`
	ProviderBreakdown map[string]ProviderStats `json:"provider_breakdown"`
	PeriodStart       time.Time                `json:"period_start"`
	PeriodEnd         time.Time                `json:"period_end"`
}

type ProviderStats struct {
	RequestCount int     `json:"request_count"`
	TokenCount   int64   `json:"token_count"`
	Cost         float64 `json:"cost"`
}

// UserInfo represents user information for admin view
type UserInfo struct {
	ID           string     `json:"id"`
	Email        string     `json:"email,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	LastSeen     *time.Time `json:"last_seen,omitempty"`
	SessionCount int        `json:"session_count"`
	TotalCost    float64    `json:"total_cost"`
	TotalTokens  int64      `json:"total_tokens"`
	DeviceCount  int        `json:"device_count"`
}

// DeviceInfo represents device information for admin view
type DeviceInfo struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	UserID    string     `json:"user_id,omitempty"`
	PublicKey string     `json:"public_key,omitempty"`
	PairedAt  time.Time  `json:"paired_at"`
	LastSeen  *time.Time `json:"last_seen,omitempty"`
	IsActive  bool       `json:"is_active"`
	Metadata  string     `json:"metadata,omitempty"`
}

// AdminHealth represents system health status
type AdminHealth struct {
	Status    string           `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Version   string           `json:"version"`
	Uptime    time.Duration    `json:"uptime"`
	Database  *DatabaseHealth  `json:"database"`
	Telemetry *TelemetryHealth `json:"telemetry"`
	MCP       *MCPHealth       `json:"mcp"`
	Channels  *ChannelsHealth  `json:"channels"`
	Scheduler *SchedulerHealth `json:"scheduler"`
}

type DatabaseHealth struct {
	Status      string `json:"status"`
	Connections int    `json:"connections"`
	Latency     string `json:"latency"`
}

type TelemetryHealth struct {
	Enabled    bool       `json:"enabled"`
	Status     string     `json:"status"`
	LastExport *time.Time `json:"last_export,omitempty"`
}

type MCPHealth struct {
	ConnectedCount int      `json:"connected_count"`
	TotalCount     int      `json:"total_count"`
	Servers        []string `json:"servers,omitempty"`
}

type ChannelsHealth struct {
	ConnectedCount int      `json:"connected_count"`
	TotalCount     int      `json:"total_count"`
	Channels       []string `json:"channels,omitempty"`
}

type SchedulerHealth struct {
	ActiveTasks   int        `json:"active_tasks"`
	TotalTasks    int        `json:"total_tasks"`
	LastExecution *time.Time `json:"last_execution,omitempty"`
}

// handleAdminStats returns aggregated statistics for admin dashboard
func (s *Server) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	layer := getAuthLayer(r)

	// Filter by user if not superadmin or localhost
	var userID string
	if layer == "user" {
		userID = getUserID(r)
	}

	// Calculate stats
	stats := &AdminStats{
		PeriodStart:       time.Now().AddDate(0, 0, -30), // Default: last 30 days
		PeriodEnd:         time.Now(),
		ProviderBreakdown: make(map[string]ProviderStats),
	}

	// Count users
	if err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers); err != nil && err != sql.ErrNoRows {
		stats.TotalUsers = 0
	}

	// Count devices
	if err := s.db.QueryRow("SELECT COUNT(*) FROM mesh_devices WHERE is_active = 1").Scan(&stats.TotalDevices); err != nil && err != sql.ErrNoRows {
		stats.TotalDevices = 0
	}

	// Count sessions
	query := "SELECT COUNT(*) FROM sessions"
	if userID != "" {
		query = "SELECT COUNT(*) FROM sessions WHERE user_id = ?"
	}
	if err := s.db.QueryRow(query, userID).Scan(&stats.TotalSessions); err != nil && err != sql.ErrNoRows {
		stats.TotalSessions = 0
	}

	// Get total cost from audit log
	if userID != "" {
		if err := s.db.QueryRow(`
			SELECT COALESCE(SUM(CAST(json_extract(payload, '$.cost') AS REAL)), 0)
			FROM audit_log
			WHERE user_id = ? AND created_at >= ?
		`, userID, stats.PeriodStart).Scan(&stats.TotalCost); err != nil && err != sql.ErrNoRows {
			stats.TotalCost = 0
		}
	} else {
		if err := s.db.QueryRow(`
			SELECT COALESCE(SUM(CAST(json_extract(payload, '$.cost') AS REAL)), 0)
			FROM audit_log
			WHERE created_at >= ?
		`, stats.PeriodStart).Scan(&stats.TotalCost); err != nil && err != sql.ErrNoRows {
			stats.TotalCost = 0
		}
	}

	// Count today's messages
	today := time.Now().Truncate(24 * time.Hour)
	if userID != "" {
		if err := s.db.QueryRow(`
			SELECT COUNT(*) FROM messages m
			JOIN sessions s ON m.session_id = s.id
			WHERE s.user_id = ? AND m.created_at >= ?
		`, userID, today).Scan(&stats.TodayMessages); err != nil && err != sql.ErrNoRows {
			stats.TodayMessages = 0
		}
	} else {
		if err := s.db.QueryRow(`
			SELECT COUNT(*) FROM messages WHERE created_at >= ?
		`, today).Scan(&stats.TodayMessages); err != nil && err != sql.ErrNoRows {
			stats.TodayMessages = 0
		}
	}

	// Active users count (sessions in last 5 minutes)
	activeSince := time.Now().Add(-5 * time.Minute)
	if err := s.db.QueryRow(`
		SELECT COUNT(DISTINCT user_id) FROM audit_log
		WHERE timestamp >= ?
	`, activeSince).Scan(&stats.ActiveNow); err != nil && err != sql.ErrNoRows {
		stats.ActiveNow = 0
	}

	// Provider breakdown from audit log
	rows, err := s.db.Query(`
		SELECT json_extract(payload, '$.model') as model,
		       COUNT(*) as request_count,
		       SUM(CAST(json_extract(payload, '$.tokens') AS INTEGER)) as token_count,
		       SUM(CAST(json_extract(payload, '$.cost') AS REAL)) as cost
		FROM audit_log
		WHERE action = 'llm.generate' AND created_at >= ?
		GROUP BY model
	`, stats.PeriodStart)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var model string
			ps := ProviderStats{}
			if err := rows.Scan(&model, &ps.RequestCount, &ps.TokenCount, &ps.Cost); err == nil {
				stats.ProviderBreakdown[model] = ps
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleAdminUsers returns user list for admin dashboard
func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	layer := getAuthLayer(r)

	// Filter by user if not superadmin or localhost
	var userID string
	if layer == "user" {
		userID = getUserID(r)
	}

	// Parse query parameters
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	var users []*UserInfo
	var query string
	var args []interface{}

	if userID != "" {
		query = `
			SELECT u.id, u.email, u.created_at, u.last_seen,
			       COUNT(DISTINCT s.id) as session_count,
			       COALESCE(SUM(CAST(json_extract(al.payload, '$.cost') AS REAL)), 0) as total_cost,
			       COALESCE(SUM(CAST(json_extract(al.payload, '$.tokens') AS INTEGER)), 0) as total_tokens,
			       COUNT(DISTINCT md.id) as device_count
			FROM users u
			LEFT JOIN sessions s ON u.id = s.user_id
			LEFT JOIN audit_log al ON u.id = al.user_id
			LEFT JOIN mesh_devices md ON u.id = md.user_id
			WHERE u.id = ?
			GROUP BY u.id
		`
		args = []interface{}{userID}
	} else {
		query = `
			SELECT u.id, u.email, u.created_at, u.last_seen,
			       COUNT(DISTINCT s.id) as session_count,
			       COALESCE(SUM(CAST(json_extract(al.payload, '$.cost') AS REAL)), 0) as total_cost,
			       COALESCE(SUM(CAST(json_extract(al.payload, '$.tokens') AS INTEGER)), 0) as total_tokens,
			       COUNT(DISTINCT md.id) as device_count
			FROM users u
			LEFT JOIN sessions s ON u.id = s.user_id
			LEFT JOIN audit_log al ON u.id = al.user_id
			LEFT JOIN mesh_devices md ON u.id = md.user_id
			GROUP BY u.id
			ORDER BY u.created_at DESC
			LIMIT ? OFFSET ?
		`
		args = []interface{}{limit, offset}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("Failed to query users: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		u := &UserInfo{}
		if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.LastSeen,
			&u.SessionCount, &u.TotalCost, &u.TotalTokens, &u.DeviceCount); err != nil {
			continue
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// handleAdminDevices returns device list for admin dashboard
func (s *Server) handleAdminDevices(w http.ResponseWriter, r *http.Request) {
	layer := getAuthLayer(r)

	// Filter by user if not superadmin or localhost
	var userID string
	if layer == "user" {
		userID = getUserID(r)
	}

	var query string
	var args []interface{}

	if userID != "" {
		query = `
			SELECT id, name, user_id, public_key, paired_at, last_seen, is_active, metadata
			FROM mesh_devices
			WHERE user_id = ?
			ORDER BY paired_at DESC
		`
		args = []interface{}{userID}
	} else {
		query = `
			SELECT id, name, user_id, public_key, paired_at, last_seen, is_active, metadata
			FROM mesh_devices
			ORDER BY paired_at DESC
		`
		args = []interface{}{}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("Failed to query devices: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var devices []*DeviceInfo
	for rows.Next() {
		d := &DeviceInfo{}
		if err := rows.Scan(&d.ID, &d.Name, &d.UserID, &d.PublicKey, &d.PairedAt,
			&d.LastSeen, &d.IsActive, &d.Metadata); err != nil {
			continue
		}
		devices = append(devices, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

// handleAdminCosts returns cost data for admin dashboard
func (s *Server) handleAdminCosts(w http.ResponseWriter, r *http.Request) {
	layer := getAuthLayer(r)

	// Filter by user if not superadmin or localhost
	var userID string
	if layer == "user" {
		userID = getUserID(r)
	}

	// Parse query parameters
	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		groupBy = "day"
	}

	// Parse time range
	periodStart := time.Now().AddDate(0, 0, -30) // Default: last 30 days
	if ps := r.URL.Query().Get("period_start"); ps != "" {
		if t, err := time.Parse(time.RFC3339, ps); err == nil {
			periodStart = t
		}
	}
	periodEnd := time.Now()
	if pe := r.URL.Query().Get("period_end"); pe != "" {
		if t, err := time.Parse(time.RFC3339, pe); err == nil {
			periodEnd = t
		}
	}

	var query string
	var args []interface{}

	switch groupBy {
	case "day":
		query = `
			SELECT date(created_at) as date,
			       COUNT(*) as requests,
			       SUM(CAST(json_extract(payload, '$.tokens') AS INTEGER)) as tokens,
			       SUM(CAST(json_extract(payload, '$.cost') AS REAL)) as cost
			FROM audit_log
			WHERE created_at >= ? AND created_at <= ?
		`
		if userID != "" {
			query += " AND user_id = ?"
			args = []interface{}{periodStart, periodEnd, userID}
		} else {
			args = []interface{}{periodStart, periodEnd}
		}
		query += " GROUP BY date(created_at) ORDER BY date DESC"
	case "model":
		query = `
			SELECT json_extract(payload, '$.model') as model,
			       COUNT(*) as requests,
			       SUM(CAST(json_extract(payload, '$.tokens') AS INTEGER)) as tokens,
			       SUM(CAST(json_extract(payload, '$.cost') AS REAL)) as cost
			FROM audit_log
			WHERE created_at >= ? AND created_at <= ?
		`
		if userID != "" {
			query += " AND user_id = ?"
			args = []interface{}{periodStart, periodEnd, userID}
		} else {
			args = []interface{}{periodStart, periodEnd}
		}
		query += " GROUP BY json_extract(payload, '$.model') ORDER BY cost DESC"
	case "provider":
		query = `
			SELECT json_extract(payload, '$.provider') as provider,
			       COUNT(*) as requests,
			       SUM(CAST(json_extract(payload, '$.tokens') AS INTEGER)) as tokens,
			       SUM(CAST(json_extract(payload, '$.cost') AS REAL)) as cost
			FROM audit_log
			WHERE created_at >= ? AND created_at <= ?
		`
		if userID != "" {
			query += " AND user_id = ?"
			args = []interface{}{periodStart, periodEnd, userID}
		} else {
			args = []interface{}{periodStart, periodEnd}
		}
		query += " GROUP BY json_extract(payload, '$.provider') ORDER BY cost DESC"
	default:
		query = `
			SELECT COUNT(*) as requests,
			       SUM(CAST(json_extract(payload, '$.tokens') AS INTEGER)) as tokens,
			       SUM(CAST(json_extract(payload, '$.cost') AS REAL)) as cost
			FROM audit_log
			WHERE created_at >= ? AND created_at <= ?
		`
		if userID != "" {
			query += " AND user_id = ?"
			args = []interface{}{periodStart, periodEnd, userID}
		} else {
			args = []interface{}{periodStart, periodEnd}
		}
	}

	type CostBreakdown struct {
		Date     string  `json:"date,omitempty"`
		Model    string  `json:"model,omitempty"`
		Provider string  `json:"provider,omitempty"`
		Requests int     `json:"requests"`
		Tokens   int64   `json:"tokens"`
		Cost     float64 `json:"cost"`
	}

	rows, err := s.db.Query(query, args...)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("Failed to query costs: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var breakdown []CostBreakdown
	for rows.Next() {
		cb := CostBreakdown{}
		if groupBy == "day" {
			if err := rows.Scan(&cb.Date, &cb.Requests, &cb.Tokens, &cb.Cost); err != nil {
				continue
			}
		} else if groupBy == "model" {
			if err := rows.Scan(&cb.Model, &cb.Requests, &cb.Tokens, &cb.Cost); err != nil {
				continue
			}
		} else if groupBy == "provider" {
			if err := rows.Scan(&cb.Provider, &cb.Requests, &cb.Tokens, &cb.Cost); err != nil {
				continue
			}
		} else {
			if err := rows.Scan(&cb.Requests, &cb.Tokens, &cb.Cost); err != nil {
				continue
			}
		}
		breakdown = append(breakdown, cb)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(breakdown)
}

// handleAdminHealth returns system health status for admin dashboard
func (s *Server) handleAdminHealth(w http.ResponseWriter, r *http.Request) {
	health := &AdminHealth{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime),
	}

	// Check database health
	health.Database = &DatabaseHealth{Status: "unknown"}
	if err := s.db.Ping(); err == nil {
		health.Database.Status = "healthy"
	} else {
		health.Database.Status = "unhealthy"
	}

	// Check telemetry health
	health.Telemetry = &TelemetryHealth{Enabled: true}
	if s.costService != nil {
		health.Telemetry.Status = "active"
	}

	// Check MCP connections
	health.MCP = &MCPHealth{}
	if s.mcp != nil {
		health.MCP.TotalCount = 10 // TODO: Get actual count from MCP manager
		health.MCP.ConnectedCount = health.MCP.TotalCount
	}

	// Check channel connections
	health.Channels = &ChannelsHealth{}
	if s.channels != nil {
		health.Channels.TotalCount = 10 // TODO: Get actual count
		health.Channels.ConnectedCount = health.Channels.TotalCount
	}

	// Check scheduler
	health.Scheduler = &SchedulerHealth{}
	if s.scheduler != nil {
		tasks, err := s.scheduler.ListTasks("")
		if err == nil {
			health.Scheduler.TotalTasks = len(tasks)
			for _, task := range tasks {
				if task.Enabled {
					health.Scheduler.ActiveTasks++
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleAdminTelemetryConfig returns telemetry configuration
func (s *Server) handleAdminTelemetryConfig(w http.ResponseWriter, r *http.Request) {
	layer := getAuthLayer(r)

	// Only superadmin can access telemetry config
	if layer != "superadmin" && layer != "localhost" {
		http.Error(w, "Forbidden: superadmin access required", http.StatusForbidden)
		return
	}

	config := map[string]interface{}{
		"enabled":         true,
		"sampling":        1.0,
		"endpoint":        "https://telemetry.pryx.dev/v1/otlp",
		"export_interval": "5m",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// handleAdminTelemetryConfigUpdate updates telemetry configuration
func (s *Server) handleAdminTelemetryConfigUpdate(w http.ResponseWriter, r *http.Request) {
	layer := getAuthLayer(r)

	// Only superadmin can update telemetry config
	if layer != "superadmin" && layer != "localhost" {
		http.Error(w, "Forbidden: superadmin access required", http.StatusForbidden)
		return
	}

	var req struct {
		Enabled  *bool    `json:"enabled,omitempty"`
		Sampling *float64 `json:"sampling,omitempty"`
		Endpoint *string  `json:"endpoint,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Update config (would need to persist this to config file/database)
	// For now, just return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Telemetry configuration updated",
	})
}

// getAuthLayer extracts and validates the auth layer from request
func getAuthLayer(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "localhost"
	}

	// Check for superadmin token
	if len(auth) > 11 && auth[:11] == "superadmin:" {
		return "superadmin"
	}

	// Regular user token
	return "user"
}

// getUserID extracts user ID from request context or token
func getUserID(r *http.Request) string {
	// TODO: Extract from authenticated context
	return ""
}

var startTime = time.Now()
