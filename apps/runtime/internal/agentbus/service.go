package agentbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"pryx-core/internal/bus"
)

// AgentIdentity uniquely identifies an agent across the network
type AgentIdentity struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Namespace string   `json:"namespace,omitempty"` // For multi-tenant isolation
	Tags      []string `json:"tags,omitempty"`      // For grouping/filtering
}

// AgentInfo represents discovered agent metadata
type AgentInfo struct {
	Identity     AgentIdentity          `json:"identity"`
	Endpoint     EndpointInfo           `json:"endpoint"`
	Capabilities []string               `json:"capabilities"`
	Protocol     string                 `json:"protocol"`
	LastSeen     time.Time              `json:"last_seen"`
	HealthStatus string                 `json:"health_status"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// EndpointInfo describes how to connect to an agent
type EndpointInfo struct {
	Type       string `json:"type"`          // "websocket", "http", "grpc", "stdio", "ipc", "file"
	URL        string `json:"url,omitempty"` // Full connection URL
	Host       string `json:"host,omitempty"`
	Port       int    `json:"port,omitempty"`
	Path       string `json:"path,omitempty"`
	LocalPath  string `json:"local_path,omitempty"` // For stdio/ipc/file-based agents
	WorkingDir string `json:"working_dir,omitempty"`
}

// AgentPackage represents an installable agent package
type AgentPackage struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Protocols    []string               `json:"protocols"` // Supported protocols
	Endpoints    []EndpointInfo         `json:"endpoints"`
	Capabilities []string               `json:"capabilities"`
	Install      InstallConfig          `json:"install"`
	Permissions  []string               `json:"permissions"`
	Dependencies map[string]string      `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// InstallConfig specifies how to install an agent
type InstallConfig struct {
	Type       string            `json:"type"`                  // "npm", "git", "url", "local"
	Source     string            `json:"source"`                // Package reference, URL, or path
	BinaryName string            `json:"binary_name,omitempty"` // Expected binary name
	Args       []string          `json:"args,omitempty"`        // Default arguments
	Env        map[string]string `json:"env,omitempty"`         // Environment variables
}

// UniversalMessage is the standard message format for all agent communication
type UniversalMessage struct {
	ID          string                 `json:"id"`
	TraceID     string                 `json:"trace_id"`    // Correlation ID for distributed tracing
	SpanID      string                 `json:"span_id"`     // Current span
	ParentSpan  string                 `json:"parent_span"` // Parent span for hierarchy
	From        AgentIdentity          `json:"from"`
	To          AgentIdentity          `json:"to"`
	ReplyTo     string                 `json:"reply_to,omitempty"` // For request-response patterns
	Protocol    string                 `json:"protocol"`           // Original protocol used
	MessageType string                 `json:"message_type"`       // "request", "response", "event", "stream"
	Action      string                 `json:"action"`             // Semantic action (e.g., "execute", "query", "notify")
	Payload     map[string]interface{} `json:"payload"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Context     map[string]string      `json:"context,omitempty"` // For passing request context
}

// MessageType constants
const (
	MessageTypeRequest  = "request"
	MessageTypeResponse = "response"
	MessageTypeEvent    = "event"
	MessageTypeStream   = "stream"
)

// ConnectionState represents the state of an agent connection
type ConnectionState string

const (
	ConnectionStateDisconnected ConnectionState = "disconnected"
	ConnectionStateConnecting   ConnectionState = "connecting"
	ConnectionStateConnected    ConnectionState = "connected"
	ConnectionStateReconnecting ConnectionState = "reconnecting"
	ConnectionStateFailed       ConnectionState = "failed"
	ConnectionStateClosed       ConnectionState = "closed"
)

// AgentConnection represents an active connection to an agent
type AgentConnection struct {
	ID           string                 `json:"id"`
	AgentInfo    AgentInfo              `json:"agent_info"`
	State        ConnectionState        `json:"state"`
	Protocol     string                 `json:"protocol"`
	Adapter      AgentAdapter           `json:"-"` // Plugin instance
	LastActivity time.Time              `json:"last_activity"`
	MessageCount int64                  `json:"message_count"`
	ErrorCount   int64                  `json:"error_count"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	ConnectedAt  *time.Time             `json:"connected_at,omitempty"`
}

// AgentAdapter is the interface for protocol adapters
type AgentAdapter interface {
	// Protocol returns the protocol name this adapter handles
	Protocol() string

	// Priority returns the adapter priority (higher = preferred)
	Priority() int

	// Detect discovers agents using this protocol
	Detect(ctx context.Context) ([]AgentInfo, error)

	// Connect establishes a connection to an agent
	Connect(ctx context.Context, agent AgentInfo, config AgentConfig) (AgentConnection, error)

	// Send transmits a message to the agent
	Send(ctx context.Context, conn *AgentConnection, msg *UniversalMessage) error

	// Receive receives a message from the agent
	Receive(ctx context.Context, conn *AgentConnection) (*UniversalMessage, error)

	// Disconnect closes the connection
	Disconnect(ctx context.Context, conn *AgentConnection) error

	// HealthCheck checks the agent health
	HealthCheck(ctx context.Context, conn *AgentConnection) error

	// Install installs an agent package
	Install(ctx context.Context, pkg AgentPackage) error

	// Uninstall removes an agent package
	Uninstall(ctx context.Context, pkg AgentPackage) error
}

// AgentConfig contains connection configuration
type AgentConfig struct {
	Timeout              time.Duration     `json:"timeout"`
	ReconnectEnabled     bool              `json:"reconnect_enabled"`
	ReconnectDelay       time.Duration     `json:"reconnect_delay"`
	MaxReconnectAttempts int               `json:"max_reconnect_attempts"`
	HeartbeatInterval    time.Duration     `json:"heartbeat_interval"`
	HealthCheckInterval  time.Duration     `json:"health_check_interval"`
	Permissions          []string          `json:"permissions"`
	Credentials          map[string]string `json:"credentials,omitempty"`
	Sandbox              *SandboxConfig    `json:"sandbox,omitempty"`
}

// SandboxConfig contains sandboxing configuration
type SandboxConfig struct {
	Type       string         `json:"type"` // "docker", "firejail", "namespace", "none"
	Limits     ResourceLimits `json:"limits"`
	Mounts     []MountConfig  `json:"mounts"`
	Networks   []string       `json:"networks"`
	WorkingDir string         `json:"working_dir"`
}

// ResourceLimits defines resource constraints
type ResourceLimits struct {
	MemoryMB    int64   `json:"memory_mb"`
	CPULimit    float64 `json:"cpu_limit"` // Percentage
	DiskQuotaMB int64   `json:"disk_quota_mb"`
	MaxProcs    int     `json:"max_procs"`
	TimeoutSec  int     `json:"timeout_sec"`
}

// MountConfig defines filesystem mounts
type MountConfig struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"read_only"`
	Type     string `json:"type"` // "bind", "tmpfs", "proc"
}

// ConnectionMetrics tracks connection statistics
type ConnectionMetrics struct {
	TotalConnections  int64            `json:"total_connections"`
	ActiveConnections int64            `json:"active_connections"`
	MessagesSent      int64            `json:"messages_sent"`
	MessagesReceived  int64            `json:"messages_received"`
	ErrorsTotal       int64            `json:"errors_total"`
	BytesSent         int64            `json:"bytes_sent"`
	BytesReceived     int64            `json:"bytes_received"`
	LastActivity      time.Time        `json:"last_activity"`
	ProtocolStats     map[string]int64 `json:"protocol_stats"`
}

// HubConfig contains hub configuration
type HubConfig struct {
	Name               string               `json:"name"`
	Namespace          string               `json:"namespace"`
	LogLevel           string               `json:"log_level"`
	MetricsEnabled     bool                 `json:"metrics_enabled"`
	AutoDetectEnabled  bool                 `json:"auto_detect_enabled"`
	AutoDetectInterval time.Duration        `json:"auto_detect_interval"`
	PackageDir         string               `json:"package_dir"`
	CacheDir           string               `json:"cache_dir"`
	MaxConnections     int                  `json:"max_connections"`
	ReconnectEnabled   bool                 `json:"reconnect_enabled"`
	CircuitBreaker     CircuitBreakerConfig `json:"circuit_breaker"`
}

// CircuitBreakerConfig contains circuit breaker settings
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`
	HalfOpenRequests int           `json:"half_open_requests"`
}

