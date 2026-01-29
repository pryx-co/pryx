package mesh

import (
	"context"
	"testing"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/store"
)

func TestNewManager(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	eventBus := bus.New()
	store := &store.Store{}
	kc := keychain.New("pryx")

	manager := NewManager(cfg, eventBus, store, kc)

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.cfg != cfg {
		t.Error("NewManager() manager.cfg not set correctly")
	}

	if manager.bus != eventBus {
		t.Error("NewManager() manager.bus not set correctly")
	}

	if manager.store != store {
		t.Error("NewManager() manager.store not set correctly")
	}

	if manager.keychain != kc {
		t.Error("NewManager() manager.keychain not set correctly")
	}

	if manager.sendCh == nil {
		t.Error("NewManager() manager.sendCh not initialized")
	}

	if manager.stopCh == nil {
		t.Error("NewManager() manager.stopCh not initialized")
	}
}

func TestManager_GetDeviceID(t *testing.T) {
	tests := []struct {
		name           string
		setupDeviceID  string
		expectNewID    bool
	}{
		{
			name:          "existing device ID",
			setupDeviceID: "test-device-123",
			expectNewID:   false,
		},
		{
			name:          "new device ID generated",
			setupDeviceID: "",
			expectNewID:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				CloudAPIUrl: "https://api.pryx.io",
			}
			eventBus := bus.New()
			store := &store.Store{}
			kc := keychain.New("pryx")

			// Pre-set device ID if needed
			if tt.setupDeviceID != "" {
				kc.Set("device_id", tt.setupDeviceID)
			}

			manager := NewManager(cfg, eventBus, store, kc)

			// Start manager to trigger device ID initialization
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			manager.Start(ctx)

			deviceID := manager.GetDeviceID()

			if deviceID == "" {
				t.Error("GetDeviceID() returned empty string")
			}

			if !tt.expectNewID && deviceID != tt.setupDeviceID {
				t.Errorf("GetDeviceID() = %v, want %v", deviceID, tt.setupDeviceID)
			}

			if tt.expectNewID && deviceID == "" {
				t.Error("GetDeviceID() expected new ID but got empty")
			}

			// Verify device ID was stored
			storedID, _ := kc.Get("device_id")
			if storedID == "" {
				t.Error("Device ID not stored in keychain")
			}

			manager.Stop()
		})
	}
}

func TestManager_IsConnected(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	eventBus := bus.New()
	store := &store.Store{}
	kc := keychain.New("pryx")

	manager := NewManager(cfg, eventBus, store, kc)

	// Initially not connected
	if manager.IsConnected() {
		t.Error("IsConnected() = true, want false (initially)")
	}
}

func TestManager_Stop(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	eventBus := bus.New()
	store := &store.Store{}
	kc := keychain.New("pryx")

	manager := NewManager(cfg, eventBus, store, kc)

	// Start manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager.Start(ctx)

	// Give it time to start
	time.Sleep(50 * time.Millisecond)

	// Stop should not panic
	manager.Stop()

	// After stop, should not be connected
	if manager.IsConnected() {
		t.Error("IsConnected() = true after Stop()")
	}
}

