package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/mcp"
	"pryx-core/internal/models"
	"pryx-core/internal/policy"
	"pryx-core/internal/skills"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"nhooyr.io/websocket"
)

type Server struct {
	cfg      *config.Config
	db       *sql.DB
	keychain *keychain.Keychain
	router   *chi.Mux
	bus      *bus.Bus
	mcp      *mcp.Manager
	skills   *skills.Registry
	catalog  *models.Catalog

	httpMu     sync.Mutex
	httpServer *http.Server
}

func New(cfg *config.Config, db *sql.DB, kc *keychain.Keychain) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	p := policy.NewEngine(nil)

	s := &Server{
		cfg:      cfg,
		db:       db,
		keychain: kc,
		router:   r,
		bus:      bus.New(),
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		reg, err := skills.Discover(ctx, skills.DefaultOptions())
		s.skills = reg
		if err != nil {
			s.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, "", map[string]interface{}{
				"kind":  "skills.load_failed",
				"error": err.Error(),
			}))
		} else {
			s.bus.Publish(bus.NewEvent(bus.EventTraceEvent, "", map[string]interface{}{
				"kind":  "skills.loaded",
				"count": len(reg.List()),
			}))
		}
	}

	s.mcp = mcp.NewManager(s.bus, p, kc)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		path, err := s.mcp.LoadAndConnect(ctx)
		if err != nil {
			s.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, "", map[string]interface{}{
				"kind":  "mcp.connect_failed",
				"error": err.Error(),
				"path":  path,
			}))
			return
		}
		if path != "" {
			s.bus.Publish(bus.NewEvent(bus.EventTraceEvent, "", map[string]interface{}{
				"kind": "mcp.connected",
				"path": path,
			}))
		}
	}()

	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ws", s.handleWS)
	s.router.Get("/mcp/tools", s.handleMCPTools)
	s.router.Post("/mcp/tools/call", s.handleMCPCall)
	s.router.Get("/skills", s.handleSkillsList)
	s.router.Get("/skills/{id}", s.handleSkillsInfo)
	s.router.Get("/skills/{id}/body", s.handleSkillsBody)
	s.router.Get("/api/v1/providers", s.handleProvidersList)
	s.router.Get("/api/v1/providers/{id}/models", s.handleProviderModels)
	s.router.Get("/api/v1/models", s.handleModelsList)
}

func (s *Server) Bus() *bus.Bus {
	return s.bus
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Allow all origins for local dev
	})
	if err != nil {
		// Log error to stdout for now
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	query := r.URL.Query()
	surface := strings.TrimSpace(query.Get("surface"))
	sessionFilter := strings.TrimSpace(query.Get("session_id"))
	eventFilters := query["event"]

	var topics []bus.EventType
	for _, ev := range eventFilters {
		ev = strings.TrimSpace(ev)
		if ev == "" {
			continue
		}
		topics = append(topics, bus.EventType(ev))
	}

	var events <-chan bus.Event
	var cancel func()
	if len(topics) == 0 {
		events, cancel = s.bus.Subscribe()
	} else {
		events, cancel = s.bus.Subscribe(topics...)
	}
	defer cancel()

	ctx := r.Context()

	s.bus.Publish(bus.NewEvent(bus.EventTraceEvent, sessionFilter, map[string]interface{}{
		"kind":        "ws.connected",
		"remote_addr": r.RemoteAddr,
		"surface":     surface,
	}))

	// Writer goroutine: Listen to bus, write to WS
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-events:
				if !ok {
					return
				}
				if sessionFilter != "" && evt.SessionID != sessionFilter {
					continue
				}
				bytes, err := json.Marshal(evt)
				if err != nil {
					continue
				}
				if err := c.Write(ctx, websocket.MessageText, bytes); err != nil {
					return
				}
			}
		}
	}()

	// Reader loop: Keep connection alive and handle incoming messages
	for {
		msgType, data, err := c.Read(ctx)
		if err != nil {
			break
		}
		if msgType != websocket.MessageText {
			continue
		}

		// Parse generic message structure
		in := struct {
			Event      string                 `json:"event"`
			Type       string                 `json:"type"`
			SessionID  string                 `json:"session_id"`
			Payload    map[string]interface{} `json:"payload"`
			ApprovalID string                 `json:"approval_id"`
			Approved   bool                   `json:"approved"`
		}{}
		if err := json.Unmarshal(data, &in); err != nil {
			continue
		}

		// Handle different message types
		eventType := in.Event
		if eventType == "" {
			eventType = in.Type
		}

		switch eventType {
		case "approval.resolve":
			if strings.TrimSpace(in.ApprovalID) != "" {
				_ = s.mcp.ResolveApproval(in.ApprovalID, in.Approved)
			}
		case "chat.send":
			if in.Payload != nil && in.Payload["content"] != nil {
				// Publish chat request for Agent to handle
				s.bus.Publish(bus.NewEvent(bus.EventChatRequest, sessionFilter, in.Payload))
			}
		}
	}

	s.bus.Publish(bus.NewEvent(bus.EventTraceEvent, sessionFilter, map[string]interface{}{
		"kind":        "ws.disconnected",
		"remote_addr": r.RemoteAddr,
		"surface":     surface,
	}))

	c.Close(websocket.StatusNormalClosure, "")
}

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

