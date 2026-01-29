package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"pryx-core/internal/bus"
	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/store"
)

// Message types for WebSocket protocol
const (
	MsgTypeEvent        = "event"
	MsgTypeSyncRequest  = "sync_request"
	MsgTypeSyncResponse = "sync_response"
	MsgTypePing         = "ping"
	MsgTypePong         = "pong"
	MsgTypePresence     = "presence"
)

// WebSocketMessage is the envelope for all mesh messages
type WebSocketMessage struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	DeviceID  string          `json:"device_id"`
	SessionID string          `json:"session_id,omitempty"`
}

// Manager coordinates multi-device sync via WebSocket
type Manager struct {
	cfg      *config.Config
	bus      *bus.Bus
	store    *store.Store
	keychain *keychain.Keychain

	// WebSocket connection
	wsConn *websocket.Conn
	connMu sync.RWMutex

	// State
	deviceID  string
	connected bool

	// Channels
	sendCh chan WebSocketMessage
	stopCh chan struct{}
}

// NewManager creates a new mesh manager
func NewManager(cfg *config.Config, b *bus.Bus, s *store.Store, kc *keychain.Keychain) *Manager {
	return &Manager{
		cfg:      cfg,
		bus:      b,
		store:    s,
		keychain: kc,
		sendCh:   make(chan WebSocketMessage, 100),
		stopCh:   make(chan struct{}),
	}
}

// Start begins mesh coordination
func (m *Manager) Start(ctx context.Context) {
	// Get device ID
	deviceID, err := m.keychain.Get("device_id")
	if err != nil {
		deviceID = generateDeviceID()
		m.keychain.Set("device_id", deviceID)
	}
	m.deviceID = deviceID

	// Start WebSocket connection manager
	go m.connectionManager(ctx)

	// Listen for local events to broadcast
	go m.listenForBroadcasts(ctx)

	// Handle incoming messages
	go m.handleIncoming(ctx)

	// Send periodic presence updates
	go m.presenceLoop(ctx)

	log.Printf("Pryx Mesh Manager started (device: %s)", m.deviceID)
}

// Stop gracefully shuts down the mesh manager
func (m *Manager) Stop() {
	close(m.stopCh)
	m.disconnect()
}

// connectionManager maintains the WebSocket connection with exponential backoff
func (m *Manager) connectionManager(ctx context.Context) {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		default:
		}

		if !m.isConnected() {
			if err := m.connect(ctx); err != nil {
				log.Printf("Mesh: Connection failed: %v, retrying in %v", err, backoff)
				time.Sleep(backoff)
				backoff = min(backoff*2, maxBackoff)
				continue
			}
			backoff = time.Second // Reset backoff on success
		}

		time.Sleep(5 * time.Second)
	}
}

// connect establishes WebSocket connection to coordinator
func (m *Manager) connect(ctx context.Context) error {
	token, err := m.keychain.Get("cloud_access_token")
	if err != nil {
		return fmt.Errorf("not authenticated")
	}

	// Build WebSocket URL
	wsURL := m.cfg.CloudAPIUrl + "/mesh/ws"

	opts := &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + token},
			"X-Device-ID":   []string{m.deviceID},
		},
	}

	conn, _, err := websocket.Dial(ctx, wsURL, opts)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	m.connMu.Lock()
	m.wsConn = conn
	m.connected = true
	m.connMu.Unlock()

	log.Println("Mesh: WebSocket connected")

	// Send initial presence
	m.sendPresence()

	return nil
}

// disconnect closes the WebSocket connection
func (m *Manager) disconnect() {
	m.connMu.Lock()
	defer m.connMu.Unlock()

	if m.wsConn != nil {
		m.wsConn.Close(websocket.StatusNormalClosure, "disconnecting")
		m.wsConn = nil
		m.connected = false
		log.Println("Mesh: WebSocket disconnected")
	}
}

// isConnected returns connection status
func (m *Manager) isConnected() bool {
	m.connMu.RLock()
	defer m.connMu.RUnlock()
	return m.connected && m.wsConn != nil
}

// listenForBroadcasts watches for local events to send to mesh
func (m *Manager) listenForBroadcasts(ctx context.Context) {
	events, closer := m.bus.Subscribe(
		bus.EventSessionMessage,
		bus.EventSessionTyping,
		bus.EventToolRequest,
		bus.EventToolComplete,
		bus.EventApprovalNeeded,
	)
	defer closer()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case evt, ok := <-events:
			if !ok {
				return
			}
			m.broadcastEvent(evt)
		}
	}
}

