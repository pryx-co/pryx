package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestClient_Initialize(t *testing.T) {
	server := NewMockServer()
	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	ctx := context.Background()
	err := client.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !server.IsInitialized() {
		t.Error("Server should be initialized")
	}

	if !client.initialized.Load() {
		t.Error("Client should be marked as initialized")
	}
}

func TestClient_Initialize_AlreadyInitialized(t *testing.T) {
	server := NewMockServer()
	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	ctx := context.Background()

	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("First initialize failed: %v", err)
	}

	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("Second initialize should be no-op but failed: %v", err)
	}
}

func TestClient_ListTools(t *testing.T) {
	server := NewMockServer()
	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	ctx := context.Background()
	tools, err := client.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	foundEcho := false
	foundAdd := false
	for _, tool := range tools {
		if tool.Name == "echo" {
			foundEcho = true
		}
		if tool.Name == "add" {
			foundAdd = true
		}
	}

	if !foundEcho {
		t.Error("Expected to find 'echo' tool")
	}
	if !foundAdd {
		t.Error("Expected to find 'add' tool")
	}
}

func TestClient_ListTools_WithPagination(t *testing.T) {
	server := NewMockServer()

	for i := 0; i < 10; i++ {
		server.AddTool(Tool{
			Name:        "tool" + string(rune('0'+i)),
			Title:       "Tool " + string(rune('0'+i)),
			Description: "Test tool",
		})
	}

	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	server.ListToolsFunc = func(ctx context.Context) ([]Tool, error) {
		return server.defaultListTools(ctx)
	}

	ctx := context.Background()
	tools, err := client.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	if len(tools) < 2 {
		t.Errorf("Expected at least 2 tools, got %d", len(tools))
	}
}

func TestClient_CallTool(t *testing.T) {
	server := NewMockServer()
	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	ctx := context.Background()
	result, err := client.CallTool(ctx, "echo", map[string]interface{}{"message": "hello"})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if result.IsError {
		t.Error("Result should not be an error")
	}

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}

	if result.Content[0].Text != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result.Content[0].Text)
	}

	if server.GetCallCount("echo") != 1 {
		t.Errorf("Expected 1 call to echo, got %d", server.GetCallCount("echo"))
	}
}

func TestClient_CallTool_Add(t *testing.T) {
	server := NewMockServer()
	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	ctx := context.Background()
	result, err := client.CallTool(ctx, "add", map[string]interface{}{"a": 5.0, "b": 3.0})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}

	if result.Content[0].Text != "8" {
		t.Errorf("Expected '8', got '%s'", result.Content[0].Text)
	}
}

func TestClient_CallTool_Error(t *testing.T) {
	server := NewMockServer()
	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	server.CallToolFunc = func(ctx context.Context, name string, args map[string]interface{}) (ToolResult, error) {
		return ToolResult{}, errors.New("tool execution failed")
	}

	ctx := context.Background()
	_, err := client.CallTool(ctx, "echo", map[string]interface{}{"message": "test"})
	if err == nil {
		t.Error("Expected error from tool call")
	}
}

func TestClient_CallTool_NotInitialized(t *testing.T) {
	server := NewMockServer()
	server.initialized.Store(false)

	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	// Client auto-initializes on first call, so this should succeed
	ctx := context.Background()
	result, err := client.CallTool(ctx, "echo", map[string]interface{}{"message": "test"})
	if err != nil {
		t.Errorf("Client should auto-initialize: %v", err)
	}
	if result.Content == nil {
		t.Error("Expected result content after auto-initialization")
	}
}

func TestClient_Ping(t *testing.T) {
	server := NewMockServer()
	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	ctx := context.Background()
	err := client.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestClient_Ping_NotInitialized(t *testing.T) {
	server := NewMockServer()
	server.initialized.Store(false)

	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	// Client auto-initializes on first call, so this should succeed
	ctx := context.Background()
	err := client.Ping(ctx)
	if err != nil {
		t.Errorf("Client should auto-initialize: %v", err)
	}
}

func TestClient_Timeout(t *testing.T) {
	server := NewMockServer()
	transport := NewMockTransport(server)
	transport.SetCallDelay(100)

	client := NewClient(transport, "2024-11-05")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.ListTools(ctx)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestClient_TransportClosed(t *testing.T) {
	server := NewMockServer()
	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	if err := transport.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	ctx := context.Background()
	_, err := client.ListTools(ctx)
	if err == nil {
		t.Error("Expected error when transport is closed")
	}
}

func TestClient_DefaultProtocolVersion(t *testing.T) {
	server := NewMockServer()
	transport := NewMockTransport(server)
	client := NewClient(transport, "")

	if client.protocolVersion != "2025-11-25" {
		t.Errorf("Expected default version '2025-11-25', got '%s'", client.protocolVersion)
	}
}

func TestClient_ServerCapabilities(t *testing.T) {
	server := NewMockServer()
	server.InitializeFunc = func(ctx context.Context, req RPCRequest) RPCResponse {
		result := map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
				"resources": map[string]interface{}{
					"subscribe": true,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "test-server",
				"version": "1.0.0",
			},
		}
		b, _ := json.Marshal(result)
		return RPCResponse{JSONRPC: "2.0", ID: mustMarshalID(req.ID), Result: b}
	}

	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	ctx := context.Background()
	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	client.mu.RLock()
	caps := client.serverCapabilities
	client.mu.RUnlock()

	if len(caps) == 0 {
		t.Error("Expected server capabilities to be stored")
	}
}

func TestClient_Initialize_InvalidResponse(t *testing.T) {
	server := NewMockServer()
	server.InitializeFunc = func(ctx context.Context, req RPCRequest) RPCResponse {
		return RPCResponse{
			JSONRPC: "2.0",
			ID:      mustMarshalID(req.ID),
			Error:   &RPCError{Code: -32000, Message: "initialization failed"},
		}
	}

	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	ctx := context.Background()
	err := client.Initialize(ctx)
	if err == nil {
		t.Error("Expected error from failed initialization")
	}
}

func TestClient_ListTools_MethodNotFound(t *testing.T) {
	server := NewMockServer()
	server.ListToolsFunc = func(ctx context.Context) ([]Tool, error) {
		return nil, errors.New("method not available")
	}

	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	ctx := context.Background()
	_, err := client.ListTools(ctx)
	if err == nil {
		t.Error("Expected error when method not found")
	}
}

func TestClient_Close(t *testing.T) {
	server := NewMockServer()
	transport := NewMockTransport(server)
	client := NewClient(transport, "2024-11-05")

	if err := client.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if !transport.IsClosed() {
		t.Error("Transport should be closed")
	}
}