type mcpCallRequest struct {
	SessionID string                 `json:"session_id"`
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

func (s *Server) handleMCPCall(w http.ResponseWriter, r *http.Request) {
	req := mcpCallRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid json body",
		})
		return
	}
	if strings.TrimSpace(req.Tool) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "missing tool",
		})
		return
	}
	if req.Arguments == nil {
		req.Arguments = map[string]interface{}{}
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

func (s *Server) handleSkillsInfo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "missing id",
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

func (s *Server) handleSkillsBody(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "missing id",
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

func (s *Server) Start() error {
	// Find an available port if not explicitly set or if set to :0
	addr := s.cfg.ListenAddr
	var port int

	if addr == ":3000" || addr == ":0" || addr == "" {
		// Dynamic port allocation
		availablePort, err := GetAvailablePort()
		if err != nil {
			return fmt.Errorf("failed to find available port: %w", err)
		}
		port = availablePort
		addr = fmt.Sprintf(":%d", port)

		// Write port to file for clients to discover
		if err := WritePortFile(port); err != nil {
			log.Printf("Warning: failed to write port file: %v", err)
		} else {
			log.Printf("Runtime port written to ~/.pryx/runtime.port: %d", port)
		}

		// Clean up port file on shutdown
		defer func() {
			if err := CleanupPortFile(); err != nil {
				log.Printf("Warning: failed to cleanup port file: %v", err)
			}
		}()
	}

	log.Printf("Starting server on http://localhost%s", addr)
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) Serve(l net.Listener) error {
	s.httpMu.Lock()
	if s.httpServer == nil {
		s.httpServer = &http.Server{
			Handler: s.router,
		}
	}
	srv := s.httpServer
	s.httpMu.Unlock()

	return srv.Serve(l)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.httpMu.Lock()
	srv := s.httpServer
	s.httpMu.Unlock()
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) SetCatalog(catalog *models.Catalog) {
	s.catalog = catalog
}

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

func (s *Server) handleProviderModels(w http.ResponseWriter, r *http.Request) {
	providerID := strings.TrimSpace(chi.URLParam(r, "id"))
	if providerID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing provider id"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if s.catalog != nil {
		models := s.catalog.GetProviderModels(providerID)
		var result []map[string]interface{}
		for _, m := range models {
			result = append(result, map[string]interface{}{
				"id":   m.ID,
				"name": m.Name,
			})
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

func (s *Server) handleModelsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.catalog != nil {
		var result []map[string]interface{}
		for _, m := range s.catalog.Models {
			result = append(result, map[string]interface{}{
				"id":       m.ID,
				"name":     m.Name,
				"provider": m.Provider,
			})
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
