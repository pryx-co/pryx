package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pryx-core/internal/llm"
)

func TestNewOpenAI(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		baseURL string
		wantURL string
	}{
		{
			name:    "default URL",
			apiKey:  "test-key",
			baseURL: "",
			wantURL: "https://api.openai.com/v1",
		},
		{
			name:    "custom URL",
			apiKey:  "test-key",
			baseURL: "https://custom.openai.com",
			wantURL: "https://custom.openai.com",
		},
		{
			name:    "URL with trailing slash",
			apiKey:  "test-key",
			baseURL: "https://custom.openai.com/",
			wantURL: "https://custom.openai.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewOpenAI(tt.apiKey, tt.baseURL)
			if p == nil {
				t.Fatal("NewOpenAI() returned nil")
			}
			if p.apiKey != tt.apiKey {
				t.Errorf("apiKey = %v, want %v", p.apiKey, tt.apiKey)
			}
			if p.baseURL != tt.wantURL {
				t.Errorf("baseURL = %v, want %v", p.baseURL, tt.wantURL)
			}
		})
	}
}

func TestOpenAIProvider_Complete(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Check content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Check authorization
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header Bearer test-api-key, got %s", r.Header.Get("Authorization"))
		}

		// Check URL path
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
		}

		// Return mock response
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "Hello! This is a test response.",
						"role":    "assistant",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]int{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewOpenAI("test-api-key", server.URL)

	req := llm.ChatRequest{
		Model: "gpt-4",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "Hello"},
		},
	}

	resp, err := provider.Complete(context.Background(), req)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if resp.Content != "Hello! This is a test response." {
		t.Errorf("Content = %v, want %v", resp.Content, "Hello! This is a test response.")
	}

	if resp.Role != llm.RoleAssistant {
		t.Errorf("Role = %v, want %v", resp.Role, llm.RoleAssistant)
	}

	if resp.FinishReason != "stop" {
		t.Errorf("FinishReason = %v, want %v", resp.FinishReason, "stop")
	}

	if resp.Usage.TotalTokens != 30 {
		t.Errorf("Usage.TotalTokens = %v, want %v", resp.Usage.TotalTokens, 30)
	}
}

func TestOpenAIProvider_Complete_Error(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid API key",
		})
	}))
	defer server.Close()

	provider := NewOpenAI("invalid-key", server.URL)

	req := llm.ChatRequest{
		Model: "gpt-4",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "Hello"},
		},
	}

	_, err := provider.Complete(context.Background(), req)
	if err == nil {
		t.Error("Complete() expected error, got nil")
	}
}

func TestOpenAIProvider_Stream(t *testing.T) {
	// Create test server with SSE streaming
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Send SSE events
		events := []string{
			`data: {"choices":[{"delta":{"content":"Hello"}}]}` + "\n\n",
			`data: {"choices":[{"delta":{"content":" world"}}]}` + "\n\n",
			`data: {"choices":[{"delta":{},"finish_reason":"stop"}]}` + "\n\n",
			"data: [DONE]\n\n",
		}

		for _, event := range events {
			w.Write([]byte(event))
			w.(http.Flusher).Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	provider := NewOpenAI("test-api-key", server.URL)

	req := llm.ChatRequest{
		Model: "gpt-4",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "Hello"},
		},
	}

	stream, err := provider.Stream(context.Background(), req)
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}

	var chunks []string
	for chunk := range stream {
		if chunk.Err != nil {
			t.Errorf("Stream chunk error: %v", chunk.Err)
			continue
		}
		chunks = append(chunks, chunk.Content)
	}

	// Filter out empty chunks
	var nonEmptyChunks []string
	for _, c := range chunks {
		if c != "" {
			nonEmptyChunks = append(nonEmptyChunks, c)
		}
	}

	if len(nonEmptyChunks) != 2 {
		t.Errorf("Expected 2 non-empty chunks, got %d (total: %d)", len(nonEmptyChunks), len(chunks))
	}

	if len(nonEmptyChunks) >= 1 && nonEmptyChunks[0] != "Hello" {
		t.Errorf("First chunk = %v, want %v", nonEmptyChunks[0], "Hello")
	}

	if len(nonEmptyChunks) >= 2 && nonEmptyChunks[1] != " world" {
		t.Errorf("Second chunk = %v, want %v", nonEmptyChunks[1], " world")
	}
}
