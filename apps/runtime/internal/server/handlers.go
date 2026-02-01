package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"pryx-core/internal/memory"
	"pryx-core/internal/skills"
	"pryx-core/internal/validation"
)

// handleHealth returns a simple health check response.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleMCPTools returns the list of available MCP tools.
func (s *Server) handleMCPTools(w http.ResponseWriter, r *http.Request) {
	refresh := strings.TrimSpace(r.URL.Query().Get("refresh")) == "1"
	tools, err := s.mcp.ListToolsFlat(r.Context(), refresh)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": tools,
	})
}

// mcpCallRequest represents a request to call an MCP tool.
type mcpCallRequest struct {
	SessionID string                 `json:"session_id"`
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

// handleMCPCall executes an MCP tool call.
func (s *Server) handleMCPCall(w http.ResponseWriter, r *http.Request) {
	req := mcpCallRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid json body",
		})
		return
	}

	validator := validation.NewValidator()

	if err := validator.ValidateSessionID(req.SessionID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if err := validator.ValidateToolName(req.Tool); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if req.Arguments == nil {
		req.Arguments = map[string]interface{}{}
	}

	if err := validator.ValidateMap("arguments", req.Arguments); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	res, err := s.mcp.CallTool(r.Context(), strings.TrimSpace(req.SessionID), req.Tool, req.Arguments)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	_ = json.NewEncoder(w).Encode(res)
}

// handleSkillsList returns the list of available skills.
func (s *Server) handleSkillsList(w http.ResponseWriter, r *http.Request) {
	reg := s.skills
	if reg == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"skills": []skills.Skill{},
		})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"skills": reg.List(),
	})
}

// handleSkillsInfo returns detailed information about a specific skill.
func (s *Server) handleSkillsInfo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	validator := validation.NewValidator()
	if err := validator.ValidateID("id", id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	reg := s.skills
	if reg == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "not found",
		})
		return
	}
	skill, ok := reg.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "not found",
		})
		return
	}
	_ = json.NewEncoder(w).Encode(skill)
}

// handleSkillsBody returns the body/content of a specific skill.
func (s *Server) handleSkillsBody(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	validator := validation.NewValidator()
	if err := validator.ValidateID("id", id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	reg := s.skills
	if reg == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "not found",
		})
		return
	}
	skill, ok := reg.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "not found",
		})
		return
	}
	body, err := skill.Body()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"body": body,
	})
}

// handleProvidersList returns the list of available LLM providers.
func (s *Server) handleProvidersList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.catalog != nil {
		var providers []map[string]interface{}
		for id, info := range s.catalog.Providers {
			requiresKey := len(info.Env) > 0
			providers = append(providers, map[string]interface{}{
				"id":               id,
				"name":             info.Name,
				"requires_api_key": requiresKey,
			})
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"providers": providers})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers": []map[string]interface{}{
			{"id": "openai", "name": "OpenAI", "requires_api_key": true},
			{"id": "anthropic", "name": "Anthropic", "requires_api_key": true},
			{"id": "google", "name": "Google AI", "requires_api_key": true},
			{"id": "ollama", "name": "Ollama (Local)", "requires_api_key": false},
		},
	})
}

// handleProviderModels returns the list of models available for a specific provider.
func (s *Server) handleProviderModels(w http.ResponseWriter, r *http.Request) {
	providerID := strings.TrimSpace(chi.URLParam(r, "id"))

	validator := validation.NewValidator()
	if err := validator.ValidateID("id", providerID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if s.catalog != nil {
		models := s.catalog.GetProviderModels(providerID)
		var result []map[string]interface{}
		for _, m := range models {
			modelData := map[string]interface{}{
				"id":                 m.ID,
				"name":               m.Name,
				"context_window":     m.Limit.Context,
				"max_output_tokens":  m.Limit.Output,
				"supports_tools":     m.ToolCall,
				"supports_vision":    m.SupportsVision(),
				"supports_reasoning": m.Reasoning,
				"input_price_1m":     m.Cost.Input,
				"output_price_1m":    m.Cost.Output,
			}
			result = append(result, modelData)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"models": result})
		return
	}

	staticModels := map[string][]map[string]interface{}{
		"openai": {
			{"id": "gpt-4", "name": "GPT-4"},
			{"id": "gpt-4-turbo", "name": "GPT-4 Turbo"},
			{"id": "gpt-3.5-turbo", "name": "GPT-3.5 Turbo"},
		},
		"anthropic": {
			{"id": "claude-3-opus", "name": "Claude 3 Opus"},
			{"id": "claude-3-sonnet", "name": "Claude 3 Sonnet"},
			{"id": "claude-3-haiku", "name": "Claude 3 Haiku"},
		},
		"google": {
			{"id": "gemini-pro", "name": "Gemini Pro"},
			{"id": "gemini-ultra", "name": "Gemini Ultra"},
		},
		"ollama": {
			{"id": "llama3", "name": "Llama 3"},
			{"id": "llama2", "name": "Llama 2"},
			{"id": "mistral", "name": "Mistral"},
		},
	}

	if providerModels, ok := staticModels[providerID]; ok {
		json.NewEncoder(w).Encode(map[string]interface{}{"models": providerModels})
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "provider not found"})
	}
}

