package llm

import "context"

// Provider defines the interface for an LLM provider
type Provider interface {
	// Complete performs a non-streaming completion
	Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// Stream performs a streaming completion, returning a channel of chunks
	Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
}
