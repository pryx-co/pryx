package hostrpc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

type PermissionRequest struct {
	Description string `json:"description"`
	Intent      string `json:"intent,omitempty"`
}

type PermissionResult struct {
	Approved bool `json:"approved"`
}

type rpcRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      int64                  `json:"id"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
	ID      int64           `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Client struct {
	in     *bufio.Reader
	out    io.Writer
	mu     sync.Mutex
	nextID atomic.Int64
}

func NewClient(in io.Reader, out io.Writer) *Client {
	c := &Client{
		in:  bufio.NewReader(in),
		out: out,
	}
	c.nextID.Store(1)
	return c
}

func NewDefaultClient() *Client {
	return NewClient(os.Stdin, os.Stdout)
}

func (c *Client) RequestPermission(req PermissionRequest) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.nextID.Add(1)
	params := map[string]interface{}{
		"description": req.Description,
	}
	if req.Intent != "" {
		params["intent"] = req.Intent
	}
	r := rpcRequest{
		JSONRPC: "2.0",
		Method:  "permission.request",
		Params:  params,
		ID:      id,
	}
	b, err := json.Marshal(r)
	if err != nil {
		return false, err
	}
	if _, err := fmt.Fprintf(c.out, "%s\n", b); err != nil {
		return false, err
	}

	line, err := c.in.ReadBytes('\n')
	if err != nil {
		return false, err
	}
	var resp rpcResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return false, err
	}
	if resp.ID != id {
		return false, errors.New("mismatched rpc response id")
	}
	if resp.Error != nil {
		return false, errors.New(resp.Error.Message)
	}
	var pr PermissionResult
	if err := json.Unmarshal(resp.Result, &pr); err != nil {
		return false, err
	}
	return pr.Approved, nil
}
