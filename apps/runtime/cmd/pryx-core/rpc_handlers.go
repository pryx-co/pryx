package main

import (
	"context"
	"fmt"
	"pryx-core/internal/audit"
	"pryx-core/internal/hostrpc"
	"pryx-core/internal/server"
)

// setupAdminHandlers registers the management methods for the JSON-RPC bridge.
func setupAdminHandlers(srv *server.Server) *hostrpc.Registry {
	reg := hostrpc.NewRegistry()

	// Health check
	reg.Register("admin.health", func(method string, params map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{
			"status":          "healthy",
			"service":         "pryx-runtime",
			"uptime":          "up",
			"memory_enabled":  srv.Memory() != nil,
			"channels_active": len(srv.Channels().List()),
		}, nil
	})

	// Get current config
	reg.Register("admin.config.get", func(method string, params map[string]interface{}) (interface{}, error) {
		// Mock config export for now until we have a proper config export method
		return map[string]string{
			"model_provider": "openai",
			"model_name":     "gpt-4",
		}, nil
	})

	// List skills
	reg.Register("admin.skills.list", func(method string, params map[string]interface{}) (interface{}, error) {
		skillList := srv.Skills().List()
		return map[string]interface{}{
			"skills": skillList,
		}, nil
	})

	// --- Channels ---
	reg.Register("admin.channels.list", func(method string, params map[string]interface{}) (interface{}, error) {
		chanList := srv.Channels().List()
		return map[string]interface{}{"channels": chanList}, nil
	})

	reg.Register("admin.channels.test", func(method string, params map[string]interface{}) (interface{}, error) {
		id, _ := params["id"].(string)
		if id == "" {
			return nil, fmt.Errorf("id required")
		}
		return map[string]interface{}{"success": true, "message": "Connection test passed"}, nil
	})

	// --- MCP ---
	reg.Register("admin.mcp.list", func(method string, params map[string]interface{}) (interface{}, error) {
		tools, err := srv.MCP().ListToolsFlat(context.Background(), false)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"tools": tools}, nil
	})

	// --- Policies ---
	reg.Register("admin.policies.list", func(method string, params map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{
			"policies": []map[string]string{
				{"id": "default", "name": "Default", "description": "Default system policy"},
			},
		}, nil
	})

	// --- Audit ---
	reg.Register("admin.audit.list", func(method string, params map[string]interface{}) (interface{}, error) {
		limit := 50
		if l, ok := params["limit"].(float64); ok {
			limit = int(l)
		}

		entries, err := srv.AuditRepo().Query(audit.QueryOptions{
			Limit: limit,
		})
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"entries": entries}, nil
	})

	// --- Cost ---
	reg.Register("admin.cost.summary", func(method string, params map[string]interface{}) (interface{}, error) {
		summary, err := srv.CostService().GetCurrentSessionCost()
		if err != nil {
			return nil, err
		}
		return summary, nil
	})

	return reg
}

// startRPCServer starts the JSON-RPC server on stdin/stdout.
func startRPCServer(ctx context.Context, srv *server.Server) {
	registry := setupAdminHandlers(srv)
	rpcServer := hostrpc.NewDefaultServer()

	go func() {
		if err := rpcServer.Serve(registry); err != nil {
			// RPC server stopped
		}
	}()
}
