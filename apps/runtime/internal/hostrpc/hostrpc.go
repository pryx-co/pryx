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

// JSON-RPC 2.0 types
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

// Client for sending requests to host
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

// Server for receiving requests from host (sidecar mode)
type Server struct {
	in     io.Reader
	out    io.Writer
	mu     sync.Mutex
	nextID int64
}

func NewServer(in io.Reader, out io.Writer) *Server {
	return &Server{
		in:     in,
		out:    out,
		nextID: 1,
	}
}

func NewDefaultServer() *Server {
	return NewServer(os.Stdin, os.Stdout)
}

// Handler function type for RPC methods
type Handler func(method string, params map[string]interface{}) (interface{}, error)

// Registry of RPC method handlers
type Registry struct {
	handlers map[string]Handler
	mu       sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]Handler),
	}
}

func (r *Registry) Register(method string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[method] = handler
}

func (r *Registry) GetHandler(method string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[method]
	return h, ok
}

// Start the RPC server loop
func (s *Server) Serve(registry *Registry) error {
	scanner := bufio.NewScanner(s.in)
	for scanner.Scan() {
		line := scanner.Bytes()

		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendResponse(req.ID, nil, &rpcError{
				Code:    -32700,
				Message: "Parse error",
			})
			continue
		}

		if req.JSONRPC != "2.0" {
			s.sendResponse(req.ID, nil, &rpcError{
				Code:    -32600,
				Message: "Invalid Request",
			})
			continue
		}

		handler, ok := registry.GetHandler(req.Method)
		if !ok {
			s.sendResponse(req.ID, nil, &rpcError{
				Code:    -32601,
				Message: "Method not found",
			})
			continue
		}

		result, err := handler(req.Method, req.Params)
		if err != nil {
			s.sendResponse(req.ID, nil, &rpcError{
				Code:    -32603,
				Message: err.Error(),
			})
			continue
		}

		s.sendResponse(req.ID, result, nil)
	}

	return scanner.Err()
}

func (s *Server) sendResponse(id int64, result interface{}, err *rpcError) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var resultBytes json.RawMessage
	if result != nil {
		var errMarshal error
		resultBytes, errMarshal = json.Marshal(result)
		if errMarshal != nil {
			resultBytes = json.RawMessage(`{"error": "marshal error"}`)
		}
	}

	resp := rpcResponse{
		JSONRPC: "2.0",
		Result:  resultBytes,
		Error:   err,
		ID:      id,
	}

	b, _ := json.Marshal(resp)
	fmt.Fprintf(s.out, "%s\n", b)
}
