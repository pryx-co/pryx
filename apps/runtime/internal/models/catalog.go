package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	ModelsDevAPI  = "https://models.dev/api.json"
	CacheTTL      = 24 * time.Hour
	CacheFileName = "models.json"
)

type ModelInfo struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Provider         string `json:"-"`
	Family           string `json:"family"`
	Attachment       bool   `json:"attachment"`
	Reasoning        bool   `json:"reasoning"`
	ToolCall         bool   `json:"tool_call"`
	StructuredOutput bool   `json:"structured_output"`
	Temperature      bool   `json:"temperature"`
	Limit            struct {
		Context int `json:"context"`
		Input   int `json:"input"`
		Output  int `json:"output"`
	} `json:"limit"`
	Cost struct {
		Input       float64 `json:"input"`
		Output      float64 `json:"output"`
		Reasoning   float64 `json:"reasoning,omitempty"`
		CacheRead   float64 `json:"cache_read,omitempty"`
		CacheWrite  float64 `json:"cache_write,omitempty"`
		InputAudio  float64 `json:"input_audio,omitempty"`
		OutputAudio float64 `json:"output_audio,omitempty"`
	} `json:"cost"`
	Modalities struct {
		Input  []string `json:"input"`
		Output []string `json:"output"`
	} `json:"modalities"`
	Knowledge   string `json:"knowledge"`
	ReleaseDate string `json:"release_date"`
	LastUpdated string `json:"last_updated"`
	OpenWeights bool   `json:"open_weights"`
}

type ProviderInfo struct {
	Name string   `json:"name"`
	NPM  string   `json:"npm"`
	Env  []string `json:"env"`
	Doc  string   `json:"doc"`
	API  string   `json:"api,omitempty"`
}

// RawProviderData represents the structure of each provider in the models.dev API
// The API returns providers as top-level keys, each containing provider info and models
type RawProviderData struct {
	ID     string               `json:"id"`
	Name   string               `json:"name"`
	NPM    string               `json:"npm"`
	Env    []string             `json:"env"`
	Doc    string               `json:"doc"`
	API    string               `json:"api,omitempty"`
	Models map[string]ModelInfo `json:"models"`
}

// Catalog represents the processed catalog with flattened models and providers
type Catalog struct {
	Models    map[string]ModelInfo    `json:"models"`
	Providers map[string]ProviderInfo `json:"providers"`
	FetchedAt time.Time               `json:"fetched_at"`
	CachedAt  time.Time               `json:"cached_at"`
}

func (c *Catalog) IsStale() bool {
	return time.Since(c.CachedAt) > CacheTTL
}

func (c *Catalog) GetProviderModels(providerID string) []ModelInfo {
	var models []ModelInfo
	for _, model := range c.Models {
		if model.Provider == providerID {
			models = append(models, model)
		}
	}
	return models
}

func (c *Catalog) GetModel(modelID string) (ModelInfo, bool) {
	model, ok := c.Models[modelID]
	return model, ok
}

func (c *Catalog) GetProvider(providerID string) (ProviderInfo, bool) {
	provider, ok := c.Providers[providerID]
	return provider, ok
}

func (m ModelInfo) SupportsTools() bool {
	return m.ToolCall
}

func (m ModelInfo) SupportsVision() bool {
	for _, mod := range m.Modalities.Input {
		if mod == "image" {
			return true
		}
	}
	return false
}

func (m ModelInfo) CalculateCost(inputTokens, outputTokens int) float64 {
	inputCost := (float64(inputTokens) / 1_000_000) * m.Cost.Input
	outputCost := (float64(outputTokens) / 1_000_000) * m.Cost.Output
	return inputCost + outputCost
}

type Service struct {
	cachePath string
	catalog   *Catalog
}

func NewService() *Service {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".pryx", "cache")
	os.MkdirAll(cacheDir, 0755)

	return &Service{
		cachePath: filepath.Join(cacheDir, CacheFileName),
	}
}

