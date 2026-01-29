package constraints

import (
	_ "embed"
	"encoding/json"
	"io"
	"regexp"
	"strings"

	"pryx-core/internal/models"
)

//go:embed default_models.json
var defaultModelsJSON []byte

type ModelCapabilities struct {
	ContextWindow        int  `json:"context_window"`
	MaxOutputTokens      int  `json:"max_output_tokens"`
	SupportsVision       bool `json:"supports_vision"`
	SupportsTools        bool `json:"supports_tools"`
	SupportsThinking     bool `json:"supports_thinking"`
	SupportsStreaming    bool `json:"supports_streaming,omitempty"`
	MaxToolsPerRequest   int  `json:"max_tools_per_request,omitempty"`
	MaxToolCallsParallel int  `json:"max_tool_calls_parallel,omitempty"`
	MaxImagesPerRequest  int  `json:"max_images_per_request,omitempty"`
	MaxThinkingTokens    int  `json:"max_thinking_tokens,omitempty"`

	InputPrice1M        float64 `json:"input_price_1m"`  // USD per 1M tokens
	OutputPrice1M       float64 `json:"output_price_1m"` // USD per 1M tokens
	RequestFixedCostUSD float64 `json:"request_fixed_cost_usd,omitempty"`
	SupportsCaching     bool    `json:"supports_caching,omitempty"`

	ProviderOverrides map[string]ProviderOverride `json:"provider_overrides,omitempty"`

	FallbackChain []string `json:"fallback_chain,omitempty"` // Primary → secondary → tertiary
}

// FallbackModel provides backward compatibility with single fallback
func (m ModelCapabilities) FallbackModel() string {
	if len(m.FallbackChain) > 0 {
		return m.FallbackChain[0]
	}
	return ""
}

type ProviderOverride struct {
	ContextWindow     int `json:"context_window,omitempty"`
	MaxOutputTokens   int `json:"max_output_tokens,omitempty"`
	MaxThinkingTokens int `json:"max_thinking_tokens,omitempty"`
}

func (c ModelCapabilities) Effective(providerID string) ModelCapabilities {
	if providerID == "" || len(c.ProviderOverrides) == 0 {
		return c
	}
	ov, ok := c.ProviderOverrides[providerID]
	if !ok {
		return c
	}
	if ov.ContextWindow != 0 {
		c.ContextWindow = ov.ContextWindow
	}
	if ov.MaxOutputTokens != 0 {
		c.MaxOutputTokens = ov.MaxOutputTokens
	}
	if ov.MaxThinkingTokens != 0 {
		c.MaxThinkingTokens = ov.MaxThinkingTokens
	}
	return c
}

func (c ModelCapabilities) EstimateCost(req Request) CostEstimate {
	inputCost := float64(req.PromptTokens) / 1000000.0 * c.InputPrice1M
	outputCost := float64(req.OutputTokens) / 1000000.0 * c.OutputPrice1M
	thinkingCost := float64(req.ThinkingTokens) / 1000000.0 * c.OutputPrice1M

	return CostEstimate{
		InputTokensCost:    inputCost,
		OutputTokensCost:   outputCost,
		ThinkingTokensCost: thinkingCost,
		FixedCost:          c.RequestFixedCostUSD,
		TotalUSD:           inputCost + outputCost + thinkingCost + c.RequestFixedCostUSD,
	}
}

type CatalogConfig struct {
	Exact    map[string]ModelCapabilities `json:"exact"`
	Patterns map[string]ModelCapabilities `json:"patterns"`
}

type Catalog struct {
	exact    map[string]ModelCapabilities
	patterns []patternEntry
}

type patternEntry struct {
	raw  string
	re   *regexp.Regexp
	caps ModelCapabilities
}

func NewCatalog() *Catalog {
	return &Catalog{
		exact:    make(map[string]ModelCapabilities),
		patterns: []patternEntry{},
	}
}

func DefaultCatalog() *Catalog {
	c := NewCatalog()
	// Ignore error for embedded default, should be valid at build time
	if err := c.LoadFromBytes(defaultModelsJSON); err != nil {
		panic("failed to load default models: " + err.Error())
	}
	return c
}

func (c *Catalog) LoadFromBytes(data []byte) error {
	var config CatalogConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	for id, caps := range config.Exact {
		c.RegisterExact(id, caps)
	}

	for pattern, caps := range config.Patterns {
		if err := c.RegisterPattern(pattern, caps); err != nil {
			return err
		}
	}

	return nil
}

func (c *Catalog) LoadFromReader(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return c.LoadFromBytes(data)
}

func (c *Catalog) RegisterExact(id string, caps ModelCapabilities) {
	c.exact[id] = caps
}

func (c *Catalog) RegisterPattern(pattern string, caps ModelCapabilities) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	c.patterns = append(c.patterns, patternEntry{
		raw:  pattern,
		re:   re,
		caps: caps,
	})
	return nil
}

func (c *Catalog) AllModelIDs() []string {
	var out []string
	for id := range c.exact {
		out = append(out, id)
	}
	return out
}

func (c *Catalog) Merge(other *Catalog) {
	if other == nil {
		return
	}
	for id, caps := range other.exact {
		c.exact[id] = caps
	}
	c.patterns = append(c.patterns, other.patterns...)
}

func (c *Catalog) Get(modelID string) (ModelCapabilities, bool) {
	// 1. Exact Match
	if caps, ok := c.exact[modelID]; ok {
		return caps, true
	}

	// 2. Pattern Match
	for _, entry := range c.patterns {
		if entry.re.MatchString(modelID) {
			return entry.caps, true
		}
	}

	// 3. Heuristic fallback (e.g. if name contains "vision")
	if strings.Contains(modelID, "vision") {
		return ModelCapabilities{SupportsVision: true, ContextWindow: 4096}, true
	}

	return ModelCapabilities{}, false
}

func FromModelsDevCatalog(catalog *models.Catalog) *Catalog {
	c := NewCatalog()
	if catalog == nil {
		return c
	}

	for modelID, modelInfo := range catalog.Models {
		caps := ModelCapabilities{
			ContextWindow:       modelInfo.Limit.Context,
			MaxOutputTokens:     modelInfo.Limit.Output,
			SupportsVision:      modelInfo.SupportsVision(),
			SupportsTools:       modelInfo.SupportsTools(),
			SupportsThinking:    modelInfo.Reasoning,
			InputPrice1M:        modelInfo.Cost.Input,
			OutputPrice1M:       modelInfo.Cost.Output,
			RequestFixedCostUSD: 0,
		}
		c.RegisterExact(modelID, caps)
	}

	return c
}
