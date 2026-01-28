package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"pryx-core/internal/llm"
)

type OpenAIProvider struct {
	apiKey  string
	baseURL string
}

func NewOpenAI(apiKey string, baseURL string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	// Normalize base URL (remove trailing slash)
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

func (p *OpenAIProvider) Complete(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	req.Stream = false
	respBody, err := p.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer respBody.Close()

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
				Role    string `json:"role"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage llm.Usage `json:"usage"`
	}

	if err := json.NewDecoder(respBody).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding error: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	choice := apiResp.Choices[0]
	return &llm.ChatResponse{
		Content:      choice.Message.Content,
		Role:         llm.Role(choice.Message.Role),
		FinishReason: choice.FinishReason,
		Usage:        apiResp.Usage,
	}, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamChunk, error) {
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
			if string(data) == "[DONE]" {
				return
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
			}

			if err := json.Unmarshal(data, &chunk); err != nil {
				continue // skip bad chunks
			}

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta.Content
				if delta != "" {
					ch <- llm.StreamChunk{Content: delta}
				}
				if chunk.Choices[0].FinishReason != "" {
					ch <- llm.StreamChunk{Done: true}
					return
				}
			}
		}
	}()

	return ch, nil
}

func (p *OpenAIProvider) sendRequest(ctx context.Context, req llm.ChatRequest) (io.ReadCloser, error) { // Updated to use standard io.ReadCloser
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/chat/completions", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Add OpenRouter specific headers if needed, generally HTTP-Referer and X-Title are polite
	httpReq.Header.Set("HTTP-Referer", "https://pryx.app") // TODO: Make configurable
	httpReq.Header.Set("X-Title", "Pryx")

	client := &http.Client{} // TODO: Shared client
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
