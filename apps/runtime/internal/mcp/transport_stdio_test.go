package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestStdioTransport_ClientFlow(t *testing.T) {
	cmd := []string{os.Args[0], "-test.run=TestMCPHelperProcess", "--"}
	tr := NewStdioTransport(cmd, "", map[string]string{"GO_WANT_MCP_HELPER_PROCESS": "1"})
	c := NewClient(tr, "2025-11-25")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "t1" {
		t.Fatalf("unexpected tools: %#v", tools)
	}

	res, err := c.CallTool(ctx, "t1", map[string]interface{}{"x": 1})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if len(res.Content) != 1 || res.Content[0].Text != "ok" {
		t.Fatalf("unexpected result: %#v", res)
	}
}

func TestMCPHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_MCP_HELPER_PROCESS") != "1" {
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		req := RPCRequest{}
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}

		if req.ID == nil {
			continue
		}

		var result interface{}
		switch req.Method {
		case "initialize":
			result = map[string]interface{}{"capabilities": map[string]interface{}{"tools": map[string]interface{}{}}}
		case "tools/list":
			result = map[string]interface{}{"tools": []map[string]interface{}{{"name": "t1"}}}
		case "tools/call":
			result = map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": "ok"}}, "isError": false}
		default:
			resp := RPCResponse{JSONRPC: "2.0", ID: mustJSON(req.ID), Error: &RPCError{Code: -32601, Message: "method not found"}}
			b, _ := json.Marshal(resp)
			fmt.Fprintln(os.Stdout, string(b))
			continue
		}

		resp := RPCResponse{JSONRPC: "2.0", ID: mustJSON(req.ID), Result: mustJSON(result)}
		b, _ := json.Marshal(resp)
		fmt.Fprintln(os.Stdout, string(b))
	}
	os.Exit(0)
}
