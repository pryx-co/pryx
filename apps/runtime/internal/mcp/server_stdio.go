package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type ToolProvider interface {
	ServerInfo() map[string]interface{}
	ListTools(ctx context.Context) ([]Tool, error)
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (ToolResult, error)
}

func ServeStdio(ctx context.Context, provider ToolProvider) error {
	if provider == nil {
		return errors.New("missing provider")
	}

	reader := bufio.NewScanner(os.Stdin)
	reader.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	initialized := false
	protoVersion := ""

	for reader.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}

		var envelope struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      json.RawMessage `json:"id,omitempty"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params,omitempty"`
		}
		if err := json.Unmarshal([]byte(line), &envelope); err != nil {
			continue
		}
		if envelope.Method == "" {
			continue
		}

		if len(envelope.ID) == 0 {
			if envelope.Method == "initialized" {
				initialized = true
			}
			continue
		}

		switch envelope.Method {
		case "initialize":
			var params struct {
				ProtocolVersion string                 `json:"protocolVersion"`
				Capabilities    map[string]interface{} `json:"capabilities"`
				ClientInfo      map[string]interface{} `json:"clientInfo"`
			}
			_ = json.Unmarshal(envelope.Params, &params)
			protoVersion = strings.TrimSpace(params.ProtocolVersion)
			if protoVersion == "" {
				protoVersion = "2025-11-25"
			}

			result := map[string]interface{}{
				"protocolVersion": protoVersion,
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{
						"listChanged": false,
					},
				},
				"serverInfo": provider.ServerInfo(),
			}
			_ = params.Capabilities
			_ = params.ClientInfo

			if err := writeResponse(writer, envelope.ID, result, nil); err != nil {
				return err
			}

		case "tools/list":
			tools, err := provider.ListTools(ctx)
			if err != nil {
				if err := writeResponse(writer, envelope.ID, nil, &RPCError{Code: -32000, Message: err.Error()}); err != nil {
					return err
				}
				continue
			}
			result := map[string]interface{}{
				"tools": tools,
			}
			if err := writeResponse(writer, envelope.ID, result, nil); err != nil {
				return err
			}

		case "tools/call":
			var params struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			if err := json.Unmarshal(envelope.Params, &params); err != nil {
				if err := writeResponse(writer, envelope.ID, nil, &RPCError{Code: -32602, Message: "invalid params"}); err != nil {
					return err
				}
				continue
			}
			if strings.TrimSpace(params.Name) == "" {
				if err := writeResponse(writer, envelope.ID, nil, &RPCError{Code: -32602, Message: "missing tool name"}); err != nil {
					return err
				}
				continue
			}
			if params.Arguments == nil {
				params.Arguments = map[string]interface{}{}
			}

			if !initialized {
				_ = writeResponse(writer, envelope.ID, nil, &RPCError{Code: -32000, Message: "not initialized"})
				continue
			}

			res, err := provider.CallTool(ctx, params.Name, params.Arguments)
			if err != nil {
				if err := writeResponse(writer, envelope.ID, nil, &RPCError{Code: -32000, Message: err.Error()}); err != nil {
					return err
				}
				continue
			}
			if err := writeResponse(writer, envelope.ID, res, nil); err != nil {
				return err
			}

		default:
			if err := writeResponse(writer, envelope.ID, nil, &RPCError{Code: -32601, Message: "method not found"}); err != nil {
				return err
			}
		}
	}

	if err := reader.Err(); err != nil {
		return err
	}
	return io.EOF
}

func writeResponse(w *bufio.Writer, id json.RawMessage, result interface{}, rpcErr *RPCError) error {
	resp := RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
	}
	if rpcErr != nil {
		resp.Error = rpcErr
	} else {
		b, err := json.Marshal(result)
		if err != nil {
			return err
		}
		resp.Result = b
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, string(b)); err != nil {
		return err
	}
	return w.Flush()
}
