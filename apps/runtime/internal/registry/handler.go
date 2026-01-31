package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"pryx-core/internal/registry"
)

// Handler wraps the registry service for HTTP handlers
type Handler struct {
	service *registry.Service
}

// NewHandler creates a new handler for agent registry API
func NewHandler(service *registry.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterHandler handles POST /api/v1/agents requests
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var agent registry.Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate agent
	if err := registry.ValidateAgent(&agent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Register agent
	if err := h.service.Register(r.Context(), &agent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return registered agent
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(agent)
}

// GetHandler handles GET /api/v1/agents/{id} requests
func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	agent, err := h.service.Get(agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// ListHandler handles GET /api/v1/agents requests
func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	agents := h.service.List()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": agents,
	})
}

// UnregisterHandler handles DELETE /api/v1/agents/{id} requests
func (h *Handler) UnregisterHandler(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	if err := h.service.Unregister(agentID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DiscoverHandler handles GET /api/v1/agents/discover requests
func (h *Handler) DiscoverHandler(w http.ResponseWriter, r *http.Request) {
	var criteria registry.DiscoveryCriteria
	if err := json.NewDecoder(r.Body).Decode(&criteria); err != nil && r.Body != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	agents, err := h.service.Discover(criteria)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": agents,
	})
}

// HealthHandler handles PUT /api/v1/agents/{id}/health requests
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	var payload struct {
		Status registry.HealthStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateHealth(agentID, payload.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetupRoutes registers all agent registry routes
func SetupRoutes(router chi.Router, handler *Handler) {
	router.Post("/api/v1/agents", handler.RegisterHandler)
	router.Get("/api/v1/agents/{id}", handler.GetHandler)
	router.Delete("/api/v1/agents/{id}", handler.UnregisterHandler)
	router.Get("/api/v1/agents", handler.ListHandler)
	router.Get("/api/v1/agents/discover", handler.DiscoverHandler)
	router.Put("/api/v1/agents/{id}/health", handler.HealthHandler)
}
