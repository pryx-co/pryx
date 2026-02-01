// Package factory provides LLM provider factory functionality.
// It creates and configures LLM providers based on the models catalog and user configuration.
package factory

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"pryx-core/internal/auth"
	"pryx-core/internal/keychain"
	"pryx-core/internal/llm"
	"pryx-core/internal/llm/providers"
	"pryx-core/internal/models"
)

// ProviderFactory creates LLM provider instances based on configuration.
type ProviderFactory struct {
	catalog  *models.Catalog
	keychain *keychain.Keychain
}

// NewProviderFactory creates a new provider factory with the given catalog and keychain.
func NewProviderFactory(catalog *models.Catalog, kc *keychain.Keychain) *ProviderFactory {
	return &ProviderFactory{
		catalog:  catalog,
		keychain: kc,
	}
}

// CreateProvider creates an LLM provider for the specified provider and model.
// It resolves the API key from the provided value, keychain, or environment variables.
func (f *ProviderFactory) CreateProvider(providerID, modelID, apiKey string) (llm.Provider, error) {
	if f.catalog == nil {
		return nil, fmt.Errorf("catalog not loaded")
	}

	providerInfo, ok := f.catalog.GetProvider(providerID)
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", providerID)
	}

	_, ok = f.catalog.GetModel(modelID)
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", modelID)
	}

	apiKey = f.resolveAPIKey(providerID, apiKey, providerInfo)

	implType := f.getImplementationType(providerID, providerInfo)

	switch implType {
	case "openai", "openai-compatible":
		baseURL := f.getBaseURL(providerID, providerInfo)
		return providers.NewOpenAI(apiKey, baseURL), nil

	case "anthropic":
		return providers.NewAnthropic(apiKey), nil

	default:
		baseURL := f.getBaseURL(providerID, providerInfo)
		return providers.NewOpenAI(apiKey, baseURL), nil
	}
}

// CreateProviderFromConfig creates an LLM provider using configuration defaults for the model.
func (f *ProviderFactory) CreateProviderFromConfig(providerID, apiKey string) (llm.Provider, error) {
	if f.catalog == nil {
		return nil, fmt.Errorf("catalog not loaded")
	}

	providerInfo, ok := f.catalog.GetProvider(providerID)
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", providerID)
	}

	apiKey = f.resolveAPIKey(providerID, apiKey, providerInfo)

	implType := f.getImplementationType(providerID, providerInfo)
	baseURL := f.getBaseURL(providerID, providerInfo)

	switch implType {
	case "openai", "openai-compatible":
		return providers.NewOpenAI(apiKey, baseURL), nil
	case "anthropic":
		return providers.NewAnthropic(apiKey), nil
	default:
		return providers.NewOpenAI(apiKey, baseURL), nil
	}
}

// IsProviderSupported checks if the given provider ID is supported.
func (f *ProviderFactory) IsProviderSupported(providerID string) bool {
	_, ok := models.DefaultProviderMapping[providerID]
	return ok
}

// GetSupportedProviders returns a list of all supported provider IDs.
func (f *ProviderFactory) GetSupportedProviders() []string {
	return models.GetSupportedProviders()
}

// GetProviderModels returns all models available for the specified provider.
func (f *ProviderFactory) GetProviderModels(providerID string) ([]models.ModelInfo, error) {
	if f.catalog == nil {
		return nil, fmt.Errorf("catalog not loaded")
	}
	return f.catalog.GetProviderModels(providerID), nil
}

// GetModelInfo returns information about a specific model.
func (f *ProviderFactory) GetModelInfo(modelID string) (models.ModelInfo, bool) {
	if f.catalog == nil {
		return models.ModelInfo{}, false
	}
	return f.catalog.GetModel(modelID)
}

// GetCatalog returns the models catalog used by this factory.
func (f *ProviderFactory) GetCatalog() *models.Catalog {
	return f.catalog
}

func (f *ProviderFactory) resolveAPIKey(providerID, providedKey string, providerInfo models.ProviderInfo) string {
	if providedKey != "" {
		return providedKey
	}

	// Try OAuth token first (for providers that support it)
	if f.supportsOAuth(providerID) {
		if token := f.getOAuthToken(providerID); token != "" {
			return token
		}
	}

	if f.keychain != nil {
		if key, err := f.keychain.GetProviderKey(providerID); err == nil && key != "" {
			return key
		}
	}

	return f.getAPIKeyFromEnv(providerID, providerInfo)
}

func (f *ProviderFactory) supportsOAuth(providerID string) bool {
	return providerID == "google" // Currently only Google supports OAuth
}

