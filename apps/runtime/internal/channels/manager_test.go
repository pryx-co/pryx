package channels

import (
	"context"
	"sync"
	"testing"
	"time"

	"pryx-core/internal/bus"
)

type mockChannel struct {
	mu               sync.Mutex
	id               string
	status           Status
	connectCalled    bool
	disconnectCalled bool
}

func (m *mockChannel) ID() string {
	return m.id
}

func (m *mockChannel) Type() string {
	return "mock"
}

func (m *mockChannel) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectCalled = true
	m.status = StatusConnected
	return nil
}

func (m *mockChannel) Disconnect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.disconnectCalled = true
	m.status = StatusDisconnected
	return nil
}

func (m *mockChannel) Send(ctx context.Context, msg Message) error {
	return nil
}

func (m *mockChannel) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

func TestManager_Register(t *testing.T) {
	b := bus.New()
	m := NewManager(b)
	defer m.Shutdown()

	c := &mockChannel{id: "test-channel", status: StatusDisconnected}

	if err := m.Register(c); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify retrieval
	retrieved, ok := m.Get("test-channel")
	if !ok {
		t.Error("Channel not found after registration")
	}
	if retrieved.ID() != "test-channel" {
		t.Errorf("Expected ID test-channel, got %s", retrieved.ID())
	}

	// Verify duplicate registration fails
	if err := m.Register(c); err == nil {
		t.Error("Expected error for duplicate registration")
	}
}

func TestManager_AutoConnect(t *testing.T) {
	b := bus.New()
	m := NewManager(b)
	defer m.Shutdown()

	c := &mockChannel{id: "connect-test", status: StatusDisconnected}

	// Subscribe to status events
	statusCh, _ := b.Subscribe(bus.EventChannelStatus)

	if err := m.Register(c); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Should receive connected event
	select {
	case event := <-statusCh:
		if event.Event != bus.EventChannelStatus {
			t.Errorf("Expected status event, got %v", event.Type)
		}
		data := event.Payload.(map[string]interface{})
		if data["channel_id"] != "connect-test" {
			t.Errorf("Expected channel_id connect-test, got %v", data["channel_id"])
		}
		if data["status"] != StatusConnected {
			t.Errorf("Expected status connected, got %v", data["status"])
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for status event")
	}

	// Verify Connect was called
	// Note: Register calls maintainConnection in goroutine, so we used the event to sync.
	c.mu.Lock()
	connectCalled := c.connectCalled
	c.mu.Unlock()
	if !connectCalled {
		t.Error("Connect was not called on channel")
	}
}
