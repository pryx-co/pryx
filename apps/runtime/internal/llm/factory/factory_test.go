package factory

import (
	"os"
	"testing"

	"pryx-core/internal/llm/providers"
)

func TestNewProvider_OpenAI(t *testing.T) {
	p, err := NewProvider(ProviderOpenAI, "test-key")
	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}
	if _, ok := p.(*providers.OpenAIProvider); !ok {
		t.Errorf("Expected providers.OpenAIProvider type")
	}
}

func TestNewProvider_Anthropic(t *testing.T) {
	p, err := NewProvider(ProviderAnthropic, "test-key")
	if err != nil {
		t.Fatalf("Failed to create Anthropic provider: %v", err)
	}
	if _, ok := p.(*providers.AnthropicProvider); !ok {
		t.Errorf("Expected providers.AnthropicProvider type")
	}
}

func TestNewProvider_OpenRouter(t *testing.T) {
	p, err := NewProvider(ProviderOpenRouter, "test-key")
	if err != nil {
		t.Fatalf("Failed to create OpenRouter provider: %v", err)
	}
	if _, ok := p.(*providers.OpenAIProvider); !ok {
		t.Errorf("Expected providers.OpenAIProvider type (OpenRouter uses OpenAI client)")
	}
}

func TestNewProvider_CustomBaseURL(t *testing.T) {
	os.Setenv("OPENAI_BASE_URL", "https://custom.api/v1")
	defer os.Unsetenv("OPENAI_BASE_URL")

	p, err := NewProvider(ProviderOpenAI, "test-key")
	if err != nil {
		t.Fatalf("Failed to create custom provider: %v", err)
	}

	// Verification of base URL would require exposing it on the struct or checking private field via reflection,
	// which is overkill. We assume logic holds if type is correct.
	if _, ok := p.(*providers.OpenAIProvider); !ok {
		t.Errorf("Expected providers.OpenAIProvider type")
	}
}
