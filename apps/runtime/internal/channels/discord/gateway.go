package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"nhooyr.io/websocket"
)

const (
	gatewayURL        = "wss://gateway.discord.gg/?v=10&encoding=json"
	gatewayVersion    = 10
	identifyTimeout   = 30 * time.Second
	heartbeatTimeout  = 5 * time.Second
	reconnectMinDelay = 1 * time.Second
	reconnectMaxDelay = 60 * time.Second
)

// Gateway represents a connection to the Discord Gateway
type Gateway struct {
	config     *Config
	conn       *websocket.Conn
	httpClient *http.Client

	// Connection state
	connected int32
	sessionID string
	sequence  int64

	// Heartbeat
	heartbeatInterval time.Duration
	lastHeartbeat     time.Time
	lastAck           time.Time

	// Event handling
	eventHandler func(*GatewayEvent)

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Reconnection
	reconnectCount int
	mu             sync.RWMutex
}

// NewGateway creates a new Gateway connection
func NewGateway(config *Config, eventHandler func(*GatewayEvent)) *Gateway {
	return &Gateway{
		config:       config,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		eventHandler: eventHandler,
	}
}

// Connect establishes a connection to the Discord Gateway
func (g *Gateway) Connect(ctx context.Context) error {
	if g.IsConnected() {
		return fmt.Errorf("gateway already connected")
	}

	g.ctx, g.cancel = context.WithCancel(ctx)

	// Connect to Gateway
	conn, _, err := websocket.Dial(g.ctx, gatewayURL, nil)
	if err != nil {
		return fmt.Errorf("failed to dial gateway: %w", err)
	}

	g.conn = conn

	// Wait for Hello
	if err := g.handleHello(); err != nil {
		g.conn.Close(websocket.StatusInternalError, "hello failed")
		return err
	}

	// Identify
	if err := g.identify(); err != nil {
		g.conn.Close(websocket.StatusInternalError, "identify failed")
		return err
	}

	atomic.StoreInt32(&g.connected, 1)
	g.reconnectCount = 0

	// Start goroutines
	g.wg.Add(2)
	go g.readLoop()
	go g.heartbeatLoop()

	return nil
}

// Disconnect closes the Gateway connection
func (g *Gateway) Disconnect() error {
	if !g.IsConnected() {
		return nil
	}

	g.cancel()

	if g.conn != nil {
		g.conn.Close(websocket.StatusNormalClosure, "disconnecting")
	}

	g.wg.Wait()

	atomic.StoreInt32(&g.connected, 0)
	return nil
}

// IsConnected returns whether the gateway is connected
func (g *Gateway) IsConnected() bool {
	return atomic.LoadInt32(&g.connected) == 1
}

// handleHello processes the Hello payload
func (g *Gateway) handleHello() error {
	ctx, cancel := context.WithTimeout(g.ctx, identifyTimeout)
	defer cancel()

	payload, err := g.readPayload(ctx)
	if err != nil {
		return fmt.Errorf("failed to read hello: %w", err)
	}

	if payload.Op != int(GatewayOpHello) {
		return fmt.Errorf("expected hello, got op %d", payload.Op)
	}

	var hello HelloData
	if err := json.Unmarshal(payload.Data, &hello); err != nil {
		return fmt.Errorf("failed to unmarshal hello: %w", err)
	}

	g.heartbeatInterval = time.Duration(hello.HeartbeatInterval) * time.Millisecond

	return nil
}

// identify sends the Identify payload
func (g *Gateway) identify() error {
	identify := IdentifyData{
		Token: g.config.Token,
		Properties: Properties{
			OS:      "linux",
			Browser: "Pryx",
			Device:  "Pryx",
		},
		Intents:        g.config.Intents,
		LargeThreshold: g.config.LargeThreshold,
	}

	if g.config.Presence != nil {
		identify.Presence = g.config.Presence
	}

	if g.config.NumShards > 1 {
		identify.Shard = &[2]int{g.config.ShardID, g.config.NumShards}
	}

	data, err := marshalJSON(identify)
	if err != nil {
		return fmt.Errorf("failed to marshal identify: %w", err)
	}

	payload := GatewayPayload{
		Op:   int(GatewayOpIdentify),
		Data: data,
	}

	if err := g.writePayload(payload); err != nil {
		return fmt.Errorf("failed to send identify: %w", err)
	}

	return nil
}

// resume attempts to resume a previous session
func (g *Gateway) resume() error {
	g.mu.RLock()
	sessionID := g.sessionID
	sequence := g.sequence
	g.mu.RUnlock()

	if sessionID == "" {
		return fmt.Errorf("no session to resume")
	}

	resume := ResumeData{
		Token:     g.config.Token,
		SessionID: sessionID,
		Seq:       int(sequence),
	}

	data, err := marshalJSON(resume)
	if err != nil {
		return fmt.Errorf("failed to marshal resume: %w", err)
	}

	payload := GatewayPayload{
		Op:   int(GatewayOpResume),
		Data: data,
	}

	if err := g.writePayload(payload); err != nil {
		return fmt.Errorf("failed to send resume: %w", err)
	}

	return nil
}

// readLoop continuously reads payloads from the Gateway
func (g *Gateway) readLoop() {
	defer g.wg.Done()

	for {
		select {
		case <-g.ctx.Done():
			return
		default:
		}

		payload, err := g.readPayload(g.ctx)
		if err != nil {
			if g.ctx.Err() != nil {
				return
			}

			// Connection lost, attempt reconnect
			g.handleDisconnect()
			return
		}

		if err := g.handlePayload(payload); err != nil {
			// Log error but continue
			continue
		}
	}
}

