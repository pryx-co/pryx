package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type BundledTransport struct {
	provider ToolProvider

	mu          sync.Mutex
	initialized bool
}

func NewBundledTransport(provider ToolProvider) *BundledTransport {
	return &BundledTransport{provider: provider}
}

func (t *BundledTransport) Close() error {
	return nil
}

func (t *BundledTransport) Notify(ctx context.Context, notif RPCNotification) error {
	_ = ctx
	if strings.TrimSpace(notif.Method) == "initialized" {
		t.mu.Lock()
		t.initialized = true
		t.mu.Unlock()
	}
	return nil
}

func (t *BundledTransport) Call(ctx context.Context, req RPCRequest) (RPCResponse, error) {
	if t.provider == nil {
		return RPCResponse{}, errors.New("missing provider")
	}

	switch strings.TrimSpace(req.Method) {
	case "initialize":
		result := map[string]interface{}{
			"protocolVersion": "2025-11-25",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": false,
				},
			},
			"serverInfo": t.provider.ServerInfo(),
		}
		b, _ := json.Marshal(result)
		return RPCResponse{JSONRPC: "2.0", ID: mustMarshalID(req.ID), Result: b}, nil

	case "tools/list":
		tools, err := t.provider.ListTools(ctx)
		if err != nil {
			return RPCResponse{JSONRPC: "2.0", ID: mustMarshalID(req.ID), Error: &RPCError{Code: -32000, Message: err.Error()}}, nil
		}
		b, _ := json.Marshal(map[string]interface{}{"tools": tools})
		return RPCResponse{JSONRPC: "2.0", ID: mustMarshalID(req.ID), Result: b}, nil

	case "tools/call":
		t.mu.Lock()
		initialized := t.initialized
		t.mu.Unlock()
		if !initialized {
			return RPCResponse{JSONRPC: "2.0", ID: mustMarshalID(req.ID), Error: &RPCError{Code: -32000, Message: "not initialized"}}, nil
		}

		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if b, err := json.Marshal(req.Params); err == nil {
			_ = json.Unmarshal(b, &params)
		}
		if strings.TrimSpace(params.Name) == "" {
			return RPCResponse{JSONRPC: "2.0", ID: mustMarshalID(req.ID), Error: &RPCError{Code: -32602, Message: "missing tool name"}}, nil
		}
		if params.Arguments == nil {
			params.Arguments = map[string]interface{}{}
		}

		res, err := t.provider.CallTool(ctx, params.Name, params.Arguments)
		if err != nil {
			return RPCResponse{JSONRPC: "2.0", ID: mustMarshalID(req.ID), Error: &RPCError{Code: -32000, Message: err.Error()}}, nil
		}
		b, _ := json.Marshal(res)
		return RPCResponse{JSONRPC: "2.0", ID: mustMarshalID(req.ID), Result: b}, nil

	default:
		return RPCResponse{JSONRPC: "2.0", ID: mustMarshalID(req.ID), Error: &RPCError{Code: -32601, Message: "method not found"}}, nil
	}
}

func mustMarshalID(v interface{}) json.RawMessage {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case json.RawMessage:
		return t
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return json.RawMessage([]byte(fmt.Sprintf("%q", err.Error())))
		}
		return b
	}
}
