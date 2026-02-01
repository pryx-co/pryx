package agentbus

import (
	"context"
	"sync"
	"time"

	"pryx-core/internal/bus"
)

// ConnectionManager manages active connections with circuit breakers
type ConnectionManager struct {
	mu              sync.RWMutex
	bus             *bus.Bus
	logger          *StructuredLogger
	connections     map[string]*AgentConnection
	metrics         ConnectionMetrics
	circuitBreakers map[string]*CircuitBreaker
	running         bool
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(b *bus.Bus) *ConnectionManager {
	return &ConnectionManager{
		bus:             b,
		logger:          NewStructuredLogger("connections", "info"),
		connections:     make(map[string]*AgentConnection),
		circuitBreakers: make(map[string]*CircuitBreaker),
		metrics: ConnectionMetrics{
			ProtocolStats: make(map[string]int64),
		},
		stopCh: make(chan struct{}),
	}
}

// Start initializes the connection manager
func (cm *ConnectionManager) Start(ctx context.Context) error {
	cm.mu.Lock()
	if cm.running {
		cm.mu.Unlock()
		return nil
	}
	cm.running = true
	cm.mu.Unlock()

	cm.logger.Info("connection manager started", nil)
	cm.bus.Publish(bus.NewEvent("agentbus.connections.started", "", nil))

	return nil
}

// Stop gracefully shuts down the connection manager
func (cm *ConnectionManager) Stop(ctx context.Context) error {
	cm.mu.Lock()
	if !cm.running {
		cm.mu.Unlock()
		return nil
	}
	cm.running = false
	cm.mu.Unlock()

	close(cm.stopCh)
	cm.wg.Wait()

	// Close all connections
	cm.mu.Lock()
	for id, conn := range cm.connections {
		if conn.Adapter != nil {
			conn.Adapter.Disconnect(ctx, conn)
		}
		delete(cm.connections, id)
	}
	cm.mu.Unlock()

	cm.logger.Info("connection manager stopped", nil)
	cm.bus.Publish(bus.NewEvent("agentbus.connections.stopped", "", nil))

	return nil
}

// Add registers a new connection
func (cm *ConnectionManager) Add(ctx context.Context, conn *AgentConnection) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if already exists
	if _, exists := cm.connections[conn.ID]; exists {
		cm.logger.Warn("connection already exists", map[string]interface{}{
			"connection_id": conn.ID,
		})
		return
	}

	// Set creation time
	now := time.Now().UTC()
	conn.CreatedAt = now
	conn.ConnectedAt = &now

	// Store connection
	cm.connections[conn.ID] = conn

	// Update metrics
	cm.metrics.TotalConnections++
	cm.metrics.ActiveConnections++
	cm.metrics.ProtocolStats[conn.Protocol]++

	cm.logger.Info("connection added", map[string]interface{}{
		"connection_id": conn.ID,
		"agent_name":    conn.AgentInfo.Identity.Name,
		"protocol":      conn.Protocol,
	})

	cm.bus.Publish(bus.NewEvent("agentbus.connection.added", "", map[string]interface{}{
		"connection_id": conn.ID,
		"agent_id":      conn.AgentInfo.Identity.ID,
		"agent_name":    conn.AgentInfo.Identity.Name,
		"protocol":      conn.Protocol,
	}))
}

// Remove closes and removes a connection
func (cm *ConnectionManager) Remove(ctx context.Context, connID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return
	}

	// Close connection if adapter exists
	if conn.Adapter != nil {
		conn.Adapter.Disconnect(ctx, conn)
	}

	// Remove from registry
	delete(cm.connections, connID)

	// Update metrics
	cm.metrics.ActiveConnections--

	cm.logger.Info("connection removed", map[string]interface{}{
		"connection_id": connID,
		"agent_name":    conn.AgentInfo.Identity.Name,
	})

	cm.bus.Publish(bus.NewEvent("agentbus.connection.removed", "", map[string]interface{}{
		"connection_id": connID,
		"agent_id":      conn.AgentInfo.Identity.ID,
	}))
}

// Get retrieves a connection by ID
func (cm *ConnectionManager) Get(ctx context.Context, connID string) (*AgentConnection, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return nil, nil
	}

	return conn, nil
}

// GetByAgentID retrieves connections for a specific agent
func (cm *ConnectionManager) GetByAgentID(ctx context.Context, agentID string) []*AgentConnection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var connections []*AgentConnection
	for _, conn := range cm.connections {
		if conn.AgentInfo.Identity.ID == agentID {
			connections = append(connections, conn)
		}
	}

	return connections
}

// List returns all active connections
func (cm *ConnectionManager) List(ctx context.Context) []*AgentConnection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	connections := make([]*AgentConnection, 0, len(cm.connections))
	for _, conn := range cm.connections {
		connections = append(connections, conn)
	}

	return connections
}

// UpdateMetrics updates message and error counts
func (cm *ConnectionManager) UpdateMetrics(connID string, messagesSent, messagesReceived int64, bytesSent, bytesReceived int64, errorCount int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return
	}

	conn.MessageCount += messagesSent + messagesReceived
	conn.ErrorCount += errorCount
	conn.LastActivity = time.Now().UTC()

	// Update global metrics
	cm.metrics.MessagesSent += messagesSent
	cm.metrics.MessagesReceived += messagesReceived
	cm.metrics.BytesSent += bytesSent
	cm.metrics.BytesReceived += bytesReceived
	cm.metrics.ErrorsTotal += errorCount
	cm.metrics.LastActivity = time.Now().UTC()
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

// HealthCheck performs health check on all connections
func (cm *ConnectionManager) HealthCheck(ctx context.Context) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var errs []error
	for id, conn := range cm.connections {
		if err := conn.Adapter.HealthCheck(ctx, conn); err != nil {
			conn.ErrorCount++
			cm.metrics.ErrorsTotal++
			errs = append(errs, err)

			cm.logger.Warn("health check failed", map[string]interface{}{
				"connection_id": id,
				"agent_name":    conn.AgentInfo.Identity.Name,
				"error":         err.Error(),
			})
		} else {
			conn.AgentInfo.HealthStatus = "healthy"
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// Count returns the number of active connections
func (cm *ConnectionManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.connections)
}