// broadcastEvent sends an event to the mesh
func (m *Manager) broadcastEvent(evt bus.Event) {
	if !m.isConnected() {
		return // Queue for later or drop
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return
	}

	msg := WebSocketMessage{
		Type:      MsgTypeEvent,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
		DeviceID:  m.deviceID,
		SessionID: evt.SessionID,
	}

	select {
	case m.sendCh <- msg:
	default:
		// Channel full, drop message
		log.Println("Mesh: Send queue full, dropping message")
	}
}

// handleIncoming processes messages from the WebSocket
func (m *Manager) handleIncoming(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		default:
		}

		if !m.isConnected() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		m.connMu.RLock()
		conn := m.wsConn
		m.connMu.RUnlock()

		var msg WebSocketMessage
		err := wsjson.Read(ctx, conn, &msg)
		if err != nil {
			log.Printf("Mesh: Read error: %v", err)
			m.disconnect()
			continue
		}

		m.handleMessage(msg)
	}
}

// handleMessage processes a received message
func (m *Manager) handleMessage(msg WebSocketMessage) {
	switch msg.Type {
	case MsgTypeEvent:
		m.handleRemoteEvent(msg)
	case MsgTypeSyncResponse:
		m.handleSyncResponse(msg)
	case MsgTypePing:
		m.sendPong()
	case MsgTypePresence:
		m.handlePresence(msg)
	default:
		log.Printf("Mesh: Unknown message type: %s", msg.Type)
	}
}

// handleRemoteEvent processes events from other devices
func (m *Manager) handleRemoteEvent(msg WebSocketMessage) {
	// Don't process our own events
	if msg.DeviceID == m.deviceID {
		return
	}

	var evt bus.Event
	if err := json.Unmarshal(msg.Payload, &evt); err != nil {
		return
	}

	// Mark as from remote device
	evt.Surface = "mesh:" + msg.DeviceID

	// Publish to local bus
	m.bus.Publish(evt)

	log.Printf("Mesh: Received event %s from %s", evt.Event, msg.DeviceID)
}

// handleSyncResponse processes sync responses
func (m *Manager) handleSyncResponse(msg WebSocketMessage) {
	var events []bus.Event
	if err := json.Unmarshal(msg.Payload, &events); err != nil {
		return
	}

	for _, evt := range events {
		if evt.SessionID != "" {
			m.bus.Publish(evt)
		}
	}
}

// handlePresence updates device presence info
func (m *Manager) handlePresence(msg WebSocketMessage) {
	log.Printf("Mesh: Device %s is online", msg.DeviceID)
}

// sendLoop sends queued messages
func (m *Manager) sendLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case msg := <-m.sendCh:
			if !m.isConnected() {
				continue
			}

			m.connMu.RLock()
			conn := m.wsConn
			m.connMu.RUnlock()

			if err := wsjson.Write(ctx, conn, msg); err != nil {
				log.Printf("Mesh: Write error: %v", err)
				m.disconnect()
			}
		}
	}
}

// presenceLoop sends periodic presence updates
func (m *Manager) presenceLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.sendPresence()
		}
	}
}

// sendPresence sends a presence heartbeat
func (m *Manager) sendPresence() {
	if !m.isConnected() {
		return
	}

	msg := WebSocketMessage{
		Type:      MsgTypePresence,
		Timestamp: time.Now().UTC(),
		DeviceID:  m.deviceID,
	}

	select {
	case m.sendCh <- msg:
	default:
	}
}

// sendPong responds to ping
func (m *Manager) sendPong() {
	msg := WebSocketMessage{
		Type:      MsgTypePong,
		Timestamp: time.Now().UTC(),
		DeviceID:  m.deviceID,
	}

	select {
	case m.sendCh <- msg:
	default:
	}
}

// RequestSync requests session sync from cloud
func (m *Manager) RequestSync(sessionID string) {
	if !m.isConnected() {
		return
	}

	payload, _ := json.Marshal(map[string]string{"session_id": sessionID})

	msg := WebSocketMessage{
		Type:      MsgTypeSyncRequest,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
		DeviceID:  m.deviceID,
		SessionID: sessionID,
	}

	select {
	case m.sendCh <- msg:
	default:
	}
}

// GetDeviceID returns this device's ID
func (m *Manager) GetDeviceID() string {
	return m.deviceID
}

// IsConnected returns mesh connection status
func (m *Manager) IsConnected() bool {
	return m.isConnected()
}

func generateDeviceID() string {
	return fmt.Sprintf("pryx-%d", time.Now().UnixNano())
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
