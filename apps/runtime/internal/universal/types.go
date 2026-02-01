package universal

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"pryx-core/internal/bus"
)

// UniversalMessage is the standard message format for all agent communication
type UniversalMessage struct {
	ID          string                 `json:"id"`
	TraceID     string                 `json:"trace_id"`
	SpanID      string                 `json:"span_id"`
	ParentSpan  string                 `json:"parent_span"`
	From        AgentIdentity          `json:"from"`
	To          AgentIdentity          `json:"to"`
	ReplyTo     string                 `json:"reply_to,omitempty"`
	Protocol    string                 `json:"protocol"`
	MessageType string                 `json:"message_type"` // request, response, event, stream
	Action      string                 `json:"action"`
	Payload     map[string]interface{} `json:"payload"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Context     map[string]string      `json:"context,omitempty"`
}

// MessageType constants
const (
	MessageTypeRequest  = "request"
	MessageTypeResponse = "response"
	MessageTypeEvent    = "event"
	MessageTypeStream   = "stream"
)

// AgentIdentity uniquely identifies an agent
type AgentIdentity struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Namespace string   `json:"namespace,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

// AgentInfo represents discovered agent metadata
type AgentInfo struct {
	Identity     AgentIdentity  `json:"identity"`
	Protocol     string         `json:"protocol"`
	Endpoint     EndpointInfo   `json:"endpoint"`
	Capabilities []string       `json:"capabilities"`
	LastSeen     time.Time      `json:"last_seen"`
	HealthStatus string         `json:"health_status"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// EndpointInfo describes how to connect to an agent
type EndpointInfo struct {
	Type       string `json:"type"` // websocket, http, grpc, stdio, ipc
	URL        string `json:"url,omitempty"`
	Host       string `json:"host,omitempty"`
	Port       int    `json:"port,omitempty"`
	Path       string `json:"path,omitempty"`
	LocalPath  string `json:"local_path,omitempty"`
	WorkingDir string `json:"working_dir,omitempty"`
}

// AgentPackage represents an installable agent package
type AgentPackage struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Protocols    []string               `json:"protocols"`
	Endpoints    []EndpointInfo         `json:"endpoints"`
	Capabilities []string               `json:"capabilities"`
	Install      InstallConfig          `json:"install"`
	Permissions  []string               `json:"permissions"`
	Dependencies map[string]string      `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// InstallConfig specifies how to install an agent
type InstallConfig struct {
	Type       string            `json:"type"` // npm, git, url, local
	Source     string            `json:"source"`
	BinaryName string            `json:"binary_name,omitempty"`
	Args       []string          `json:"args,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
}

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
	Adapter      AgentAdapter           `json:"-"`
	LastActivity time.Time              `json:"last_activity"`
	MessageCount int64                  `json:"message_count"`
	ErrorCount   int64                  `json:"error_count"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	ConnectedAt  *time.Time             `json:"connected_at,omitempty"`
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
	Type       string         `json:"type"`
	Limits     ResourceLimits `json:"limits"`
	Mounts     []MountConfig  `json:"mounts"`
	Networks   []string       `json:"networks"`
	WorkingDir string         `json:"working_dir"`
}

// ResourceLimits defines resource constraints
type ResourceLimits struct {
	MemoryMB    int64   `json:"memory_mb"`
	CPULimit    float64 `json:"cpu_limit"`
	DiskQuotaMB int64   `json:"disk_quota_mb"`
	MaxProcs    int     `json:"max_procs"`
	TimeoutSec  int     `json:"timeout_sec"`
}

// MountConfig defines filesystem mounts
type MountConfig struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"read_only"`
	Type     string `json:"type"`
}

// DetectedAgent represents an agent detected during discovery
type DetectedAgent struct {
	AgentInfo       `json:"agent_info"`
	DetectionMethod string            `json:"detection_method"` // port, mdns, filesystem, handshake
	Confidence      float64           `json:"confidence"`
	HandshakeData   map[string]string `json:"handshake_data,omitempty"`
}

// UniversalHub is the main orchestrator
type UniversalHub struct {
	mu       sync.RWMutex
	bus      *bus.Bus
	config   HubConfig
	adapters map[string]AgentAdapter

	connections *ConnectionManager
	registry    *Registry
	detector    *Detector
	router      *MessageRouter
	translator  *Translator

	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// HubConfig contains hub configuration
type HubConfig struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	LogLevel          string `json:"log_level"`
	AutoDetectEnabled bool   `json:"auto_detect_enabled"`
	MaxConnections    int    `json:"max_connections"`
	ReconnectEnabled  bool   `json:"reconnect_enabled"`
	ScanPorts         []int  `json:"scan_ports"`
	ScanInterval      string `json:"scan_interval"`
}

// NewUniversalHub creates a new universal agent hub
func NewUniversalHub(b *bus.Bus, config HubConfig) *UniversalHub {
	return &UniversalHub{
		bus:      b,
		config:   config,
		adapters: make(map[string]AgentAdapter),

		connections: NewConnectionManager(),
		registry:    NewRegistry(),
		detector:    NewDetector(config.ScanPorts),
		router:      NewMessageRouter(),
		translator:  NewTranslator(),

		stopCh: make(chan struct{}),
	}
}

// RegisterAdapter registers a protocol adapter
func (h *UniversalHub) RegisterAdapter(adapter AgentAdapter) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.adapters[adapter.Protocol()] = adapter
}

// Start initializes the hub
func (h *UniversalHub) Start(ctx context.Context) error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return fmt.Errorf("hub already running")
	}
	h.running = true
	h.mu.Unlock()

	// Start managers
	h.connections.Start(ctx)
	h.registry.Start(ctx)
	h.detector.Start(ctx)
	h.router.Start(ctx)

	return nil
}

// Stop gracefully shuts down the hub
func (h *UniversalHub) Stop(ctx context.Context) error {
	h.mu.Lock()
	if !h.running {
		h.mu.Unlock()
		return nil
	}
	h.running = false
	h.mu.Unlock()

	close(h.stopCh)
	h.wg.Wait()

	h.connections.Stop(ctx)
	h.registry.Stop(ctx)
	h.detector.Stop(ctx)
	h.router.Stop(ctx)

	return nil
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

// PrettyPrint returns a human-readable representation
func (m *UniversalMessage) PrettyPrint() string {
	return fmt.Sprintf("UniversalMessage{ID: %s, From: %s, To: %s, Action: %s}",
		m.ID, m.From.Name, m.To.Name, m.Action)
}

// CorrelationID generates a new correlation ID
func CorrelationID() string {
	return uuid.New().String()
}
