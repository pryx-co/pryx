package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pryx-core/internal/auth"
	"pryx-core/internal/memory"
	"pryx-core/internal/skills"
	"pryx-core/internal/validation"

	"github.com/go-chi/chi/v5"
)

// handleHealth returns a simple health check response.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	activeProvider := strings.TrimSpace(s.cfg.ModelProvider)
	configuredProviders := []string{}
	cloudLoggedIn := false
	if s.keychain != nil {
		if token, err := s.keychain.Get("cloud_access_token"); err == nil && strings.TrimSpace(token) != "" {
			cloudLoggedIn = true
		}
	}

	switch activeProvider {
	case "":
	case "ollama":
		if strings.TrimSpace(s.cfg.OllamaEndpoint) != "" {
			configuredProviders = append(configuredProviders, "ollama")
		}
	default:
		if s.keychain != nil {
			if key, err := s.keychain.GetProviderKey(activeProvider); err == nil && strings.TrimSpace(key) != "" {
				configuredProviders = append(configuredProviders, activeProvider)
			}
			if activeProvider == "google" {
				if token, err := s.keychain.Get("oauth_google_access"); err == nil && strings.TrimSpace(token) != "" {
					configuredProviders = append(configuredProviders, activeProvider)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":          "ok",
		"providers":       configuredProviders,
		"cloud_logged_in": cloudLoggedIn,
	})
}

func (s *Server) handleCloudStatus(w http.ResponseWriter, r *http.Request) {
	if s.keychain == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "keychain not available"})
		return
	}

	token, err := s.keychain.Get("cloud_access_token")
	if err != nil {
		token = ""
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"logged_in": strings.TrimSpace(token) != "",
	})
}

func (s *Server) handleCloudLoginStart(w http.ResponseWriter, r *http.Request) {
	apiUrl := strings.TrimSpace(s.cfg.CloudAPIUrl)
	if apiUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "missing cloud api url"})
		return
	}

	// Use PKCE-enabled device flow for enhanced security (RFC 7636)
	res, pkce, err := auth.StartDeviceFlowWithPKCE(apiUrl)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	// Store PKCE parameters temporarily (they'll be used during polling)
	// In production, store in session or encrypted cookie
	s.mu.Lock()
	if s.pkceParams == nil {
		s.pkceParams = make(map[string]*auth.PKCEParams)
	}
	s.pkceParams[res.DeviceCode] = pkce
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

type cloudLoginPollRequest struct {
	DeviceCode string `json:"device_code"`
	Interval   int    `json:"interval"`
	ExpiresIn  int    `json:"expires_in"`
}

func (s *Server) handleCloudLoginPoll(w http.ResponseWriter, r *http.Request) {
	if s.keychain == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "keychain not available"})
		return
	}

	apiUrl := strings.TrimSpace(s.cfg.CloudAPIUrl)
	if apiUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "missing cloud api url"})
		return
	}

	req := cloudLoginPollRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid json body"})
		return
	}
	deviceCode := strings.TrimSpace(req.DeviceCode)
	if deviceCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "missing device_code"})
		return
	}

	timeoutSeconds := req.ExpiresIn
	if timeoutSeconds <= 0 || timeoutSeconds > 1800 {
		timeoutSeconds = 600
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Retrieve PKCE parameters for this device code
	s.mu.Lock()
	pkce := s.pkceParams[deviceCode]
	// Clean up PKCE params after use
	delete(s.pkceParams, deviceCode)
	s.mu.Unlock()

	var token *auth.TokenResponse
	var err error

	if pkce != nil {
		// Use PKCE-enabled polling
		token, err = auth.PollForTokenWithPKCE(ctx, apiUrl, deviceCode, req.Interval, pkce.CodeVerifier)
	} else {
		// Fallback to legacy polling (for backwards compatibility)
		token, err = auth.PollForTokenWithContext(ctx, apiUrl, deviceCode, req.Interval)
	}

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			w.WriteHeader(http.StatusRequestTimeout)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "login timed out"})
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	if err := s.keychain.Set("cloud_access_token", token.AccessToken); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "failed to store token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

