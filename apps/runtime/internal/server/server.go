package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"pryx-core/internal/agentbus"
	"pryx-core/internal/audit"
	"pryx-core/internal/auth"
	"pryx-core/internal/bus"
	"pryx-core/internal/channels"
	"pryx-core/internal/config"
	"pryx-core/internal/cost"
	"pryx-core/internal/keychain"
	"pryx-core/internal/mcp"
	"pryx-core/internal/mcp/discovery"
	"pryx-core/internal/memory"
	"pryx-core/internal/models"
	"pryx-core/internal/policy"
	"pryx-core/internal/scheduler"
	"pryx-core/internal/skills"
	"pryx-core/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SpawnTool defines the interface for the agent spawning capability.
type SpawnTool interface {
	Name() string
	Description() string
	Schema() map[string]interface{}
	Execute(ctx context.Context, params json.RawMessage, parentID string) (interface{}, error)
	GetAgentStatus(agentID string) (map[string]interface{}, error)
	ListAgents() []map[string]interface{}
	ForkSession(sourceSessionID string) (string, error)
}

type pkceEntry struct {
	params    *auth.PKCEParams
	expiresAt time.Time
}

// Server is the main HTTP server for the Pryx runtime.
// It manages routing, middleware, and integration with various subsystems like MCP, skills, and agents.
type Server struct {
	cfg          *config.Config
	cfgMu        sync.RWMutex
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
	store        *store.Store
	auditRepo    *audit.AuditRepository
	costService  *cost.CostService
	channels     *channels.ChannelManager
	scheduler    *scheduler.Scheduler
	pkceParams   map[string]pkceEntry // Temporary storage for PKCE during OAuth flow
	mu           sync.Mutex           // Protects pkceParams

	httpMu     sync.Mutex
	httpServer *http.Server
}

