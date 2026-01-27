package constraints

import (
	_ "embed"
	"encoding/json"
	"io"
	"regexp"
	"strings"
)

//go:embed default_models.json
var defaultModelsJSON []byte

type ModelCapabilities struct {
	ContextWindow    int     `json:"context_window"`
	MaxOutputTokens  int     `json:"max_output_tokens"`
	SupportsVision   bool    `json:"supports_vision"`
	SupportsTools    bool    `json:"supports_tools"`
	SupportsThinking bool    `json:"supports_thinking"`
	InputPrice1M     float64 `json:"input_price_1m"`  // USD per 1M tokens
	OutputPrice1M    float64 `json:"output_price_1m"` // USD per 1M tokens
	FallbackModel    string  `json:"fallback_model,omitempty"`
}

type CatalogConfig struct {
	Exact    map[string]ModelCapabilities `json:"exact"`
	Patterns map[string]ModelCapabilities `json:"patterns"`
}

type Catalog struct {
	exact    map[string]ModelCapabilities
	patterns map[string]ModelCapabilities
}

func NewCatalog() *Catalog {
	return &Catalog{
		exact:    make(map[string]ModelCapabilities),
		patterns: make(map[string]ModelCapabilities),
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
		c.RegisterPattern(pattern, caps)
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

func (c *Catalog) RegisterPattern(pattern string, caps ModelCapabilities) {
	c.patterns[pattern] = caps
}

func (c *Catalog) Get(modelID string) (ModelCapabilities, bool) {
	// 1. Exact Match
	if caps, ok := c.exact[modelID]; ok {
		return caps, true
	}

	// 2. Pattern Match
	for pattern, caps := range c.patterns {
		if matched, _ := regexp.MatchString(pattern, modelID); matched {
			return caps, true
		}
	}

	// 3. Heuristic fallback (e.g. if name contains "vision")
	if strings.Contains(modelID, "vision") {
		return ModelCapabilities{SupportsVision: true, ContextWindow: 4096}, true
	}

	return ModelCapabilities{}, false
}