// handleMCPTools returns the list of available MCP tools.
func (s *Server) handleMCPTools(w http.ResponseWriter, r *http.Request) {
	refresh := strings.TrimSpace(r.URL.Query().Get("refresh")) == "1"
	tools, err := s.mcp.ListToolsFlat(r.Context(), refresh)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": err.Error(),
		})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
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
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid json body",
		})
		return
	}

	validator := validation.NewValidator()

	if err := validator.ValidateSessionID(req.SessionID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": err.Error(),
		})
		return
	}

	if err := validator.ValidateToolName(req.Tool); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": err.Error(),
		})
		return
	}

	if req.Arguments == nil {
		req.Arguments = map[string]interface{}{}
	}

	if err := validator.ValidateMap("arguments", req.Arguments); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": err.Error(),
		})
		return
	}

	res, err := s.mcp.CallTool(r.Context(), strings.TrimSpace(req.SessionID), req.Tool, req.Arguments)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{
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

type skillActionRequest struct {
	ID string `json:"id"`
}

func (s *Server) handleSkillsEnable(w http.ResponseWriter, r *http.Request) {
	req := skillActionRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid json body"})
		return
	}
	id := strings.TrimSpace(req.ID)

	validator := validation.NewValidator()
	if err := validator.ValidateID("id", id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	reg := s.skills
	if reg == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "skills registry not available"})
		return
	}
	if _, ok := reg.Get(id); !ok {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
		return
	}

	configPath := skills.EnabledConfigPath()
	enabledCfg, err := skills.LoadEnabledConfig(configPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	enabledCfg.EnabledSkills[id] = true
	if err := skills.SaveEnabledConfig(configPath, enabledCfg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	reg.Enable(id)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

func (s *Server) handleSkillsDisable(w http.ResponseWriter, r *http.Request) {
	req := skillActionRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid json body"})
		return
	}
	id := strings.TrimSpace(req.ID)

	validator := validation.NewValidator()
	if err := validator.ValidateID("id", id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	reg := s.skills
	if reg == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "skills registry not available"})
		return
	}
	if _, ok := reg.Get(id); !ok {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
		return
	}

	configPath := skills.EnabledConfigPath()
	enabledCfg, err := skills.LoadEnabledConfig(configPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	delete(enabledCfg.EnabledSkills, id)
	if err := skills.SaveEnabledConfig(configPath, enabledCfg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	reg.Disable(id)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

func (s *Server) handleSkillsInstall(w http.ResponseWriter, r *http.Request) {
	req := skillActionRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid json body"})
		return
	}
	id := strings.TrimSpace(req.ID)
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "id is required"})
		return
	}

	reg := s.skills
	if reg == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "skills registry not available"})
		return
	}

	if existing, ok := reg.Get(id); ok {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "skill": existing})
		return
	}

	if strings.HasPrefix(id, "http://") || strings.HasPrefix(id, "https://") {
		opts := skills.DefaultOptions()
		res, err := skills.InstallFromURL(r.Context(), id, opts)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}

		reg.Upsert(res.Skill)

		configPath := skills.EnabledConfigPath()
		enabledCfg, err := skills.LoadEnabledConfig(configPath)
		if err == nil {
			enabledCfg.EnabledSkills[res.Skill.ID] = true
			_ = skills.SaveEnabledConfig(configPath, enabledCfg)
			reg.Enable(res.Skill.ID)
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "skill": res.Skill})
		return
	}

	validator := validation.NewValidator()
	if err := validator.ValidateID("id", id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
}

func (s *Server) handleSkillsUninstall(w http.ResponseWriter, r *http.Request) {
	req := skillActionRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid json body"})
		return
	}
	id := strings.TrimSpace(req.ID)

	validator := validation.NewValidator()
	if err := validator.ValidateID("id", id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	reg := s.skills
	if reg == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "skills registry not available"})
		return
	}
	skill, ok := reg.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
		return
	}

	if skill.Source != skills.SourceRemote && skill.Source != skills.SourceManaged {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "cannot uninstall non-managed skill"})
		return
	}

	opts := skills.DefaultOptions()
	if err := skills.UninstallSkill(id, opts); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	reg.Delete(id)

	configPath := skills.EnabledConfigPath()
	enabledCfg, err := skills.LoadEnabledConfig(configPath)
	if err == nil {
		delete(enabledCfg.EnabledSkills, id)
		_ = skills.SaveEnabledConfig(configPath, enabledCfg)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
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

func (s *Server) handleProviderKeyStatus(w http.ResponseWriter, r *http.Request) {
	providerID := strings.TrimSpace(chi.URLParam(r, "id"))

	validator := validation.NewValidator()
	if err := validator.ValidateID("id", providerID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if s.keychain == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "keychain not available"})
		return
	}

	_, err := s.keychain.GetProviderKey(providerID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"provider_id": providerID,
		"configured":  err == nil,
	})
}

func (s *Server) handleProviderKeySet(w http.ResponseWriter, r *http.Request) {
	providerID := strings.TrimSpace(chi.URLParam(r, "id"))

	validator := validation.NewValidator()
	if err := validator.ValidateID("id", providerID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if s.keychain == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "keychain not available"})
		return
	}

	var req struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	key := strings.TrimSpace(req.APIKey)
	if err := validator.ValidateRequired("api_key", key); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := s.keychain.SetProviderKey(providerID, key); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to store key"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":          true,
		"provider_id": providerID,
	})
}

func (s *Server) handleProviderKeyDelete(w http.ResponseWriter, r *http.Request) {
	providerID := strings.TrimSpace(chi.URLParam(r, "id"))

	validator := validation.NewValidator()
	if err := validator.ValidateID("id", providerID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if s.keychain == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "keychain not available"})
		return
	}

	if err := s.keychain.DeleteProviderKey(providerID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to delete key"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