// Service is the main hub orchestrating all agent connectivity
type Service struct {
	mu     sync.RWMutex
	bus    *bus.Bus
	config HubConfig
	logger *StructuredLogger

	// Managers
	registry    *RegistryManager
	connections *ConnectionManager
	packages    *PackageManager
	detector    *DetectionManager
	router      *MessageRouter

	// Protocol adapters
	adapters     map[string]AgentAdapter
	adapterOrder []string // Ordered by priority

	// State
	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// NewService creates a new agent connectivity hub
func NewService(b *bus.Bus, config HubConfig) *Service {
	return &Service{
		bus:    b,
		config: config,
		logger: NewStructuredLogger(config.Name, config.LogLevel),

		registry:    NewRegistryManager(b),
		connections: NewConnectionManager(b),
		packages:    NewPackageManager(b, config.PackageDir),
		detector:    NewDetectionManager(b),
		router:      NewMessageRouter(b),

		adapters:     make(map[string]AgentAdapter),
		adapterOrder: []string{},
		stopCh:       make(chan struct{}),
	}
}

// RegisterAdapter registers a protocol adapter
func (s *Service) RegisterAdapter(adapter AgentAdapter) {
	s.mu.Lock()
	defer s.mu.Unlock()

	protocol := adapter.Protocol()
	s.adapters[protocol] = adapter

	// Maintain priority order
	inserted := false
	for i, p := range s.adapterOrder {
		if adapter.Priority() > s.adapters[p].Priority() {
			s.adapterOrder = append(s.adapterOrder[:i], append([]string{protocol}, s.adapterOrder[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		s.adapterOrder = append(s.adapterOrder, protocol)
	}

	s.logger.Info("registered protocol adapter", map[string]interface{}{
		"protocol": protocol,
		"priority": adapter.Priority(),
	})
}

// Start initializes the hub and starts background services
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("hub already running")
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("starting agent connectivity hub", map[string]interface{}{
		"name":      s.config.Name,
		"namespace": s.config.Namespace,
	})

	// Initialize managers
	if err := s.registry.Start(ctx); err != nil {
		return fmt.Errorf("failed to start registry: %w", err)
	}
	if err := s.connections.Start(ctx); err != nil {
		return fmt.Errorf("failed to start connection manager: %w", err)
	}
	if err := s.packages.Start(ctx); err != nil {
		return fmt.Errorf("failed to start package manager: %w", err)
	}

	// Start auto-detection if enabled
	if s.config.AutoDetectEnabled {
		s.wg.Add(1)
		go s.autoDetectLoop(ctx)
	}

	// Start health monitoring
	s.wg.Add(1)
	go s.healthCheckLoop(ctx)

	// Register default adapters
	s.registerDefaultAdapters()

	s.logger.Info("agent connectivity hub started", nil)

	// Publish event
	s.bus.Publish(bus.NewEvent("agentbus.started", "", map[string]interface{}{
		"name":      s.config.Name,
		"namespace": s.config.Namespace,
	}))

	return nil
}

// Stop gracefully shuts down the hub
func (s *Service) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	s.logger.Info("stopping agent connectivity hub", nil)

	// Signal background services to stop
	close(s.stopCh)
	s.wg.Wait()

	// Stop managers
	if err := s.connections.Stop(ctx); err != nil {
		s.logger.Error("error stopping connection manager", map[string]interface{}{"error": err.Error()})
	}
	if err := s.registry.Stop(ctx); err != nil {
		s.logger.Error("error stopping registry", map[string]interface{}{"error": err.Error()})
	}
	if err := s.packages.Stop(ctx); err != nil {
		s.logger.Error("error stopping package manager", map[string]interface{}{"error": err.Error()})
	}

	s.logger.Info("agent connectivity hub stopped", nil)

	s.bus.Publish(bus.NewEvent("agentbus.stopped", "", map[string]interface{}{
		"name": s.config.Name,
	}))

	return nil
}

// registerDefaultAdapters registers built-in protocol adapters
func (s *Service) registerDefaultAdapters() {
	// Adapters will be registered by the runtime based on available packages
	s.logger.Info("registered default adapters", map[string]interface{}{
		"count": len(s.adapters),
	})
}

// autoDetectLoop periodically discovers new agents
func (s *Service) autoDetectLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.AutoDetectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			agents, err := s.detector.DetectAll(ctx, s.adapters)
			if err != nil {
				s.logger.Error("auto-detection failed", map[string]interface{}{"error": err.Error()})
				continue
			}

			for _, agent := range agents {
				if _, err := s.registry.Register(ctx, &agent); err != nil {
					s.logger.Debug("failed to register auto-detected agent", map[string]interface{}{
						"agent": agent.Identity.Name,
						"error": err.Error(),
					})
				}
			}
		}
	}
}