func (s *Service) Load() (*Catalog, error) {
	if catalog, err := s.loadFromCache(); err == nil && !catalog.IsStale() {
		s.catalog = catalog
		return catalog, nil
	}

	catalog, err := s.fetchFromAPI()
	if err != nil {
		if cached, cacheErr := s.loadFromCache(); cacheErr == nil {
			s.catalog = cached
			return cached, nil
		}
		return nil, fmt.Errorf("failed to fetch catalog and no cache available: %w", err)
	}

	if err := s.saveToCache(catalog); err != nil {
		fmt.Printf("Warning: failed to cache catalog: %v\n", err)
	}

	s.catalog = catalog
	return catalog, nil
}

func (s *Service) Refresh() (*Catalog, error) {
	catalog, err := s.fetchFromAPI()
	if err != nil {
		return nil, err
	}

	if err := s.saveToCache(catalog); err != nil {
		return nil, fmt.Errorf("failed to cache catalog: %w", err)
	}

	s.catalog = catalog
	return catalog, nil
}

func (s *Service) GetCatalog() *Catalog {
	return s.catalog
}

// GetPricingData returns pricing information for all models
func (s *Service) GetPricingData() map[string]interface{} {
	if s.catalog == nil {
		return nil
	}

	result := make(map[string]interface{})
	for modelID, model := range s.catalog.Models {
		result[modelID] = map[string]interface{}{
			"input_price_1m":  model.Cost.Input,
			"output_price_1m": model.Cost.Output,
			"context_window":  model.Limit.Context,
			"max_output":      model.Limit.Output,
		}
	}
	return result
}

// GetModelPricing returns pricing for a specific model
func (s *Service) GetModelPricing(modelID string) (map[string]interface{}, bool) {
	if s.catalog == nil {
		return nil, false
	}

	model, ok := s.catalog.GetModel(modelID)
	if !ok {
		return nil, false
	}

	return map[string]interface{}{
		"input_price_1m":  model.Cost.Input,
		"output_price_1m": model.Cost.Output,
		"context_window":  model.Limit.Context,
		"max_output":      model.Limit.Output,
	}, true
}

func (s *Service) fetchFromAPI() (*Catalog, error) {
	resp, err := http.Get(ModelsDevAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the raw API response where providers are top-level keys
	var rawData map[string]RawProviderData
	if err := json.Unmarshal(body, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Transform raw data into Catalog structure
	catalog := &Catalog{
		Models:    make(map[string]ModelInfo),
		Providers: make(map[string]ProviderInfo),
		FetchedAt: time.Now(),
		CachedAt:  time.Now(),
	}

	for providerID, providerData := range rawData {
		// Create ProviderInfo
		provider := ProviderInfo{
			Name: providerData.Name,
			NPM:  providerData.NPM,
			Env:  providerData.Env,
			Doc:  providerData.Doc,
			API:  providerData.API,
		}
		catalog.Providers[providerID] = provider

		// Add models with provider field populated
		for modelID, model := range providerData.Models {
			model.ID = modelID
			model.Provider = providerID
			catalog.Models[modelID] = model
		}
	}

	return catalog, nil
}

func (s *Service) loadFromCache() (*Catalog, error) {
	data, err := os.ReadFile(s.cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}

	return &catalog, nil
}

func (s *Service) saveToCache(catalog *Catalog) error {
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal catalog: %w", err)
	}

	if err := os.WriteFile(s.cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

var DefaultProviderMapping = map[string]string{
	"openai":     "openai",
	"anthropic":  "anthropic",
	"google":     "google",
	"ollama":     "ollama",
	"openrouter": "openrouter",
	"together":   "together",
	"mistral":    "mistral",
	"cohere":     "cohere",
	"groq":       "groq",
	"xai":        "xai",
}

func GetSupportedProviders() []string {
	providers := make([]string, 0, len(DefaultProviderMapping))
	for k := range DefaultProviderMapping {
		providers = append(providers, k)
	}
	return providers
}