// New creates a new Server instance with the provided configuration and dependencies.
func New(cfg *config.Config, db *sql.DB, kc *keychain.Keychain) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(MetricsMiddleware)
	r.Use(corsMiddleware(cfg))
	r.Use(DefaultRateLimiter().Middleware)

	p := policy.NewEngine(nil)

	s := &Server{
		cfg:      cfg,
		db:       db,
		keychain: kc,
		router:   r,
		bus:      bus.New(),
	}
	s.store = store.NewFromDB(db)
	s.auditRepo = audit.NewAuditRepository(db)

	pricingMgr := cost.NewPricingManager()
	costTracker := cost.NewCostTracker(s.auditRepo, pricingMgr)
	costCalc := cost.NewCostCalculator(pricingMgr)
	s.costService = cost.NewCostService(costTracker, costCalc, pricingMgr, s.store)

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

	dataDir := filepath.Dir(cfg.DatabasePath)
	mcp.InitTruncator(dataDir)

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

	s.channels = channels.NewManager(s.bus)
	s.scheduler = scheduler.New(db)

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
	s.router.Post("/skills/enable", s.handleSkillsEnable)
	s.router.Post("/skills/disable", s.handleSkillsDisable)
	s.router.Post("/skills/install", s.handleSkillsInstall)
	s.router.Post("/skills/uninstall", s.handleSkillsUninstall)
	s.router.Get("/api/v1/providers", s.handleProvidersList)
	s.router.Get("/api/v1/providers/{id}/models", s.handleProviderModels)
	s.router.Get("/api/v1/providers/{id}/key", s.handleProviderKeyStatus)
	s.router.Post("/api/v1/providers/{id}/key", s.handleProviderKeySet)
	s.router.Delete("/api/v1/providers/{id}/key", s.handleProviderKeyDelete)
	s.router.Get("/api/v1/cloud/status", s.handleCloudStatus)
	s.router.Post("/api/v1/cloud/login/start", s.handleCloudLoginStart)
	s.router.Post("/api/v1/cloud/login/poll", s.handleCloudLoginPoll)
	s.router.Get("/api/v1/config", s.handleConfigGet)
	s.router.Patch("/api/v1/config", s.handleConfigPatch)
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

	// Mesh pairing endpoints (pryx-jot)
	s.router.Post("/api/mesh/pair", s.handleMeshPair)
	s.router.Post("/api/mesh/qrcode", s.handleMeshQRCode)
	s.router.Get("/api/mesh/devices", s.handleMeshDevicesList)
	s.router.Post("/api/mesh/devices/{id}/unpair", s.handleMeshDevicesUnpair)
	s.router.Get("/api/mesh/events", s.handleMeshEventsList)

	// Channel management endpoints
	s.router.Get("/api/v1/channels", s.handleChannelsList)
	s.router.Get("/api/v1/channels/{id}", s.handleChannelGet)
	s.router.Post("/api/v1/channels", s.handleChannelCreate)
	s.router.Put("/api/v1/channels/{id}", s.handleChannelUpdate)
	s.router.Delete("/api/v1/channels/{id}", s.handleChannelDelete)
	s.router.Post("/api/v1/channels/{id}/test", s.handleChannelTest)
	s.router.Get("/api/v1/channels/{id}/health", s.handleChannelHealth)
	s.router.Post("/api/v1/channels/{id}/connect", s.handleChannelConnect)
	s.router.Post("/api/v1/channels/{id}/disconnect", s.handleChannelDisconnect)
	s.router.Get("/api/v1/channels/{id}/activity", s.handleChannelActivity)
	s.router.Get("/api/v1/channels/types", s.handleChannelTypes)

	s.router.Get("/api/v1/tasks", s.handleTasksList)
	s.router.Post("/api/v1/tasks", s.handleTaskCreate)
	s.router.Get("/api/v1/tasks/{id}", s.handleTaskGet)
	s.router.Put("/api/v1/tasks/{id}", s.handleTaskUpdate)
	s.router.Delete("/api/v1/tasks/{id}", s.handleTaskDelete)
	s.router.Post("/api/v1/tasks/{id}/enable", s.handleTaskEnable)
	s.router.Post("/api/v1/tasks/{id}/disable", s.handleTaskDisable)
	s.router.Get("/api/v1/tasks/{id}/runs", s.handleTaskRuns)
	s.router.Post("/api/v1/tasks/validate", s.handleTaskValidate)

	s.router.Get("/api/admin/stats", s.handleAdminStats)
	s.router.Get("/api/admin/users", s.handleAdminUsers)
	s.router.Get("/api/admin/devices", s.handleAdminDevices)
	s.router.Get("/api/admin/costs", s.handleAdminCosts)
	s.router.Get("/api/admin/health", s.handleAdminHealth)
	s.router.Get("/api/admin/telemetry/config", s.handleAdminTelemetryConfig)
	s.router.Put("/api/admin/telemetry/config", s.handleAdminTelemetryConfigUpdate)
}

// Bus returns the event bus instance.
func (s *Server) Bus() *bus.Bus {
	return s.bus
}

// Skills returns the skills registry instance.
func (s *Server) Skills() *skills.Registry {
	return s.skills
}

// MCP returns the MCP manager instance.
func (s *Server) MCP() *mcp.Manager {
	return s.mcp
}

// Agents returns the agent bus service instance.
func (s *Server) Agents() *agentbus.Service {
	return s.agentbus
}

// Memory returns the RAG memory manager instance.
func (s *Server) Memory() *memory.RAGManager {
	return s.ragMemory
}

// Channels returns the channel manager instance.
func (s *Server) Channels() *channels.ChannelManager {
	return s.channels
}

// Scheduler returns the scheduler instance.
func (s *Server) Scheduler() *scheduler.Scheduler {
	return s.scheduler
}

// AuditRepo returns the audit repository instance.
func (s *Server) AuditRepo() *audit.AuditRepository {
	return s.auditRepo
}

// CostService returns the cost service instance.
func (s *Server) CostService() *cost.CostService {
	return s.costService
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.router
}

// Start starts the HTTP server and blocks until it stops.
// It automatically allocates a port if configured to do so.
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

// Serve serves HTTP requests on the provided listener.
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

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.httpMu.Lock()
	srv := s.httpServer
	s.httpMu.Unlock()
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

// SetCatalog sets the model catalog for the server.
func (s *Server) SetCatalog(catalog *models.Catalog) {
	s.catalog = catalog
}

// SetSpawnTool sets the spawn tool for the server.
func (s *Server) SetSpawnTool(tool SpawnTool) {
	s.spawnTool = tool
}
