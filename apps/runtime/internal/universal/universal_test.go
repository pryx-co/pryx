package universal

import (
	"context"
	"testing"
	"time"
)

func TestUniversalHub(t *testing.T) {
	config := HubConfig{
		Name:              "test-hub",
		Namespace:         "test",
		AutoDetectEnabled: true,
		MaxConnections:    1000,
		ReconnectEnabled:  true,
	}

	hub := NewUniversalHub(nil, config)
	ctx := context.Background()

	// Test Start/Stop
	if err := hub.Start(ctx); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}

	if err := hub.Stop(ctx); err != nil {
		t.Fatalf("failed to stop hub: %v", err)
	}
}

func TestConnectionManager(t *testing.T) {
	cm := NewConnectionManager()
	ctx := context.Background()

	cm.Start(ctx)
	defer cm.Stop(ctx)

	// Test Add/Get/Remove
	conn := &AgentConnection{
		ID: "test-conn-1",
		AgentInfo: AgentInfo{
			Identity: AgentIdentity{
				ID:      "agent-1",
				Name:    "Test Agent",
				Version: "1.0.0",
			},
			Protocol: "websocket",
		},
		State:    ConnectionStateConnected,
		Protocol: "websocket",
	}

	cm.Add(conn)

	if cm.GetMetrics().TotalConnections != 1 {
		t.Errorf("expected 1 total connection, got %d", cm.GetMetrics().TotalConnections)
	}

	retrieved, ok := cm.Get("test-conn-1")
	if !ok {
		t.Error("expected to get connection")
	}
	if retrieved.ID != conn.ID {
		t.Errorf("expected connection ID %s, got %s", conn.ID, retrieved.ID)
	}

	cm.Remove("test-conn-1")

	if cm.GetMetrics().ActiveConnections != 0 {
		t.Errorf("expected 0 active connections, got %d", cm.GetMetrics().ActiveConnections)
	}
}

func TestCircuitBreaker(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 3,
		RecoveryTimeout:  1 * time.Second,
		HalfOpenRequests: 2,
	}

	cb := NewCircuitBreaker(config)

	// Initial state should be closed
	if cb.State() != CircuitBreakerClosed {
		t.Errorf("expected closed state, got %s", cb.State())
	}

	// Test AllowRequest
	if !cb.AllowRequest() {
		t.Error("expected request to be allowed in closed state")
	}

	// Record failures
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	// Should be open now
	if cb.State() != CircuitBreakerOpen {
		t.Errorf("expected open state after failures, got %s", cb.State())
	}

	// Should not allow requests in open state
	if cb.AllowRequest() {
		t.Error("expected request to be denied in open state")
	}

	// Wait for recovery timeout
	time.Sleep(2 * time.Second)

	// Should transition to half-open
	if !cb.AllowRequest() {
		t.Error("expected request to be allowed after recovery timeout")
	}

	// Record success
	cb.RecordSuccess()
	if cb.State() != CircuitBreakerHalfOpen {
		t.Errorf("expected half-open state, got %s", cb.State())
	}

	// Record more successes to close
	cb.RecordSuccess()
	if cb.State() != CircuitBreakerClosed {
		t.Errorf("expected closed state after successes, got %s", cb.State())
	}
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	registry.Start(ctx)
	defer registry.Stop(ctx)

	// Test Register/Get
	agent := &AgentInfo{
		Identity: AgentIdentity{
			ID:      "agent-reg-1",
			Name:    "Registry Test Agent",
			Version: "2.0.0",
		},
		Protocol:     "http",
		Capabilities: []string{"messaging"},
	}

	registry.Register(agent)

	retrieved, ok := registry.Get("agent-reg-1")
	if !ok {
		t.Error("expected to get agent from registry")
	}
	if retrieved.Identity.ID != agent.Identity.ID {
		t.Errorf("expected agent ID %s, got %s", agent.Identity.ID, retrieved.Identity.ID)
	}

	// Test List
	agents := registry.List()
	if len(agents) != 1 {
		t.Errorf("expected 1 agent in registry, got %d", len(agents))
	}

	// Test ListByProtocol
	httpAgents := registry.ListByProtocol("http")
	if len(httpAgents) != 1 {
		t.Errorf("expected 1 http agent, got %d", len(httpAgents))
	}

	// Test Unregister
	registry.Unregister("agent-reg-1")
	if registry.Count() != 0 {
		t.Errorf("expected 0 agents after unregister, got %d", registry.Count())
	}
}

func TestMessageRouter(t *testing.T) {
	router := NewMessageRouter()
	ctx := context.Background()

	router.Start(ctx)
	defer router.Stop(ctx)

	// Test AddRoute/RemoveRoute
	router.AddRoute("agent-a", "agent-b", "test", func(msg *UniversalMessage) error {
		return nil
	})

	if router.RouteCount() != 1 {
		t.Errorf("expected 1 route, got %d", router.RouteCount())
	}

	router.RemoveRoute("agent-a", "agent-b")

	if router.RouteCount() != 0 {
		t.Errorf("expected 0 routes after remove, got %d", router.RouteCount())
	}

	// Test Subscribe
	ch := router.Subscribe("*")
	defer router.Unsubscribe("*", ch)

	// Test Route
	msg := &UniversalMessage{
		ID:   "msg-router-test",
		From: AgentIdentity{ID: "from-1"},
		To:   AgentIdentity{ID: "to-1"},
	}

	routed, err := router.Route(ctx, msg)
	if err != nil {
		t.Fatalf("route failed: %v", err)
	}
	if routed {
		t.Error("expected route to fail (no direct route)")
	}
}

