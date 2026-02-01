package universal

import (
	"context"
	"sync"
	"time"
)

// AgentAdapter is the interface that all protocol adapters must implement
type AgentAdapter interface {
	// Protocol returns the protocol name (websocket, http, stdio, etc.)
	Protocol() string

	// Name returns the adapter name
	Name() string

	// Version returns the adapter version
	Version() string

	// Detect discovers agents using this protocol
	Detect(ctx context.Context) ([]DetectedAgent, error)

	// Connect establishes a connection to an agent
	Connect(ctx context.Context, info AgentInfo, config AgentConfig) (*AgentConnection, error)

	// Send transmits a message to the agent
	Send(ctx context.Context, conn *AgentConnection, msg *UniversalMessage) error

	// Receive receives a message from the agent
	Receive(ctx context.Context, conn *AgentConnection) (*UniversalMessage, error)

	// Disconnect closes the connection
	Disconnect(ctx context.Context, conn *AgentConnection) error

	// HealthCheck checks the agent health
	HealthCheck(ctx context.Context, conn *AgentConnection) error

	// Install installs an agent package
	Install(ctx context.Context, ref string, config AgentConfig) error

	// Uninstall removes an agent package
	Uninstall(ctx context.Context, ref string) error
}

// BaseAdapter provides common functionality for all adapters
type BaseAdapter struct {
	name    string
	version string
}

func (b *BaseAdapter) Name() string    { return b.name }
func (b *BaseAdapter) Version() string { return b.version }

// ConnectionManager manages agent connections
type ConnectionManager struct {
	connections     map[string]*AgentConnection
	circuitBreakers map[string]*CircuitBreaker
	metrics         ConnectionMetrics
	mu              sync.RWMutex
	running         bool
	stopCh          chan struct{}
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

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "closed"
	CircuitBreakerOpen     CircuitBreakerState = "open"
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

// CircuitBreakerConfig contains circuit breaker settings
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`
	HalfOpenRequests int           `json:"half_open_requests"`
}

// CircuitBreaker prevents cascading failures
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	config          CircuitBreakerConfig
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections:     make(map[string]*AgentConnection),
		circuitBreakers: make(map[string]*CircuitBreaker),
		metrics: ConnectionMetrics{
			ProtocolStats: make(map[string]int64),
		},
		stopCh: make(chan struct{}),
	}
}

// Start initializes the connection manager
func (cm *ConnectionManager) Start(ctx context.Context) {
	cm.running = true
}

// Stop gracefully shuts down the connection manager
func (cm *ConnectionManager) Stop(ctx context.Context) {
	cm.running = false
	close(cm.stopCh)
}

// Add registers a new connection
func (cm *ConnectionManager) Add(conn *AgentConnection) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.connections[conn.ID] = conn
	cm.metrics.TotalConnections++
	cm.metrics.ActiveConnections++
	cm.metrics.ProtocolStats[conn.Protocol]++
}

// Remove closes and removes a connection
func (cm *ConnectionManager) Remove(connID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.connections[connID]; !exists {
		return
	}

	delete(cm.connections, connID)
	cm.metrics.ActiveConnections--
}

// Get retrieves a connection by ID
func (cm *ConnectionManager) Get(connID string) (*AgentConnection, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[connID]
	return conn, exists
}

// List returns all connections
func (cm *ConnectionManager) List() []*AgentConnection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conns := make([]*AgentConnection, 0, len(cm.connections))
	for _, conn := range cm.connections {
		conns = append(conns, conn)
	}
	return conns
}

// GetMetrics returns connection statistics
func (cm *ConnectionManager) GetMetrics() ConnectionMetrics {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.metrics
}

// GetCircuitBreaker returns or creates a circuit breaker for a connection
func (cm *ConnectionManager) GetCircuitBreaker(connID string, config CircuitBreakerConfig) *CircuitBreaker {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cb, exists := cm.circuitBreakers[connID]; exists {
		return cb
	}

	cb := NewCircuitBreaker(config)
	cm.circuitBreakers[connID] = cb
	return cb
}

// UpdateMetrics updates message and error counts for a connection
func (cm *ConnectionManager) UpdateMetrics(connID string, messagesSent, messagesReceived int64, bytesSent, bytesReceived int64, errorCount int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return
	}

	conn.MessageCount += messagesSent + messagesReceived
	conn.ErrorCount += errorCount
	conn.LastActivity = time.Now()

	// Update global metrics
	cm.metrics.MessagesSent += messagesSent
	cm.metrics.MessagesReceived += messagesReceived
	cm.metrics.BytesSent += bytesSent
	cm.metrics.BytesReceived += bytesReceived
	cm.metrics.ErrorsTotal += errorCount
	cm.metrics.LastActivity = time.Now()
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.RecoveryTimeout == 0 {
		config.RecoveryTimeout = 30 * time.Second
	}
	if config.HalfOpenRequests == 0 {
		config.HalfOpenRequests = 3
	}

	return &CircuitBreaker{
		state:  CircuitBreakerClosed,
		config: config,
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// AllowRequest checks if a request should be allowed
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		if time.Since(cb.lastFailureTime) > cb.config.RecoveryTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.successCount = 0
			cb.failureCount = 0
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return cb.successCount < cb.config.HalfOpenRequests
	}
	return false
}

// RecordSuccess marks a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++

	if cb.state == CircuitBreakerHalfOpen && cb.successCount >= cb.config.HalfOpenRequests {
		cb.state = CircuitBreakerClosed
		cb.failureCount = 0
	}
}

// RecordFailure marks a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == CircuitBreakerClosed && cb.failureCount >= cb.config.FailureThreshold {
		cb.state = CircuitBreakerOpen
	} else if cb.state == CircuitBreakerHalfOpen {
		cb.state = CircuitBreakerOpen
	}
}