// healthCheckLoop monitors agent connections
func (s *Service) healthCheckLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			// Health checks are handled by ConnectionManager
		}
	}
}

// Connect establishes a connection to an agent
func (s *Service) Connect(ctx context.Context, agentID string, config AgentConfig) (*AgentConnection, error) {
	s.logger.Info("connecting to agent", map[string]interface{}{"agent_id": agentID})

	agent, err := s.registry.Get(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Find appropriate adapter
	adapter, err := s.findAdapter(agent.Protocol)
	if err != nil {
		return nil, err
	}

	// Apply default config if not provided
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Establish connection
	conn, err := adapter.Connect(ctx, *agent, config)
	if err != nil {
		s.logger.Error("failed to connect to agent", map[string]interface{}{
			"agent_id": agentID,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	// Register connection
	s.connections.Add(ctx, &conn)

	s.logger.Info("connected to agent", map[string]interface{}{
		"agent_id": agentID,
		"protocol": agent.Protocol,
	})

	s.bus.Publish(bus.NewEvent("agentbus.connected", "", map[string]interface{}{
		"agent_id": agentID,
		"protocol": agent.Protocol,
		"endpoint": agent.Endpoint.URL,
	}))

	return &conn, nil
}

// Disconnect closes a connection to an agent
func (s *Service) Disconnect(ctx context.Context, connID string) error {
	conn, err := s.connections.Get(ctx, connID)
	if err != nil {
		return err
	}

	if err := conn.Adapter.Disconnect(ctx, conn); err != nil {
		s.logger.Error("failed to disconnect", map[string]interface{}{
			"connection_id": connID,
			"error":         err.Error(),
		})
	}

	s.connections.Remove(ctx, connID)

	s.logger.Info("disconnected from agent", map[string]interface{}{
		"agent_id": conn.AgentInfo.Identity.ID,
	})

	s.bus.Publish(bus.NewEvent("agentbus.disconnected", "", map[string]interface{}{
		"agent_id": conn.AgentInfo.Identity.ID,
	}))

	return nil
}

// SendMessage transmits a message to an agent
func (s *Service) SendMessage(ctx context.Context, msg *UniversalMessage) error {
	// Generate trace ID if not provided
	if msg.TraceID == "" {
		msg.TraceID = uuid.New().String()
	}
	msg.Timestamp = time.Now().UTC()

	// Route message
	routed, err := s.router.Route(ctx, msg)
	if err != nil {
		s.logger.Error("failed to route message", map[string]interface{}{
			"trace_id": msg.TraceID,
			"error":    err.Error(),
		})
		return fmt.Errorf("routing failed: %w", err)
	}

	if !routed {
		return fmt.Errorf("no route to agent: %s", msg.To.ID)
	}

	s.logger.Debug("message routed", map[string]interface{}{
		"trace_id": msg.TraceID,
		"from":     msg.From.ID,
		"to":       msg.To.ID,
		"action":   msg.Action,
	})

	return nil
}

// ReceiveMessage waits for a message from an agent
func (s *Service) ReceiveMessage(ctx context.Context, connID string) (*UniversalMessage, error) {
	conn, err := s.connections.Get(ctx, connID)
	if err != nil {
		return nil, err
	}

	msg, err := conn.Adapter.Receive(ctx, conn)
	if err != nil {
		s.logger.Error("failed to receive message", map[string]interface{}{
			"connection_id": connID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("receive failed: %w", err)
	}

	s.logger.Debug("message received", map[string]interface{}{
		"trace_id": msg.TraceID,
		"from":     msg.From.ID,
		"to":       msg.To.ID,
		"action":   msg.Action,
	})

	return msg, nil
}

// InstallPackage installs an agent package
func (s *Service) InstallPackage(ctx context.Context, pkg AgentPackage) error {
	s.logger.Info("installing agent package", map[string]interface{}{
		"name":    pkg.Name,
		"version": pkg.Version,
	})

	// Find appropriate adapter for installation
	var installer AgentAdapter
	for _, adapter := range s.adapters {
		if installer == nil || adapter.Priority() > installer.Priority() {
			installer = adapter
		}
	}

	if installer == nil {
		return fmt.Errorf("no installer available")
	}

	if err := installer.Install(ctx, pkg); err != nil {
		s.logger.Error("failed to install package", map[string]interface{}{
			"name":  pkg.Name,
			"error": err.Error(),
		})
		return fmt.Errorf("installation failed: %w", err)
	}

	// Register discovered agent
	agents, err := s.detector.DetectProtocol(ctx, installer.Protocol(), s.adapters)
	if err != nil {
		s.logger.Warn("failed to detect agents after install", map[string]interface{}{
			"package": pkg.Name,
			"error":   err.Error(),
		})
	}

	for _, agent := range agents {
		if agent.capabilitiesMatches(pkg.Capabilities) {
			s.registry.Register(ctx, &agent)
		}
	}

	s.logger.Info("installed agent package", map[string]interface{}{
		"name": pkg.Name,
	})

	s.bus.Publish(bus.NewEvent("agentbus.package.installed", "", map[string]interface{}{
		"name":    pkg.Name,
		"version": pkg.Version,
	}))

	return nil
}

// UninstallPackage removes an agent package
func (s *Service) UninstallPackage(ctx context.Context, pkg AgentPackage) error {
	s.logger.Info("uninstalling agent package", map[string]interface{}{
		"name": pkg.Name,
	})

	var uninstaller AgentAdapter
	for _, adapter := range s.adapters {
		if uninstaller == nil || adapter.Priority() > uninstaller.Priority() {
			uninstaller = adapter
		}
	}

	if uninstaller == nil {
		return fmt.Errorf("no uninstaller available")
	}

	if err := uninstaller.Uninstall(ctx, pkg); err != nil {
		s.logger.Error("failed to uninstall package", map[string]interface{}{
			"name":  pkg.Name,
			"error": err.Error(),
		})
		return fmt.Errorf("uninstallation failed: %w", err)
	}

	s.logger.Info("uninstalled agent package", map[string]interface{}{
		"name": pkg.Name,
	})

	s.bus.Publish(bus.NewEvent("agentbus.package.uninstalled", "", map[string]interface{}{
		"name": pkg.Name,
	}))

	return nil
}

// GetMetrics returns connection and message metrics
func (s *Service) GetMetrics() ConnectionMetrics {
	return s.connections.GetMetrics()
}

// GetRegistry returns the registry manager
func (s *Service) GetRegistry() *RegistryManager {
	return s.registry
}

// GetConnections returns the connection manager
func (s *Service) GetConnections() *ConnectionManager {
	return s.connections
}

// GetPackages returns the package manager
func (s *Service) GetPackages() *PackageManager {
	return s.packages
}

// findAdapter finds the appropriate adapter for a protocol
func (s *Service) findAdapter(protocol string) (AgentAdapter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	adapter, exists := s.adapters[protocol]
	if !exists {
		return nil, fmt.Errorf("no adapter for protocol: %s", protocol)
	}

	return adapter, nil
}

// capabilitiesMatches checks if agent capabilities match required capabilities
func (a *AgentInfo) capabilitiesMatches(required []string) bool {
	for _, req := range required {
		found := false
		for _, cap := range a.Capabilities {
			if cap == req {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// MarshalJSON for UniversalMessage
func (m *UniversalMessage) MarshalJSON() ([]byte, error) {
	type Alias UniversalMessage
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	})
}

// MarshalJSON for AgentConnection
func (c *AgentConnection) MarshalJSON() ([]byte, error) {
	type Alias AgentConnection
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	})
}

// PrettyPrint returns a human-readable representation
func (i *AgentIdentity) PrettyPrint() string {
	return fmt.Sprintf("AgentIdentity{Name: %s (%s), Version: %s, Namespace: %s}",
		i.Name, i.ID, i.Version, i.Namespace)
}

func (m *UniversalMessage) PrettyPrint() string {
	return fmt.Sprintf("UniversalMessage{ID: %s, From: %s, To: %s, Action: %s, TraceID: %s}",
		m.ID, m.From.Name, m.To.Name, m.Action, m.TraceID)
}

func (c *AgentConnection) PrettyPrint() string {
	return fmt.Sprintf("AgentConnection{ID: %s, Agent: %s, State: %s, Protocol: %s}",
		c.ID, c.AgentInfo.Identity.Name, c.State, c.Protocol)
}