func TestTranslator(t *testing.T) {
	translator := NewTranslator()

	// Test ToUniversal
	openclawMsg := &OpenClawMessage{
		Type:   "request",
		ID:     "test-msg-1",
		From:   "openclaw-agent-1",
		To:     "pryx-agent-1",
		Action: "execute",
		Payload: map[string]interface{}{
			"command": "test",
		},
	}

	universalMsg := translator.ToUniversal(openclawMsg)

	if universalMsg.ID != openclawMsg.ID {
		t.Errorf("expected message ID %s, got %s", openclawMsg.ID, universalMsg.ID)
	}
	if universalMsg.MessageType != MessageTypeRequest {
		t.Errorf("expected message type request, got %s", universalMsg.MessageType)
	}
	if universalMsg.From.ID != openclawMsg.From {
		t.Errorf("expected from %s, got %s", openclawMsg.From, universalMsg.From.ID)
	}

	// Test ToOpenClaw
	backToOpenClaw := translator.ToOpenClaw(universalMsg)

	if backToOpenClaw.Type != "request" {
		t.Errorf("expected openclaw type request, got %s", backToOpenClaw.Type)
	}
	if backToOpenClaw.ID != universalMsg.ID {
		t.Errorf("expected openclaw ID %s, got %s", universalMsg.ID, backToOpenClaw.ID)
	}

	// Test NegotiateCapabilities
	localCaps := []string{"messaging", "tools", "sessions"}
	remoteCaps := []string{"tools", "sessions", "files"}
	common := translator.NegotiateCapabilities(localCaps, remoteCaps)

	if len(common) != 2 {
		t.Errorf("expected 2 common capabilities, got %d", len(common))
	}
}

func TestDetector(t *testing.T) {
	detector := NewDetector([]int{18789, 8080})
	ctx := context.Background()

	detector.Start(ctx)
	defer detector.Stop(ctx)

	// Test DetectAll (will likely return empty in test environment)
	agents, err := detector.DetectAll(ctx)
	if err != nil {
		t.Fatalf("detect all failed: %v", err)
	}
	// In a test environment, no agents should be detected
	_ = agents
}

func TestUniversalMessage(t *testing.T) {
	msg := &UniversalMessage{
		ID:     "test-univ-msg",
		From:   AgentIdentity{ID: "from-agent"},
		To:     AgentIdentity{ID: "to-agent"},
		Action: "test",
		Payload: map[string]interface{}{
			"data": "test",
		},
		Timestamp: time.Now().UTC(),
	}

	// Test PrettyPrint
	pretty := msg.PrettyPrint()
	if len(pretty) == 0 {
		t.Error("expected non-empty pretty print")
	}

	// Test MarshalJSON
	jsonData, err := msg.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if len(jsonData) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func TestCorrelationID(t *testing.T) {
	id1 := CorrelationID()
	id2 := CorrelationID()

	if id1 == "" {
		t.Error("expected non-empty correlation ID")
	}
	if id1 == id2 {
		t.Error("expected different correlation IDs")
	}
}

func TestConnectionMetrics(t *testing.T) {
	cm := NewConnectionManager()
	cm.Start(context.Background())
	defer cm.Stop(context.Background())

	// Add a connection
	conn := &AgentConnection{
		ID: "metrics-test",
		AgentInfo: AgentInfo{
			Identity: AgentIdentity{ID: "agent-metrics"},
		},
		Protocol: "websocket",
	}
	cm.Add(conn)

	// Update metrics
	cm.UpdateMetrics(conn.ID, 10, 5, 1000, 500, 1)

	metrics := cm.GetMetrics()
	if metrics.MessagesSent != 10 {
		t.Errorf("expected 10 messages sent, got %d", metrics.MessagesSent)
	}
	if metrics.MessagesReceived != 5 {
		t.Errorf("expected 5 messages received, got %d", metrics.MessagesReceived)
	}
	if metrics.BytesSent != 1000 {
		t.Errorf("expected 1000 bytes sent, got %d", metrics.BytesSent)
	}
	if metrics.BytesReceived != 500 {
		t.Errorf("expected 500 bytes received, got %d", metrics.BytesReceived)
	}
	if metrics.ErrorsTotal != 1 {
		t.Errorf("expected 1 error, got %d", metrics.ErrorsTotal)
	}
	if metrics.ProtocolStats["websocket"] != 1 {
		t.Errorf("expected 1 websocket protocol stat, got %d", metrics.ProtocolStats["websocket"])
	}
}
