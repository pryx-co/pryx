package skills

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AgentBridge struct {
	client    *http.Client
	skill     Skill
	baseURL   string
	authToken string
	authType  string
}

type BridgeRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    interface{}
	Query   map[string]string
}

type BridgeResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       interface{}
	RawBody    []byte
}

func NewAgentBridge(skill Skill, baseURL, authToken, authType string) *AgentBridge {
	return &AgentBridge{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		skill:     skill,
		baseURL:   baseURL,
		authToken: authToken,
		authType:  authType,
	}
}

func (b *AgentBridge) Execute(ctx context.Context, req BridgeRequest) (*BridgeResponse, error) {
	url := b.baseURL + req.Path

	if len(req.Query) > 0 {
		query := ""
		for k, v := range req.Query {
			if query != "" {
				query += "&"
			}
			query += fmt.Sprintf("%s=%s", k, v)
		}
		url += "?" + query
	}

	var bodyReader io.Reader
	if req.Body != nil {
		jsonData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if b.authToken != "" {
		switch b.authType {
		case "bearer":
			httpReq.Header.Set("Authorization", "Bearer "+b.authToken)
		case "api_key":
			httpReq.Header.Set("X-API-Key", b.authToken)
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Pryx-Agent-Bridge/1.0")

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var parsedBody interface{}
	if len(respBody) > 0 {
		_ = json.Unmarshal(respBody, &parsedBody)
	}

	return &BridgeResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       parsedBody,
		RawBody:    respBody,
	}, nil
}

func (b *AgentBridge) Get(path string, query map[string]string) (*BridgeResponse, error) {
	return b.Execute(context.Background(), BridgeRequest{
		Method: "GET",
		Path:   path,
		Query:  query,
	})
}

func (b *AgentBridge) Post(path string, body interface{}) (*BridgeResponse, error) {
	return b.Execute(context.Background(), BridgeRequest{
		Method: "POST",
		Path:   path,
		Body:   body,
	})
}

func (b *AgentBridge) IsHealthy(ctx context.Context) bool {
	resp, err := b.Execute(ctx, BridgeRequest{
		Method: "GET",
		Path:   "/health",
	})
	if err != nil {
		return false
	}
	return resp.StatusCode == 200
}