func TestGenerateDeviceID(t *testing.T) {
	id1 := generateDeviceID()
	id2 := generateDeviceID()

	if id1 == "" {
		t.Error("generateDeviceID() returned empty string")
	}

	if id1 == id2 {
		t.Error("generateDeviceID() returned duplicate IDs")
	}

	// Check prefix
	if len(id1) < 6 || id1[:5] != "pryx-" {
		t.Errorf("generateDeviceID() = %s, want prefix 'pryx-'", id1)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want time.Duration
	}{
		{time.Second, 2 * time.Second, time.Second},
		{2 * time.Second, time.Second, time.Second},
		{time.Second, time.Second, time.Second},
		{0, time.Second, 0},
		{time.Second, 0, 0},
	}

	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("min(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestWebSocketMessageTypes(t *testing.T) {
	// Verify message type constants
	if MsgTypeEvent != "event" {
		t.Errorf("MsgTypeEvent = %v, want 'event'", MsgTypeEvent)
	}
	if MsgTypeSyncRequest != "sync_request" {
		t.Errorf("MsgTypeSyncRequest = %v, want 'sync_request'", MsgTypeSyncRequest)
	}
	if MsgTypeSyncResponse != "sync_response" {
		t.Errorf("MsgTypeSyncResponse = %v, want 'sync_response'", MsgTypeSyncResponse)
	}
	if MsgTypePing != "ping" {
		t.Errorf("MsgTypePing = %v, want 'ping'", MsgTypePing)
	}
	if MsgTypePong != "pong" {
		t.Errorf("MsgTypePong = %v, want 'pong'", MsgTypePong)
	}
	if MsgTypePresence != "presence" {
		t.Errorf("MsgTypePresence = %v, want 'presence'", MsgTypePresence)
	}
}

func TestManager_RequestSync(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	eventBus := bus.New()
	store := &store.Store{}
	kc := keychain.New("pryx")

	manager := NewManager(cfg, eventBus, store, kc)

	// Should not panic when not connected
	manager.RequestSync("test-session-123")
}

func TestManager_broadcastEvent(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	eventBus := bus.New()
	store := &store.Store{}
	kc := keychain.New("pryx")

	manager := NewManager(cfg, eventBus, store, kc)
	manager.deviceID = "test-device"

	// Create test event
	evt := bus.NewEvent(bus.EventSessionMessage, "test-session", map[string]string{"content": "hello"})

	// Should not panic when not connected (event is dropped)
	manager.broadcastEvent(evt)
}

func TestManager_handleMessage(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	eventBus := bus.New()
	store := &store.Store{}
	kc := keychain.New("pryx")

	manager := NewManager(cfg, eventBus, store, kc)
	manager.deviceID = "test-device"

	tests := []struct {
		name    string
		msgType string
		panic   bool
	}{
		{
			name:    "event message",
			msgType: MsgTypeEvent,
			panic:   false,
		},
		{
			name:    "sync response",
			msgType: MsgTypeSyncResponse,
			panic:   false,
		},
		{
			name:    "ping message",
			msgType: MsgTypePing,
			panic:   false,
		},
		{
			name:    "presence message",
			msgType: MsgTypePresence,
			panic:   false,
		},
		{
			name:    "unknown message type",
			msgType: "unknown",
			panic:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.panic && r == nil {
					t.Error("Expected panic but none occurred")
				} else if !tt.panic && r != nil {
					t.Errorf("Unexpected panic: %v", r)
				}
			}()

			msg := WebSocketMessage{
				Type:      tt.msgType,
				Payload:   []byte(`{}`),
				Timestamp: time.Now(),
				DeviceID:  "other-device",
			}

			manager.handleMessage(msg)
		})
	}
}

func TestManager_handleRemoteEvent(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	eventBus := bus.New()
	store := &store.Store{}
	kc := keychain.New("pryx")

	manager := NewManager(cfg, eventBus, store, kc)
	manager.deviceID = "test-device"

	// Subscribe to events
	events, cancel := eventBus.Subscribe(bus.EventSessionMessage)
	defer cancel()

	// Create message from other device
	msg := WebSocketMessage{
		Type:      MsgTypeEvent,
		Payload:   []byte(`{"event": "chat.request", "session_id": "test-session", "payload": {"content": "hello"}}`),
		Timestamp: time.Now(),
		DeviceID:  "other-device",
		SessionID: "test-session",
	}

	manager.handleRemoteEvent(msg)

	// Should receive event on bus
	select {
	case <-events:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("handleRemoteEvent() did not publish event to bus")
	}
}

func TestManager_handleRemoteEvent_OwnEvent(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	eventBus := bus.New()
	store := &store.Store{}
	kc := keychain.New("pryx")

	manager := NewManager(cfg, eventBus, store, kc)
	manager.deviceID = "test-device"

	// Subscribe to events
	events, cancel := eventBus.Subscribe(bus.EventSessionMessage)
	defer cancel()

	// Create message from same device (should be ignored)
	msg := WebSocketMessage{
		Type:      MsgTypeEvent,
		Payload:   []byte(`{"event": "chat.request", "session_id": "test-session"}`),
		Timestamp: time.Now(),
		DeviceID:  "test-device", // Same as manager.deviceID
		SessionID: "test-session",
	}

	manager.handleRemoteEvent(msg)

	// Should NOT receive event on bus (filtered out)
	select {
	case <-events:
		t.Error("handleRemoteEvent() published own event to bus (should be filtered)")
	case <-time.After(100 * time.Millisecond):
		// Expected - no event
	}
}

func TestManager_sendPresence(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	eventBus := bus.New()
	store := &store.Store{}
	kc := keychain.New("pryx")

	manager := NewManager(cfg, eventBus, store, kc)
	manager.deviceID = "test-device"

	// Should not panic when not connected
	manager.sendPresence()
}

func TestManager_sendPong(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	eventBus := bus.New()
	store := &store.Store{}
	kc := keychain.New("pryx")

	manager := NewManager(cfg, eventBus, store, kc)
	manager.deviceID = "test-device"

	// Should not panic
	manager.sendPong()
}

func TestManager_handlePresence(t *testing.T) {
	cfg := &config.Config{
		CloudAPIUrl: "https://api.pryx.io",
	}
	eventBus := bus.New()
	store := &store.Store{}
	kc := keychain.New("pryx")

	manager := NewManager(cfg, eventBus, store, kc)

	msg := WebSocketMessage{
		Type:      MsgTypePresence,
		Timestamp: time.Now(),
		DeviceID:  "other-device",
	}

	// Should not panic
	manager.handlePresence(msg)
}
