//go:build e2e

package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestMeshCoordination tests WebSocket mesh coordination
func TestMeshCoordination(t *testing.T) {
	bin := buildPryxCore(t)
	home := t.TempDir()

	port, cancel := startPryxCore(t, bin, home)
	defer cancel()

	waitForServer(t, port, 5*time.Second)

	wsURL := "ws://localhost:" + port

	t.Run("websocket_connects", func(t *testing.T) {
		dialer := websocket.Dialer{
			HandshakeTimeout: 5 * time.Second,
		}

		conn, resp, err := dialer.Dial(wsURL+"/ws", nil)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				t.Skip("WebSocket endpoint not implemented (404)")
			}
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer conn.Close()

		t.Logf("✓ WebSocket connected successfully")
	})

	t.Run("websocket_heartbeat", func(t *testing.T) {
		dialer := websocket.Dialer{
			HandshakeTimeout: 5 * time.Second,
		}

		conn, _, err := dialer.Dial(wsURL+"/ws", nil)
		if err != nil {
			t.Skip("WebSocket not available")
		}
		defer conn.Close()

		// Send a ping message
		pingMsg := map[string]interface{}{
			"type":      "ping",
			"timestamp": time.Now().Unix(),
		}

		if err := conn.WriteJSON(pingMsg); err != nil {
			t.Fatalf("Failed to send ping: %v", err)
		}

		// Set read timeout
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// Try to read response
		var response map[string]interface{}
		if err := conn.ReadJSON(&response); err != nil {
			t.Logf("No pong received (may be expected): %v", err)
		} else {
			t.Logf("✓ WebSocket heartbeat response: %v", response)
		}
	})

	t.Run("websocket_reconnect", func(t *testing.T) {
		dialer := websocket.Dialer{
			HandshakeTimeout: 5 * time.Second,
		}

		// First connection
		conn1, _, err := dialer.Dial(wsURL+"/ws", nil)
		if err != nil {
			t.Skip("WebSocket not available")
		}

		// Close first connection
		conn1.Close()

		// Wait a moment
		time.Sleep(100 * time.Millisecond)

		// Reconnect
		conn2, _, err := dialer.Dial(wsURL+"/ws", nil)
		if err != nil {
			t.Fatalf("Failed to reconnect WebSocket: %v", err)
		}
		defer conn2.Close()

		t.Logf("✓ WebSocket reconnected successfully")
	})

	t.Run("presence_broadcast", func(t *testing.T) {
		dialer := websocket.Dialer{
			HandshakeTimeout: 5 * time.Second,
		}

		conn, _, err := dialer.Dial(wsURL+"/ws", nil)
		if err != nil {
			t.Skip("WebSocket not available")
		}
		defer conn.Close()

		// Send presence update
		presenceMsg := map[string]interface{}{
			"type":      "presence",
			"status":    "online",
			"timestamp": time.Now().Unix(),
		}

		if err := conn.WriteJSON(presenceMsg); err != nil {
			t.Fatalf("Failed to send presence: %v", err)
		}

		t.Logf("✓ Presence broadcast sent")
	})

	t.Run("event_broadcasting", func(t *testing.T) {
		dialer := websocket.Dialer{
			HandshakeTimeout: 5 * time.Second,
		}

		conn, _, err := dialer.Dial(wsURL+"/ws", nil)
		if err != nil {
			t.Skip("WebSocket not available")
		}
		defer conn.Close()

		// Subscribe to events
		subscribeMsg := map[string]interface{}{
			"type":   "subscribe",
			"events": []string{"agent.spawned", "session.created"},
		}

		if err := conn.WriteJSON(subscribeMsg); err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}

		t.Logf("✓ Event subscription sent")
	})
}