// heartbeatLoop sends periodic heartbeats
func (g *Gateway) heartbeatLoop() {
	defer g.wg.Done()

	ticker := time.NewTicker(g.heartbeatInterval)
	defer ticker.Stop()

	// Send initial heartbeat
	g.sendHeartbeat()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			if !g.IsConnected() {
				return
			}

			// Check if we missed an ACK
			if !g.lastAck.IsZero() && g.lastHeartbeat.After(g.lastAck) {
				// Missed ACK, reconnect
				g.handleDisconnect()
				return
			}

			g.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a heartbeat payload
func (g *Gateway) sendHeartbeat() {
	g.mu.RLock()
	seq := g.sequence
	g.mu.RUnlock()

	var data json.RawMessage
	if seq > 0 {
		var err error
		data, err = marshalJSON(seq)
		if err != nil {
			return
		}
	} else {
		data = json.RawMessage("null")
	}

	payload := GatewayPayload{
		Op:   int(GatewayOpHeartbeat),
		Data: data,
	}

	if err := g.writePayload(payload); err != nil {
		return
	}

	g.lastHeartbeat = time.Now()
}

// handlePayload processes a received payload
func (g *Gateway) handlePayload(payload *GatewayPayload) error {
	switch GatewayOpCode(payload.Op) {
	case GatewayOpDispatch:
		return g.handleDispatch(payload)

	case GatewayOpHeartbeat:
		// Server requested heartbeat
		g.sendHeartbeat()

	case GatewayOpReconnect:
		// Server requested reconnect
		g.handleDisconnect()

	case GatewayOpInvalidSession:
		// Session invalid, clear and reconnect
		g.mu.Lock()
		g.sessionID = ""
		g.sequence = 0
		g.mu.Unlock()
		g.handleDisconnect()

	case GatewayOpHeartbeatACK:
		g.lastAck = time.Now()

	default:
		// Unknown op, ignore
	}

	return nil
}

// handleDispatch processes a dispatch event
func (g *Gateway) handleDispatch(payload *GatewayPayload) error {
	if payload.Sequence != nil {
		atomic.StoreInt64(&g.sequence, int64(*payload.Sequence))
	}

	// Handle Ready event specially
	if payload.Type == "READY" {
		var ready ReadyData
		if err := json.Unmarshal(payload.Data, &ready); err != nil {
			return err
		}
		g.mu.Lock()
		g.sessionID = ready.SessionID
		g.mu.Unlock()
	}

	// Handle Resumed event
	if payload.Type == "RESUMED" {
		g.reconnectCount = 0
	}

	// Dispatch to handler
	if g.eventHandler != nil {
		event := &GatewayEvent{
			Op:       payload.Op,
			Type:     payload.Type,
			Data:     payload.Data,
			Sequence: 0,
		}
		if payload.Sequence != nil {
			event.Sequence = *payload.Sequence
		}

		go g.eventHandler(event)
	}

	return nil
}

// handleDisconnect handles a disconnection and attempts reconnect
func (g *Gateway) handleDisconnect() {
	if !atomic.CompareAndSwapInt32(&g.connected, 1, 0) {
		return
	}

	if g.conn != nil {
		g.conn.Close(websocket.StatusGoingAway, "reconnecting")
	}

	// Attempt reconnect with backoff
	g.reconnectCount++
	delay := g.calculateReconnectDelay()

	time.Sleep(delay)

	// Try to reconnect
	for attempt := 0; attempt < 5; attempt++ {
		if err := g.reconnect(); err == nil {
			return
		}

		delay = min(delay*2, reconnectMaxDelay)
		time.Sleep(delay)
	}
}

// reconnect attempts to reconnect to the Gateway
func (g *Gateway) reconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect
	conn, _, err := websocket.Dial(ctx, gatewayURL, nil)
	if err != nil {
		return err
	}

	g.conn = conn
	g.ctx, g.cancel = context.WithCancel(context.Background())

	// Wait for Hello
	if err := g.handleHello(); err != nil {
		return err
	}

	// Try to resume, otherwise identify
	if err := g.resume(); err != nil {
		if err := g.identify(); err != nil {
			return err
		}
	}

	atomic.StoreInt32(&g.connected, 1)

	// Restart goroutines
	g.wg.Add(2)
	go g.readLoop()
	go g.heartbeatLoop()

	return nil
}

// calculateReconnectDelay calculates the delay before reconnecting
func (g *Gateway) calculateReconnectDelay() time.Duration {
	delay := reconnectMinDelay * (1 << uint(g.reconnectCount))
	if delay > reconnectMaxDelay {
		delay = reconnectMaxDelay
	}
	return delay
}

// readPayload reads a single payload from the connection
func (g *Gateway) readPayload(ctx context.Context) (*GatewayPayload, error) {
	_, data, err := g.conn.Read(ctx)
	if err != nil {
		return nil, err
	}

	var payload GatewayPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

// writePayload writes a payload to the connection
func (g *Gateway) writePayload(payload GatewayPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(g.ctx, heartbeatTimeout)
	defer cancel()

	return g.conn.Write(ctx, websocket.MessageText, data)
}

// GetSessionID returns the current session ID
func (g *Gateway) GetSessionID() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.sessionID
}

// GetSequence returns the current sequence number
func (g *Gateway) GetSequence() int64 {
	return atomic.LoadInt64(&g.sequence)
}

// marshalJSON marshals data to JSON, returning an error on failure
func marshalJSON(v interface{}) (json.RawMessage, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return data, nil
}

// min returns the minimum of two durations
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
