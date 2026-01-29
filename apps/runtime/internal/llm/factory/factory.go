package factory

import (
	"fmt"
	"os"

	"pryx-core/internal/llm"
	"pryx-core/internal/llm/providers"
)

type ProviderType string

const (
	ProviderOpenAI     ProviderType = "openai"
	ProviderAnthropic  ProviderType = "anthropic"
	ProviderOpenRouter ProviderType = "openrouter"
	ProviderOllama     ProviderType = "ollama"
)

func NewProvider(pt ProviderType, apiKey string, baseURL string) (llm.Provider, error) {
	// Fallback to Env Vars for API Key if empty
	if apiKey == "" {
		switch pt {
		case ProviderOpenAI:
			apiKey = os.Getenv("OPENAI_API_KEY")
		case ProviderAnthropic:
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		case ProviderOpenRouter:
			apiKey = os.Getenv("OPENROUTER_API_KEY")
			// Ollama typically doesn't need a key, but one can be provided
		}
	}

	// Validate Key (Except Ollama)
	if apiKey == "" && pt != ProviderOllama {
		return nil, fmt.Errorf("api key required for provider %s", pt)
	}

	switch pt {
	case ProviderOpenAI:
		// Check for custom base URL in Env if not provided
		if baseURL == "" {
			baseURL = os.Getenv("OPENAI_BASE_URL")
		}
		return providers.NewOpenAI(apiKey, baseURL), nil

	case ProviderOpenRouter:
		// OpenRouter is just OpenAI with a different Base URL
		return providers.NewOpenAI(apiKey, "https://openrouter.ai/api/v1"), nil

	case ProviderOllama:
		// Default Ollama URL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		// Ensure /v1 suffix for OpenAI compatibility
		if !containsV1(baseURL) {
			baseURL = fmt.Sprintf("%s/v1", baseURL)
		}
		return providers.NewOpenAI(apiKey, baseURL), nil

	case ProviderAnthropic:
		return providers.NewAnthropic(apiKey), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", pt)
	}
}

func containsV1(url string) bool {
	// Simple check, can be robustified
	len := len(url)
	if len >= 3 && url[len-3:] == "/v1" {
		return true
	}
	return false
}
