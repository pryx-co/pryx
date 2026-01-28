package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPTransport_Call_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := RPCRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch req.Method {
		case "initialize":
			_ = json.NewEncoder(w).Encode(RPCResponse{JSONRPC: "2.0", ID: mustJSON(req.ID), Result: mustJSON(map[string]interface{}{"capabilities": map[string]interface{}{}})})
		case "tools/list":
			_ = json.NewEncoder(w).Encode(RPCResponse{JSONRPC: "2.0", ID: mustJSON(req.ID), Result: mustJSON(map[string]interface{}{"tools": []map[string]interface{}{{"name": "t1"}}})})
		case "tools/call":
			_ = json.NewEncoder(w).Encode(RPCResponse{JSONRPC: "2.0", ID: mustJSON(req.ID), Result: mustJSON(map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": "ok"}}, "isError": false})})
		default:
			_ = json.NewEncoder(w).Encode(RPCResponse{JSONRPC: "2.0", ID: mustJSON(req.ID), Error: &RPCError{Code: -32601, Message: "method not found"}})
		}
	}))
	defer srv.Close()

	tr := NewHTTPTransport(srv.URL, nil)
	c := NewClient(tr, "2025-11-25")

	ctx := context.Background()
	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "t1" {
		t.Fatalf("unexpected tools: %#v", tools)
	}

	res, err := c.CallTool(ctx, "t1", map[string]interface{}{})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if len(res.Content) != 1 || res.Content[0].Text != "ok" {
		t.Fatalf("unexpected result: %#v", res)
	}
}

func TestHTTPTransport_Call_SSE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := RPCRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		var result interface{}
		switch req.Method {
		case "initialize":
			result = map[string]interface{}{"capabilities": map[string]interface{}{}}
		case "tools/list":
			result = map[string]interface{}{"tools": []map[string]interface{}{{"name": "t1"}}}
		default:
			result = map[string]interface{}{}
		}

		resp := RPCResponse{JSONRPC: "2.0", ID: mustJSON(req.ID), Result: mustJSON(result)}
		b, _ := json.Marshal(resp)
		_, _ = w.Write([]byte("data: "))
		_, _ = w.Write(b)
		_, _ = w.Write([]byte("\n\n"))
	}))
	defer srv.Close()

	tr := NewHTTPTransport(srv.URL, nil)
	c := NewClient(tr, "2025-11-25")

	ctx := context.Background()
	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "t1" {
		t.Fatalf("unexpected tools: %#v", tools)
	}
}

func mustJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
