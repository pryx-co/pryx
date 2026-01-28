package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Transport interface {
	Call(ctx context.Context, req RPCRequest) (RPCResponse, error)
	Notify(ctx context.Context, notif RPCNotification) error
	Close() error
}

type Client struct {
	transport       Transport
	protocolVersion string

	initialized atomic.Bool
	idCounter   atomic.Int64

	mu                 sync.RWMutex
	serverCapabilities json.RawMessage
}

func NewClient(transport Transport, protocolVersion string) *Client {
	if protocolVersion == "" {
		protocolVersion = "2025-11-25"
	}
	return &Client{
		transport:       transport,
		protocolVersion: protocolVersion,
	}
}

func (c *Client) Close() error {
	return c.transport.Close()
}

func (c *Client) Initialize(ctx context.Context) error {
	if c.initialized.Load() {
		return nil
	}

	initCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	params := map[string]interface{}{
		"protocolVersion": c.protocolVersion,
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"clientInfo": map[string]interface{}{
			"name":    "pryx-core",
			"version": "dev",
		},
	}

	var result struct {
		Capabilities json.RawMessage `json:"capabilities"`
	}
	if err := c.call(initCtx, "initialize", params, &result); err != nil {
		return err
	}

	c.mu.Lock()
	c.serverCapabilities = result.Capabilities
	c.mu.Unlock()

	if err := c.transport.Notify(initCtx, RPCNotification{
		JSONRPC: "2.0",
		Method:  "initialized",
	}); err != nil {
		return err
	}

	c.initialized.Store(true)
	return nil
}

func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	if err := c.Initialize(ctx); err != nil {
		return nil, err
	}

	var all []Tool
	cursor := ""

	for {
		params := map[string]interface{}{}
		if cursor != "" {
			params["cursor"] = cursor
		}
		var out ListToolsResult
		if err := c.call(ctx, "tools/list", params, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Tools...)
		if out.NextCursor == "" {
			break
		}
		cursor = out.NextCursor
	}

	return all, nil
}

func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (ToolResult, error) {
	if err := c.Initialize(ctx); err != nil {
		return ToolResult{}, err
	}
	params := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}
	var out ToolResult
	if err := c.call(ctx, "tools/call", params, &out); err != nil {
		return ToolResult{}, err
	}
	if out.IsError {
		return out, errors.New("tool returned error")
	}
	return out, nil
}

func (c *Client) call(ctx context.Context, method string, params interface{}, out interface{}) error {
	id := c.idCounter.Add(1)
	req := RPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	resp, err := c.transport.Call(ctx, req)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("mcp error %d: %s", resp.Error.Code, resp.Error.Message)
	}
	if len(resp.Result) == 0 {
		return errors.New("empty result")
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(resp.Result, out); err != nil {
		return err
	}
	return nil
}
