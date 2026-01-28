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
)

func NewProvider(pt ProviderType, apiKey string) (llm.Provider, error) {
	// Fallback to Env Vars
	if apiKey == "" {
		switch pt {
		case ProviderOpenAI:
			apiKey = os.Getenv("OPENAI_API_KEY")
		case ProviderAnthropic:
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		case ProviderOpenRouter:
			apiKey = os.Getenv("OPENROUTER_API_KEY")
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("api key required for provider %s", pt)
	}

	switch pt {
	case ProviderOpenAI:
		// Check for custom base URL
		baseURL := os.Getenv("OPENAI_BASE_URL")
		return providers.NewOpenAI(apiKey, baseURL), nil

	case ProviderOpenRouter:
		// OpenRouter is just OpenAI with a different Base URL
		return providers.NewOpenAI(apiKey, "https://openrouter.ai/api/v1"), nil

	case ProviderAnthropic:
		return providers.NewAnthropic(apiKey), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", pt)
	}
}
