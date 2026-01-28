package constraints

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type openRouterModelsResponse struct {
	Data []openRouterModel `json:"data"`
}

type openRouterModel struct {
	ID            string `json:"id"`
	CanonicalSlug string `json:"canonical_slug"`
	Name          string `json:"name"`
	ContextLength int    `json:"context_length"`

	Architecture struct {
		InputModalities  []string `json:"input_modalities"`
		OutputModalities []string `json:"output_modalities"`
		Tokenizer        string   `json:"tokenizer"`
	} `json:"architecture"`

	Pricing struct {
		Prompt            string `json:"prompt"`
		Completion        string `json:"completion"`
		Request           string `json:"request"`
		InputCacheRead    string `json:"input_cache_read"`
		InputCacheWrite   string `json:"input_cache_write"`
		InternalReasoning string `json:"internal_reasoning"`
	} `json:"pricing"`

	TopProvider struct {
		ContextLength       int `json:"context_length"`
		MaxCompletionTokens int `json:"max_completion_tokens"`
	} `json:"top_provider"`

	SupportedParameters []string `json:"supported_parameters"`
}

func LoadDynamicCatalog(ctx context.Context) (*Catalog, error) {
	c := DefaultCatalog()
	key := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	if key == "" {
		return c, nil
	}

	orCat, err := LoadOpenRouterCatalogCached(ctx, key, 24*time.Hour)
	if err != nil {
		return c, err
	}
	c.Merge(orCat)
	return c, nil
}

func LoadOpenRouterCatalogCached(ctx context.Context, apiKey string, ttl time.Duration) (*Catalog, error) {
	cacheFile, err := openRouterCachePath()
	if err == nil && ttl > 0 {
		if st, statErr := os.Stat(cacheFile); statErr == nil {
			if time.Since(st.ModTime()) < ttl {
				if b, readErr := os.ReadFile(cacheFile); readErr == nil {
					return openRouterCatalogFromBytes(b)
				}
			}
		}
	}

	b, err := fetchOpenRouterModels(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	if cacheFile != "" {
		_ = os.MkdirAll(filepath.Dir(cacheFile), 0o755)
		_ = os.WriteFile(cacheFile, b, 0o600)
	}

	return openRouterCatalogFromBytes(b)
}

func fetchOpenRouterModels(ctx context.Context, apiKey string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, errors.New(strings.TrimSpace(string(b)))
	}
	return io.ReadAll(resp.Body)
}

func openRouterCatalogFromBytes(b []byte) (*Catalog, error) {
	var parsed openRouterModelsResponse
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}

	c := NewCatalog()
	for _, m := range parsed.Data {
		if strings.TrimSpace(m.ID) == "" {
			continue
		}
		c.RegisterExact(m.ID, openRouterToCaps(m))
	}
	return c, nil
}

func openRouterToCaps(m openRouterModel) ModelCapabilities {
	ctxLen := m.ContextLength
	if m.TopProvider.ContextLength != 0 {
		ctxLen = m.TopProvider.ContextLength
	}

	maxOut := m.TopProvider.MaxCompletionTokens

	supportsTools := containsStr(m.SupportedParameters, "tools")
	supportsThinking := containsStr(m.SupportedParameters, "reasoning") || containsStr(m.SupportedParameters, "include_reasoning")
	supportsVision := containsStr(m.Architecture.InputModalities, "image") || containsStr(m.Architecture.InputModalities, "file")
	supportsStreaming := containsStr(m.SupportedParameters, "stream")

	promptPerToken := parseFloat(m.Pricing.Prompt)
	completionPerToken := parseFloat(m.Pricing.Completion)
	requestFixed := parseFloat(m.Pricing.Request)

	supportsCaching := strings.TrimSpace(m.Pricing.InputCacheRead) != "" && strings.TrimSpace(m.Pricing.InputCacheRead) != "0"
	supportsCaching = supportsCaching || (strings.TrimSpace(m.Pricing.InputCacheWrite) != "" && strings.TrimSpace(m.Pricing.InputCacheWrite) != "0")

	caps := ModelCapabilities{
		ContextWindow:       ctxLen,
		MaxOutputTokens:     maxOut,
		SupportsTools:       supportsTools,
		SupportsThinking:    supportsThinking,
		SupportsVision:      supportsVision,
		SupportsStreaming:   supportsStreaming,
		InputPrice1M:        promptPerToken * 1_000_000.0,
		OutputPrice1M:       completionPerToken * 1_000_000.0,
		RequestFixedCostUSD: requestFixed,
		SupportsCaching:     supportsCaching,
		ProviderOverrides: map[string]ProviderOverride{
			"openrouter": {
				ContextWindow:   m.TopProvider.ContextLength,
				MaxOutputTokens: m.TopProvider.MaxCompletionTokens,
			},
		},
	}
	return caps
}

func openRouterCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pryx", "cache", "openrouter_models.json"), nil
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func containsStr(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