// handleModelsList returns the list of all available LLM models.
func (s *Server) handleModelsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.catalog != nil {
		var result []map[string]interface{}
		for _, m := range s.catalog.Models {
			modelData := map[string]interface{}{
				"id":                 m.ID,
				"name":               m.Name,
				"provider":           m.Provider,
				"context_window":     m.Limit.Context,
				"max_output_tokens":  m.Limit.Output,
				"supports_tools":     m.ToolCall,
				"supports_vision":    m.SupportsVision(),
				"supports_reasoning": m.Reasoning,
				"input_price_1m":     m.Cost.Input,
				"output_price_1m":    m.Cost.Output,
			}
			result = append(result, modelData)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"models": result})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": []map[string]interface{}{
			{"id": "gpt-4", "name": "GPT-4", "provider": "openai"},
			{"id": "gpt-4-turbo", "name": "GPT-4 Turbo", "provider": "openai"},
			{"id": "gpt-3.5-turbo", "name": "GPT-3.5 Turbo", "provider": "openai"},
			{"id": "claude-3-opus", "name": "Claude 3 Opus", "provider": "anthropic"},
			{"id": "claude-3-sonnet", "name": "Claude 3 Sonnet", "provider": "anthropic"},
			{"id": "claude-3-haiku", "name": "Claude 3 Haiku", "provider": "anthropic"},
		},
	})
}

// handleAgentsList returns the list of active spawned agents.
func (s *Server) handleAgentsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.spawnTool == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "spawn tool not available"})
		return
	}

	agents := s.spawnTool.ListAgents()
	json.NewEncoder(w).Encode(map[string]interface{}{"agents": agents})
}

// handleAgentGet returns the status of a specific agent.
func (s *Server) handleAgentGet(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	validator := validation.NewValidator()
	if err := validator.ValidateID("id", agentID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if s.spawnTool == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "spawn tool not available"})
		return
	}

	agent, err := s.spawnTool.GetAgentStatus(agentID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

type spawnRequest struct {
	Task      string `json:"task"`
	Context   string `json:"context,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

func (s *Server) handleAgentSpawn(w http.ResponseWriter, r *http.Request) {
	if s.spawnTool == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "spawn tool not available"})
		return
	}

	var req spawnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if req.Task == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "task is required"})
		return
	}

	params, _ := json.Marshal(req)
	result, err := s.spawnTool.Execute(r.Context(), params, "api")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleAgentCancel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "cancel not yet implemented"})
}

type forkRequest struct {
	SourceSessionID string `json:"source_session_id"`
}

func (s *Server) handleSessionFork(w http.ResponseWriter, r *http.Request) {
	if s.spawnTool == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "spawn tool not available"})
		return
	}

	var req forkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if req.SourceSessionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "source_session_id is required"})
		return
	}

	newSessionID, err := s.spawnTool.ForkSession(req.SourceSessionID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":                newSessionID,
		"source_session_id": req.SourceSessionID,
		"new_session_id":    newSessionID,
	})
}

// === Memory API Handlers ===

// handleMemoryList returns a list of memory entries
func (s *Server) handleMemoryList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.ragMemory == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "memory system not available"})
		return
	}

	memType := r.URL.Query().Get("type")
	date := r.URL.Query().Get("date")
	limit := 100

	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	opts := memory.SearchOptions{
		Type:  memory.MemoryType(memType),
		Date:  date,
		Limit: limit,
	}

	entries, err := s.ragMemory.List(opts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	})
}

// handleMemoryWrite writes a new memory entry
func (s *Server) handleMemoryWrite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.ragMemory == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "memory system not available"})
		return
	}

	var req struct {
		Type    string                `json:"type"`
		Content string                `json:"content"`
		Date    string                `json:"date,omitempty"`
		Sources []memory.MemorySource `json:"sources,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if req.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "content is required"})
		return
	}

	var entryID string
	var err error

	switch req.Type {
	case "daily":
		entryID, err = s.ragMemory.WriteDaily(req.Content, req.Sources)
	case "longterm":
		entryID, err = s.ragMemory.WriteLongterm(req.Content, req.Sources)
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid type, must be 'daily' or 'longterm'"})
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      entryID,
		"type":    req.Type,
		"content": req.Content,
	})
}

// handleMemorySearch searches memory entries
func (s *Server) handleMemorySearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.ragMemory == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "memory system not available"})
		return
	}

	var req struct {
		Query         string `json:"query"`
		Type          string `json:"type,omitempty"`
		Limit         int    `json:"limit,omitempty"`
		IncludeFTS    bool   `json:"include_fts,omitempty"`
		IncludeVector bool   `json:"include_vector,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if req.Query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "query is required"})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	opts := memory.SearchOptions{
		Type:          memory.MemoryType(req.Type),
		Limit:         req.Limit,
		IncludeFTS:    req.IncludeFTS || true,
		IncludeVector: req.IncludeVector,
	}

	results, err := s.ragMemory.Search(r.Context(), req.Query, opts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   req.Query,
		"results": results,
		"count":   len(results),
	})
}
