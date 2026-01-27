package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"pryx-core/internal/llm"
)

type AnthropicProvider struct {
	apiKey string
}

func NewAnthropic(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{apiKey: apiKey}
}

const anthropicURL = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

func (p *AnthropicProvider) Complete(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	req.Stream = false
	respBody, err := p.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer respBody.Close()

	var apiResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Role       string    `json:"role"`
		StopReason string    `json:"stop_reason"`
		Usage      llm.Usage `json:"usage"`
	}

	if err := json.NewDecoder(respBody).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding error: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("no content returned")
	}

	return &llm.ChatResponse{
		Content:      apiResp.Content[0].Text,
		Role:         llm.RoleAssistant,
		FinishReason: apiResp.StopReason,
		Usage:        apiResp.Usage,
	}, nil
}

func (p *AnthropicProvider) Stream(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	req.Stream = true
	respBody, err := p.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan llm.StreamChunk)
	go func() {
		defer close(ch)
		defer respBody.Close()

		reader := bufio.NewReader(respBody)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				ch <- llm.StreamChunk{Err: err}
				return
			}

			line = bytes.TrimSpace(line)
			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			data := bytes.TrimPrefix(line, []byte("data: "))
			if len(data) == 0 {
				continue
			}

			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}

			if err := json.Unmarshal(data, &event); err != nil {
				continue
			}

			switch event.Type {
			case "content_block_delta":
				if event.Delta.Text != "" {
					ch <- llm.StreamChunk{Content: event.Delta.Text}
				}
			case "message_stop":
				ch <- llm.StreamChunk{Done: true}
				return
			}
		}
	}()

	return ch, nil
}

func (p *AnthropicProvider) sendRequest(ctx context.Context, req llm.ChatRequest) (io.ReadCloser, error) {
	payload := map[string]interface{}{
		"model":      req.Model,
		"messages":   req.Messages,
		"max_tokens": req.MaxTokens,
		"stream":     req.Stream,
	}
	if req.MaxTokens == 0 {
		payload["max_tokens"] = 1000
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", anthropicURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf("api error: %s - %s", resp.Status, buf.String())
	}

	return resp.Body, nil
}
