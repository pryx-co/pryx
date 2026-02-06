package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"pryx-core/internal/scheduler"

	"github.com/go-chi/chi/v5"
)

// Scheduled task request/response types
type CreateTaskRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	CronExpression string `json:"cron_expression"`
	TaskType       string `json:"task_type"`
	Payload        string `json:"payload"`
	Timezone       string `json:"timezone"`
	Enabled        bool   `json:"enabled"`
}

type UpdateTaskRequest struct {
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	CronExpression *string `json:"cron_expression,omitempty"`
	TaskType       *string `json:"task_type,omitempty"`
	Payload        *string `json:"payload,omitempty"`
	Timezone       *string `json:"timezone,omitempty"`
	Enabled        *bool   `json:"enabled,omitempty"`
}

type TaskResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	CronExpression string     `json:"cron_expression"`
	TaskType       string     `json:"task_type"`
	Payload        string     `json:"payload"`
	Timezone       string     `json:"timezone"`
	Enabled        bool       `json:"enabled"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	LastRunStatus  string     `json:"last_run_status,omitempty"`
	NextRunAt      *time.Time `json:"next_run_at,omitempty"`
	RunCount       int        `json:"run_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type RunResponse struct {
	ID          string     `json:"id"`
	TaskID      string     `json:"task_id"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Status      string     `json:"status"`
	Error       string     `json:"error,omitempty"`
	Output      string     `json:"output,omitempty"`
}

// handleTasksList returns all scheduled tasks for the user
func (s *Server) handleTasksList(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	tasks, err := s.scheduler.ListTasks(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list tasks: %v", err), http.StatusInternalServerError)
		return
	}

	response := make([]TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		response = append(response, TaskResponse{
			ID:             task.ID,
			Name:           task.Name,
			Description:    task.Description,
			CronExpression: task.CronExpression,
			TaskType:       string(task.TaskType),
			Payload:        task.Payload,
			Timezone:       task.Timezone,
			Enabled:        task.Enabled,
			LastRunAt:      task.LastRunAt,
			LastRunStatus:  task.LastRunStatus,
			NextRunAt:      task.NextRunAt,
			RunCount:       task.RunCount,
			CreatedAt:      task.CreatedAt,
			UpdatedAt:      task.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleTaskGet returns a single scheduled task
func (s *Server) handleTaskGet(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	task, err := s.scheduler.GetTask(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get task: %v", err), http.StatusInternalServerError)
		return
	}
	if task == nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	response := TaskResponse{
		ID:             task.ID,
		Name:           task.Name,
		Description:    task.Description,
		CronExpression: task.CronExpression,
		TaskType:       string(task.TaskType),
		Payload:        task.Payload,
		Timezone:       task.Timezone,
		Enabled:        task.Enabled,
		LastRunAt:      task.LastRunAt,
		LastRunStatus:  task.LastRunStatus,
		NextRunAt:      task.NextRunAt,
		RunCount:       task.RunCount,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleTaskCreate creates a new scheduled task
func (s *Server) handleTaskCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate cron expression
	if err := scheduler.ValidateCronExpression(req.CronExpression); err != nil {
		http.Error(w, fmt.Sprintf("Invalid cron expression: %v", err), http.StatusBadRequest)
		return
	}

	// Validate task type
	validTypes := map[string]bool{
		string(scheduler.TaskTypeMessage):  true,
		string(scheduler.TaskTypeWorkflow): true,
		string(scheduler.TaskTypeReminder): true,
		string(scheduler.TaskTypeWebhook):  true,
	}
	if !validTypes[req.TaskType] {
		http.Error(w, fmt.Sprintf("Invalid task type: %s", req.TaskType), http.StatusBadRequest)
		return
	}

	task := &scheduler.ScheduledTask{
		Name:           req.Name,
		Description:    req.Description,
		CronExpression: req.CronExpression,
		TaskType:       scheduler.TaskType(req.TaskType),
		Payload:        req.Payload,
		Timezone:       req.Timezone,
		Enabled:        req.Enabled,
	}

	if err := s.scheduler.CreateTask(task); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create task: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TaskResponse{
		ID:             task.ID,
		Name:           task.Name,
		Description:    task.Description,
		CronExpression: task.CronExpression,
		TaskType:       string(task.TaskType),
		Payload:        task.Payload,
		Timezone:       task.Timezone,
		Enabled:        task.Enabled,
		NextRunAt:      task.NextRunAt,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	})
}

// handleTaskUpdate updates an existing scheduled task
func (s *Server) handleTaskUpdate(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	task, err := s.scheduler.GetTask(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get task: %v", err), http.StatusInternalServerError)
		return
	}
	if task == nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Apply updates
	if req.Name != nil {
		task.Name = *req.Name
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.CronExpression != nil {
		if err := scheduler.ValidateCronExpression(*req.CronExpression); err != nil {
			http.Error(w, fmt.Sprintf("Invalid cron expression: %v", err), http.StatusBadRequest)
			return
		}
		task.CronExpression = *req.CronExpression
	}
	if req.TaskType != nil {
		task.TaskType = scheduler.TaskType(*req.TaskType)
	}
	if req.Payload != nil {
		task.Payload = *req.Payload
	}
	if req.Timezone != nil {
		task.Timezone = *req.Timezone
	}
	if req.Enabled != nil {
		task.Enabled = *req.Enabled
	}

	if err := s.scheduler.UpdateTask(task); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update task: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TaskResponse{
		ID:             task.ID,
		Name:           task.Name,
		Description:    task.Description,
		CronExpression: task.CronExpression,
		TaskType:       string(task.TaskType),
		Payload:        task.Payload,
		Timezone:       task.Timezone,
		Enabled:        task.Enabled,
		LastRunAt:      task.LastRunAt,
		LastRunStatus:  task.LastRunStatus,
		NextRunAt:      task.NextRunAt,
		RunCount:       task.RunCount,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	})
}

// handleTaskDelete deletes a scheduled task
func (s *Server) handleTaskDelete(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	if err := s.scheduler.DeleteTask(taskID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete task: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleTaskEnable enables a scheduled task
func (s *Server) handleTaskEnable(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	if err := s.scheduler.EnableTask(taskID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to enable task: %v", err), http.StatusInternalServerError)
		return
	}

	task, _ := s.scheduler.GetTask(taskID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TaskResponse{
		ID:             task.ID,
		Name:           task.Name,
		Description:    task.Description,
		CronExpression: task.CronExpression,
		TaskType:       string(task.TaskType),
		Payload:        task.Payload,
		Timezone:       task.Timezone,
		Enabled:        true,
		NextRunAt:      task.NextRunAt,
		UpdatedAt:      task.UpdatedAt,
	})
}

// handleTaskDisable disables a scheduled task
func (s *Server) handleTaskDisable(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	if err := s.scheduler.DisableTask(taskID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to disable task: %v", err), http.StatusInternalServerError)
		return
	}

	task, _ := s.scheduler.GetTask(taskID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TaskResponse{
		ID:             task.ID,
		Name:           task.Name,
		Description:    task.Description,
		CronExpression: task.CronExpression,
		TaskType:       string(task.TaskType),
		Payload:        task.Payload,
		Timezone:       task.Timezone,
		Enabled:        false,
		UpdatedAt:      task.UpdatedAt,
	})
}

// handleTaskRuns returns the execution history for a task
func (s *Server) handleTaskRuns(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	runs, err := s.scheduler.GetTaskRuns(taskID, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get task runs: %v", err), http.StatusInternalServerError)
		return
	}

	response := make([]RunResponse, 0, len(runs))
	for _, run := range runs {
		response = append(response, RunResponse{
			ID:          run.ID,
			TaskID:      run.TaskID,
			StartedAt:   run.StartedAt,
			CompletedAt: run.CompletedAt,
			Status:      string(run.Status),
			Error:       run.Error,
			Output:      run.Output,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleTaskValidate validates a cron expression
func (s *Server) handleTaskValidate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CronExpression string `json:"cron_expression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if err := scheduler.ValidateCronExpression(req.CronExpression); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   false,
			"error":   err.Error(),
			"message": "Invalid cron expression",
		})
		return
	}

	nextRuns, err := scheduler.PreviewNextRuns(req.CronExpression, 5)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   false,
			"error":   err.Error(),
			"message": "Invalid schedule expression",
		})
		return
	}

	next5 := make([]string, 0, 5)
	for _, next := range nextRuns {
		next5 = append(next5, next.Format(time.RFC3339))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":     true,
		"message":   "Valid schedule expression",
		"next_runs": next5,
	})
}

// handleTaskEventTrigger triggers tasks for an event-based schedule.
func (s *Server) handleTaskEventTrigger(w http.ResponseWriter, r *http.Request) {
	eventName := chi.URLParam(r, "event")
	triggered, err := s.scheduler.TriggerEvent(eventName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to trigger event tasks: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event":       eventName,
		"triggered":   triggered,
		"message":     "Event tasks triggered",
		"triggeredAt": time.Now().Format(time.RFC3339),
	})
}
