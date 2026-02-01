package agentbus

import (
	"context"
	"testing"
	"time"

	"pryx-core/internal/bus"
)

func TestRegistryManager(t *testing.T) {
	b := bus.New()
	rm := NewRegistryManager(b)
	ctx := context.Background()

	// Test Start/Stop
	if err := rm.Start(ctx); err != nil {
		t.Fatalf("failed to start registry: %v", err)
	}
	if err := rm.Stop(ctx); err != nil {
		t.Fatalf("failed to stop registry: %v", err)
	}
}

func TestRegistryManagerRegister(t *testing.T) {
	b := bus.New()
	rm := NewRegistryManager(b)
	ctx := context.Background()

	rm.Start(ctx)
	defer rm.Stop(ctx)

	agent := &AgentInfo{
		Identity: AgentIdentity{
			ID:      "test-agent-1",
			Name:    "Test Agent",
			Version: "1.0.0",
			Tags:    []string{"test", "demo"},
		},
		Endpoint: EndpointInfo{
			Type: "websocket",
			URL:  "ws://localhost:8080",
		},
		Protocol:     "websocket",
		HealthStatus: "healthy",
	}

	// Register agent
	regAgent, err := rm.Register(ctx, agent)
	if err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	if regAgent.Identity.ID != agent.Identity.ID {
		t.Errorf("expected agent ID %s, got %s", agent.Identity.ID, regAgent.Identity.ID)
	}

	// Test Get
	retrieved, err := rm.Get(ctx, agent.Identity.ID)
	if err != nil {
		t.Fatalf("failed to get agent: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected agent, got nil")
	}

	// Test List
	agents, err := rm.List(ctx)
	if err != nil {
		t.Fatalf("failed to list agents: %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}

	// Test Count
	if rm.Count() != 1 {
		t.Errorf("expected count 1, got %d", rm.Count())
	}

	// Test Unregister
	if err := rm.Unregister(ctx, agent.Identity.ID); err != nil {
		t.Fatalf("failed to unregister agent: %v", err)
	}

	if rm.Count() != 0 {
		t.Errorf("expected count 0 after unregister, got %d", rm.Count())
	}
}

func TestRegistryManagerFiltering(t *testing.T) {
	b := bus.New()
	rm := NewRegistryManager(b)
	ctx := context.Background()

	rm.Start(ctx)
	defer rm.Stop(ctx)

	// Create test agents
	agents := []*AgentInfo{
		{
			Identity: AgentIdentity{ID: "agent-1", Name: "Agent One", Namespace: "ns1", Tags: []string{"fast"}},
			Protocol: "http",
		},
		{
			Identity: AgentIdentity{ID: "agent-2", Name: "Agent Two", Namespace: "ns1", Tags: []string{"slow"}},
			Protocol: "websocket",
		},
		{
			Identity: AgentIdentity{ID: "agent-3", Name: "Agent Three", Namespace: "ns2", Tags: []string{"fast"}},
			Protocol: "http",
		},
	}

	for _, agent := range agents {
		rm.Register(ctx, agent)
	}

	// Test ListByProtocol
	httpAgents, err := rm.ListByProtocol(ctx, "http")
	if err != nil {
		t.Fatalf("failed to list by protocol: %v", err)
	}
	if len(httpAgents) != 2 {
		t.Errorf("expected 2 http agents, got %d", len(httpAgents))
	}

	// Test ListByNamespace
	ns1Agents, err := rm.ListByNamespace(ctx, "ns1")
	if err != nil {
		t.Fatalf("failed to list by namespace: %v", err)
	}
	if len(ns1Agents) != 2 {
		t.Errorf("expected 2 ns1 agents, got %d", len(ns1Agents))
	}

	// Test ListByTag
	fastAgents, err := rm.ListByTag(ctx, "fast")
	if err != nil {
		t.Fatalf("failed to list by tag: %v", err)
	}
	if len(fastAgents) != 2 {
		t.Errorf("expected 2 fast agents, got %d", len(fastAgents))
	}
}

func TestConnectionManager(t *testing.T) {
	b := bus.New()
	cm := NewConnectionManager(b)
	ctx := context.Background()

	if err := cm.Start(ctx); err != nil {
		t.Fatalf("failed to start connection manager: %v", err)
	}
	if err := cm.Stop(ctx); err != nil {
		t.Fatalf("failed to stop connection manager: %v", err)
	}
}

func TestConnectionManagerAddRemove(t *testing.T) {
	b := bus.New()
	cm := NewConnectionManager(b)
	ctx := context.Background()

	cm.Start(ctx)
	defer cm.Stop(ctx)

	conn := &AgentConnection{
		ID: "conn-1",
		AgentInfo: AgentInfo{
			Identity: AgentIdentity{ID: "agent-1", Name: "Test Agent"},
		},
		State:    ConnectionStateConnected,
		Protocol: "websocket",
	}

	// Add connection
	cm.Add(ctx, conn)

	if cm.Count() != 1 {
		t.Errorf("expected 1 connection, got %d", cm.Count())
	}

	// Get connection
	retrieved, err := cm.Get(ctx, conn.ID)
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected connection, got nil")
	}

	// Remove connection
	cm.Remove(ctx, conn.ID)

	if cm.Count() != 0 {
		t.Errorf("expected 0 connections after remove, got %d", cm.Count())
	}
}

func TestConnectionManagerMetrics(t *testing.T) {
	b := bus.New()
	cm := NewConnectionManager(b)
	ctx := context.Background()

	cm.Start(ctx)
	defer cm.Stop(ctx)

	conn := &AgentConnection{
		ID: "conn-metrics",
		AgentInfo: AgentInfo{
			Identity: AgentIdentity{ID: "agent-metrics", Name: "Metrics Agent"},
		},
		State:    ConnectionStateConnected,
		Protocol: "http",
	}

	cm.Add(ctx, conn)

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
	if metrics.ErrorsTotal != 1 {
		t.Errorf("expected 1 error, got %d", metrics.ErrorsTotal)
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

func TestMessageRouter(t *testing.T) {
	b := bus.New()
	mr := NewMessageRouter(b)
	ctx := context.Background()

	if err := mr.Start(ctx); err != nil {
		t.Fatalf("failed to start message router: %v", err)
	}
	if err := mr.Stop(ctx); err != nil {
		t.Fatalf("failed to stop message router: %v", err)
	}
}

func TestMessageRouterRoutes(t *testing.T) {
	b := bus.New()
	mr := NewMessageRouter(b)
	ctx := context.Background()

	mr.Start(ctx)
	defer mr.Stop(ctx)

	// Add a route
	mr.AddRoute("agent-1", "agent-2", "test", 1, func(msg *UniversalMessage) error {
		return nil
	})

	if mr.RouteCount() != 1 {
		t.Errorf("expected 1 route, got %d", mr.RouteCount())
	}

	// Remove route
	mr.RemoveRoute("agent-1", "agent-2")

	if mr.RouteCount() != 0 {
		t.Errorf("expected 0 routes after remove, got %d", mr.RouteCount())
	}
}

func TestMessageRouterSubscribe(t *testing.T) {
	b := bus.New()
	mr := NewMessageRouter(b)
	ctx := context.Background()

	mr.Start(ctx)
	defer mr.Stop(ctx)

	// Subscribe to all messages
	ch := mr.Subscribe("*")
	defer mr.Unsubscribe("*", ch)

	// Create test message
	msg := &UniversalMessage{
		ID:   "msg-1",
		From: AgentIdentity{ID: "from-1"},
		To:   AgentIdentity{ID: "to-1"},
	}

	// Broadcast message
	mr.Broadcast(msg)

	// Check if received (with timeout)
	select {
	case received := <-ch:
		if received.ID != msg.ID {
			t.Errorf("expected message ID %s, got %s", msg.ID, received.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for message")
	}
}

func TestPackageManager(t *testing.T) {
	b := bus.New()
	pm := NewPackageManager(b, "")
	ctx := context.Background()

	if err := pm.Start(ctx); err != nil {
		t.Fatalf("failed to start package manager: %v", err)
	}
	if err := pm.Stop(ctx); err != nil {
		t.Fatalf("failed to stop package manager: %v", err)
	}
}

func TestPackageManagerInstallUninstall(t *testing.T) {
	b := bus.New()
	pm := NewPackageManager(b, "")
	ctx := context.Background()

	pm.Start(ctx)
	defer pm.Stop(ctx)

	pkg := AgentPackage{
		Name:    "test-pkg",
		Version: "1.0.0",
	}

	// Install package
	if err := pm.Install(ctx, pkg); err != nil {
		t.Fatalf("failed to install package: %v", err)
	}

	if pm.Count() != 1 {
		t.Errorf("expected 1 package, got %d", pm.Count())
	}

	// Get package
	retrieved, err := pm.Get(ctx, pkg.Name)
	if err != nil {
		t.Fatalf("failed to get package: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected package, got nil")
	}

	// Search package
	results := pm.Search(ctx, "test")
	if len(results) != 1 {
		t.Errorf("expected 1 search result, got %d", len(results))
	}

	// Uninstall package
	if err := pm.Uninstall(ctx, pkg); err != nil {
		t.Fatalf("failed to uninstall package: %v", err)
	}

	if pm.Count() != 0 {
		t.Errorf("expected 0 packages after uninstall, got %d", pm.Count())
	}
}

func TestDetectionManager(t *testing.T) {
	b := bus.New()
	dm := NewDetectionManager(b)
	ctx := context.Background()

	if err := dm.Start(ctx); err != nil {
		t.Fatalf("failed to start detection manager: %v", err)
	}
	if err := dm.Stop(ctx); err != nil {
		t.Fatalf("failed to stop detection manager: %v", err)
	}
}

func TestDetectionManagerDetectAll(t *testing.T) {
	b := bus.New()
	dm := NewDetectionManager(b)
	ctx := context.Background()

	dm.Start(ctx)
	defer dm.Stop(ctx)

	// Test with empty adapters
	adapters := make(map[string]AgentAdapter)
	agents, err := dm.DetectAll(ctx, adapters)
	if err != nil {
		t.Fatalf("detect all failed: %v", err)
	}
	if len(agents) != 0 {
		t.Errorf("expected 0 agents with no adapters, got %d", len(agents))
	}
}
