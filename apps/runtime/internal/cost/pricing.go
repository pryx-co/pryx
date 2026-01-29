package cost

import (
	"sync"
	"time"
)

// Pricing model constants (prices per 1K tokens)
const (
	PricePer1K = 1000.0
)

// Default pricing for popular models (USD per 1K tokens)
var defaultPricing = map[string]ModelPricing{
	// OpenAI Models
	"gpt-4o": {
		ModelID:          "gpt-4o",
		InputPricePer1K:  2.50,
		OutputPricePer1K: 10.00,
		Provider:         "openai",
		UpdatedAt:        time.Now(),
	},
	"gpt-4o-mini": {
		ModelID:          "gpt-4o-mini",
		InputPricePer1K:  0.150,
		OutputPricePer1K: 0.600,
		Provider:         "openai",
		UpdatedAt:        time.Now(),
	},
	"gpt-4-turbo": {
		ModelID:          "gpt-4-turbo",
		InputPricePer1K:  10.00,
		OutputPricePer1K: 30.00,
		Provider:         "openai",
		UpdatedAt:        time.Now(),
	},
	"gpt-4": {
		ModelID:          "gpt-4",
		InputPricePer1K:  30.00,
		OutputPricePer1K: 60.00,
		Provider:         "openai",
		UpdatedAt:        time.Now(),
	},
	"gpt-3.5-turbo": {
		ModelID:          "gpt-3.5-turbo",
		InputPricePer1K:  0.50,
		OutputPricePer1K: 1.50,
		Provider:         "openai",
		UpdatedAt:        time.Now(),
	},

	// Anthropic Models
	"claude-sonnet-4-20250514": {
		ModelID:          "claude-sonnet-4-20250514",
		InputPricePer1K:  3.00,
		OutputPricePer1K: 15.00,
		Provider:         "anthropic",
		UpdatedAt:        time.Now(),
	},
	"claude-opus-4-20250514": {
		ModelID:          "claude-opus-4-20250514",
		InputPricePer1K:  15.00,
		OutputPricePer1K: 75.00,
		Provider:         "anthropic",
		UpdatedAt:        time.Now(),
	},
	"claude-haiku-3-20250514": {
		ModelID:          "claude-haiku-3-20250514",
		InputPricePer1K:  0.25,
		OutputPricePer1K: 1.25,
		Provider:         "anthropic",
		UpdatedAt:        time.Now(),
	},

	// DeepSeek Models
	"deepseek-chat": {
		ModelID:          "deepseek-chat",
		InputPricePer1K:  0.14,
		OutputPricePer1K: 0.28,
		Provider:         "deepseek",
		UpdatedAt:        time.Now(),
	},

	// Google Models
	"gemini-1.5-pro": {
		ModelID:          "gemini-1.5-pro",
		InputPricePer1K:  1.25,
		OutputPricePer1K: 5.00,
		Provider:         "google",
		UpdatedAt:        time.Now(),
	},
	"gemini-1.5-flash": {
		ModelID:          "gemini-1.5-flash",
		InputPricePer1K:  0.075,
		OutputPricePer1K: 0.30,
		Provider:         "google",
		UpdatedAt:        time.Now(),
	},
}

// PricingManager handles model pricing lookups and updates
type PricingManager struct {
	mu      sync.RWMutex
	pricing map[string]ModelPricing
}

func NewPricingManager() *PricingManager {
	pm := &PricingManager{
		pricing: make(map[string]ModelPricing),
	}
	// Initialize with default pricing
	for id, p := range defaultPricing {
		pm.pricing[id] = p
	}
	return pm
}

// GetPricing returns pricing for a specific model
func (pm *PricingManager) GetPricing(modelID string) (ModelPricing, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pricing, exists := pm.pricing[modelID]
	return pricing, exists
}

// SetPricing updates pricing for a model
func (pm *PricingManager) SetPricing(pricing ModelPricing) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pricing.UpdatedAt = time.Now()
	pm.pricing[pricing.ModelID] = pricing
}

// GetProviderPricing returns all pricing for a specific provider
func (pm *PricingManager) GetProviderPricing(provider string) []ModelPricing {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var pricing []ModelPricing
	for _, p := range pm.pricing {
		if p.Provider == provider {
			pricing = append(pricing, p)
		}
	}
	return pricing
}

// AllPricing returns all current pricing
func (pm *PricingManager) AllPricing() []ModelPricing {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pricing := make([]ModelPricing, 0, len(pm.pricing))
	for _, p := range pm.pricing {
		pricing = append(pricing, p)
	}
	return pricing
}

// CalculateCost calculates the cost for given token usage
func (pm *PricingManager) CalculateCost(modelID string, inputTokens, outputTokens int) (float64, error) {
	pricing, exists := pm.GetPricing(modelID)
	if !exists {
		// Return 0 for unknown models instead of error
		return 0, nil
	}

	inputCost := float64(inputTokens) * pricing.InputPricePer1K / PricePer1K
	outputCost := float64(outputTokens) * pricing.OutputPricePer1K / PricePer1K

	return inputCost + outputCost, nil
}
