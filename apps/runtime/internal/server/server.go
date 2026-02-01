package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"pryx-core/internal/agentbus"
	"pryx-core/internal/bus"
	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/mcp"
	"pryx-core/internal/mcp/discovery"
	"pryx-core/internal/memory"
	"pryx-core/internal/models"
	"pryx-core/internal/policy"
	"pryx-core/internal/skills"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type SpawnTool interface {
	Name() string
	Description() string
	Schema() map[string]interface{}
	Execute(ctx context.Context, params json.RawMessage, parentID string) (interface{}, error)
	GetAgentStatus(agentID string) (map[string]interface{}, error)
	ListAgents() []map[string]interface{}
	ForkSession(sourceSessionID string) (string, error)
}

type Server struct {
	cfg          *config.Config
	db           *sql.DB
	keychain     *keychain.Keychain
	router       *chi.Mux
	bus          *bus.Bus
	agentbus     *agentbus.Service
	mcp          *mcp.Manager
	mcpDiscovery *discovery.DiscoveryService
	skills       *skills.Registry
	catalog      *models.Catalog
	spawnTool    SpawnTool
	ragMemory    *memory.RAGManager

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

	// Initialize agentbus (agent connectivity hub)
	s.agentbus = agentbus.NewService(s.bus, agentbus.HubConfig{
		Name:               "pryx-agentbus",
		Namespace:          "default",
		LogLevel:           "info",
		AutoDetectEnabled:  cfg.AgentDetectEnabled,
		AutoDetectInterval: cfg.AgentDetectInterval,
		PackageDir:         cfg.SkillsPath,
		CacheDir:           cfg.CachePath,
		MaxConnections:     20,
		ReconnectEnabled:   true,
	})

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.agentbus.Start(ctx); err != nil {
			s.bus.Publish(bus.NewEvent(bus.EventErrorOccurred, "", map[string]interface{}{
				"kind":  "agentbus.start_failed",
				"error": err.Error(),
			}))
			return
		}
		s.bus.Publish(bus.NewEvent(bus.EventTraceEvent, "", map[string]interface{}{
			"kind": "agentbus.started",
		}))
	}()
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

	s.ragMemory = memory.NewRAGManager(db, cfg.MemoryEnabled)
	log.Printf("RAG Memory system initialized (enabled: %v)", cfg.MemoryEnabled)

	return s
}

func (s *Server) routes() {
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ws", s.handleWS)
	s.router.Get("/mcp/tools", s.handleMCPTools)
	s.router.Post("/mcp/tools/call", s.handleMCPCall)
	s.router.Get("/mcp/discovery/curated", s.handleMCPDiscoveryCurated)
	s.router.Get("/mcp/discovery/categories", s.handleMCPDiscoveryCategories)
	s.router.Get("/mcp/discovery/curated/{id}", s.handleMCPDiscoveryServer)
	s.router.Post("/mcp/discovery/validate", s.handleMCPDiscoveryValidateURL)
	s.router.Post("/mcp/discovery/custom", s.handleMCPDiscoveryAddCustom)
	s.router.Get("/mcp/discovery/custom", s.handleMCPDiscoveryCustomServers)
	s.router.Delete("/mcp/discovery/custom/{id}", s.handleMCPDiscoveryRemoveCustom)
	s.router.Get("/skills", s.handleSkillsList)
	s.router.Get("/skills/{id}", s.handleSkillsInfo)
	s.router.Get("/skills/{id}/body", s.handleSkillsBody)
	s.router.Get("/api/v1/providers", s.handleProvidersList)
	s.router.Get("/api/v1/providers/{id}/models", s.handleProviderModels)
	s.router.Get("/api/v1/models", s.handleModelsList)
	s.router.Get("/api/v1/agents", s.handleAgentsList)
	s.router.Get("/api/v1/agents/{id}", s.handleAgentGet)
	s.router.Post("/api/v1/agents/spawn", s.handleAgentSpawn)
	s.router.Post("/api/v1/agents/{id}/cancel", s.handleAgentCancel)
	s.router.Get("/api/v1/sessions", s.handleSessionsList)
	s.router.Post("/api/v1/sessions", s.handleSessionCreate)
	s.router.Get("/api/v1/sessions/{id}", s.handleSessionGet)
	s.router.Delete("/api/v1/sessions/{id}", s.handleSessionDelete)
	s.router.Post("/api/v1/sessions/fork", s.handleSessionFork)

	s.router.Get("/api/v1/memory", s.handleMemoryList)
	s.router.Post("/api/v1/memory", s.handleMemoryWrite)
	s.router.Post("/api/v1/memory/search", s.handleMemorySearch)
}

func (s *Server) Bus() *bus.Bus {
	return s.bus
}

func (s *Server) Skills() *skills.Registry {
	return s.skills
}

func (s *Server) MCP() *mcp.Manager {
	return s.mcp
}

func (s *Server) Agents() *agentbus.Service {
	return s.agentbus
}

func (s *Server) Memory() *memory.RAGManager {
	return s.ragMemory
}

func (s *Server) Handler() http.Handler {
	return s.router
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

func (s *Server) SetCatalog(catalog *models.Catalog) {
	s.catalog = catalog
}

func (s *Server) SetSpawnTool(tool SpawnTool) {
	s.spawnTool = tool
}