func (f *ProviderFactory) getOAuthToken(providerID string) string {
	if f.keychain == nil {
		return ""
	}

	token, err := f.keychain.Get("oauth_" + providerID + "_access")
	if err != nil {
		return ""
	}

	oauth := auth.NewProviderOAuth(f.keychain)
	needsRefresh, _ := oauth.IsTokenExpired(providerID)
	if needsRefresh {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		// Try to refresh the token. If refresh fails, return empty string
		// to allow fallback to API key instead of using potentially expired token.
		if err := oauth.RefreshToken(ctx, providerID); err != nil {
			return ""
		}
		// Refresh succeeded, get the new token
		token, _ = f.keychain.Get("oauth_" + providerID + "_access")
	}

	return token
}

// getAPIKeyFromEnv retrieves the API key from environment variables.
func (f *ProviderFactory) getAPIKeyFromEnv(providerID string, providerInfo models.ProviderInfo) string {
	for _, envVar := range providerInfo.Env {
		if key := os.Getenv(envVar); key != "" {
			return key
		}
	}

	switch providerID {
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "openrouter":
		return os.Getenv("OPENROUTER_API_KEY")
	case "together":
		return os.Getenv("TOGETHER_API_KEY")
	case "groq":
		return os.Getenv("GROQ_API_KEY")
	case "xai":
		return os.Getenv("XAI_API_KEY")
	case "mistral":
		return os.Getenv("MISTRAL_API_KEY")
	case "cohere":
		return os.Getenv("COHERE_API_KEY")
	case "google":
		return os.Getenv("GOOGLE_API_KEY")
	}

	return ""
}

// getImplementationType determines the implementation type for a provider.
func (f *ProviderFactory) getImplementationType(providerID string, providerInfo models.ProviderInfo) string {
	npm := providerInfo.NPM

	if strings.Contains(npm, "anthropic") {
		return "anthropic"
	}

	if strings.Contains(npm, "openai") {
		return "openai"
	}

	return "openai-compatible"
}

// getBaseURL returns the base URL for the specified provider.
func (f *ProviderFactory) getBaseURL(providerID string, providerInfo models.ProviderInfo) string {
	if providerInfo.API != "" {
		return providerInfo.API
	}

	switch providerID {
	case "openai":
		if url := os.Getenv("OPENAI_BASE_URL"); url != "" {
			return url
		}
		return "https://api.openai.com/v1"

	case "openrouter":
		return "https://openrouter.ai/api/v1"

	case "together":
		return "https://api.together.xyz/v1"

	case "groq":
		return "https://api.groq.com/openai/v1"

	case "mistral":
		return "https://api.mistral.ai/v1"

	case "ollama":
		baseURL := "http://localhost:11434"
		if url := os.Getenv("OLLAMA_HOST"); url != "" {
			baseURL = url
		}
		return ensureV1Suffix(baseURL)

	case "glm":
		return "https://open.bigmodel.cn/api/paas/v4"
	}

	return ""
}

// ensureV1Suffix ensures the URL ends with "/v1" for OpenAI-compatible endpoints.
func ensureV1Suffix(url string) string {
	if !strings.HasSuffix(url, "/v1") {
		return url + "/v1"
	}
	return url
}

// Provider constants for supported LLM providers.
const (
	// ProviderOpenAI is the OpenAI provider.
	ProviderOpenAI = "openai"
	// ProviderAnthropic is the Anthropic provider.
	ProviderAnthropic = "anthropic"
	// ProviderOpenRouter is the OpenRouter provider aggregator.
	ProviderOpenRouter = "openrouter"
	// ProviderOllama is the local Ollama provider.
	ProviderOllama = "ollama"
	// ProviderGLM is the GLM (Zhipu AI) provider.
	ProviderGLM = "glm"
)

// NewProvider creates a new LLM provider instance based on the provider type.
// This is a lower-level function that creates providers without using the catalog.
func NewProvider(pt string, apiKey string, baseURL string) (llm.Provider, error) {
	switch pt {
	case ProviderOpenAI:
		if baseURL == "" {
			baseURL = os.Getenv("OPENAI_BASE_URL")
			if baseURL == "" {
				baseURL = "https://api.openai.com/v1"
			}
		}
		return providers.NewOpenAI(apiKey, baseURL), nil

	case ProviderAnthropic:
		return providers.NewAnthropic(apiKey), nil

	case ProviderOpenRouter:
		return providers.NewOpenAI(apiKey, "https://openrouter.ai/api/v1"), nil

	case ProviderOllama:
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return providers.NewOpenAI(apiKey, ensureV1Suffix(baseURL)), nil

	case ProviderGLM:
		return providers.NewOpenAI(apiKey, "https://open.bigmodel.cn/api/paas/v4"), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", pt)
	}
}
