package constraints

import (
	"testing"
)

func TestResolver_Resolve_ContextWindow(t *testing.T) {
	// Determinist catalog for testing
	catalog := NewCatalog()
	catalog.RegisterExact("mimo", ModelCapabilities{
		ContextWindow: 128000,
		FallbackChain: []string{"mimo-fallback"},
	})
	catalog.RegisterExact("mimo-fallback", ModelCapabilities{
		ContextWindow: 256000,
	})
	r := NewResolver(catalog)

	// Case 1: Within limits
	req := Request{
		Model:        "mimo",
		PromptTokens: 1000,
		OutputTokens: 1000,
	}
	res := r.Resolve(req)
	if res.Action != ActionAllow {
		t.Errorf("Expected Allow, got %s: %s", res.Action, res.Reason)
	}

	// Case 2: Exceeds limit
	// Case 2: Exceeds limit
	reqExceed := Request{
		Model:        "mimo",
		PromptTokens: 100000,
		OutputTokens: 50000, // Total 150k > 128k
	}
	resExceed := r.Resolve(reqExceed)

	// Expect Fallback
	if resExceed.Action != ActionFallback {
		t.Errorf("Expected Fallback, got %s", resExceed.Action)
	}
	if resExceed.TargetModel != "mimo-fallback" {
		t.Errorf("Expected fallback model mimo-fallback, got %s", resExceed.TargetModel)
	}
}

func TestResolver_Resolve_Vision(t *testing.T) {
	// Mock catalog with no vision model
	catalog := NewCatalog()
	catalog.RegisterExact("text-only", ModelCapabilities{
		SupportsVision: false,
	})
	r := NewResolver(catalog)

	req := Request{
		Model:  "text-only",
		Images: true,
	}
	res := r.Resolve(req)
	if res.Action != ActionDeny {
		t.Errorf("Expected Deny for vision request on text model, got %s", res.Action)
	}
}

func TestResolver_Resolve_Tools(t *testing.T) {
	catalog := NewCatalog()
	catalog.RegisterExact("no-tools", ModelCapabilities{
		SupportsTools: false,
	})
	r := NewResolver(catalog)

	req := Request{
		Model: "no-tools",
		Tools: []string{"some_tool"},
	}
	res := r.Resolve(req)
	if res.Action != ActionDeny {
		t.Errorf("Expected Deny for tools request on no-tools model, got %s", res.Action)
	}
}

func TestResolver_Resolve_Cost(t *testing.T) {
	catalog := NewCatalog()
	catalog.RegisterExact("expensive-model", ModelCapabilities{
		ContextWindow: 200000,
		InputPrice1M:  10.0,
		OutputPrice1M: 30.0,
	})
	r := NewResolver(catalog)

	// Case 1: Within cost limit
	req := Request{
		Model:        "expensive-model",
		PromptTokens: 1000,
		OutputTokens: 1000,
		MaxCostUSD:   1.0, // Should be way below (0.01 + 0.03 = 0.04)
	}
	res := r.Resolve(req)
	if res.Action != ActionAllow {
		t.Errorf("Expected Allow for low cost request, got %s", res.Action)
	}
	if res.EstimatedCostUSD < 0.039 || res.EstimatedCostUSD > 0.041 {
		t.Errorf("Expected cost ~0.04, got %f", res.EstimatedCostUSD)
	}

	// Case 2: Exceeds cost limit
	reqExceed := Request{
		Model:        "expensive-model",
		PromptTokens: 1000000, // $10
		OutputTokens: 1000000, // $30
		MaxCostUSD:   20.0,    // Limit is $20, cost is $40
	}
	resExceed := r.Resolve(reqExceed)
	if resExceed.Action != ActionDeny {
		t.Errorf("Expected Deny for expensive request, got %s", resExceed.Action)
	}
	if resExceed.EstimatedCostUSD < 39.0 {
		t.Errorf("Expected cost >= 40.0, got %f", resExceed.EstimatedCostUSD)
	}
}

func TestResolver_Patterns(t *testing.T) {
	catalog := DefaultCatalog()
	r := NewResolver(catalog)

	// Case 1: OpenAI Pattern
	reqGPT := Request{Model: "openai/gpt-5-custom-variant"}
	resGPT := r.Resolve(reqGPT)
	if resGPT.Action != ActionAllow {
		t.Errorf("Expected Allow for known pattern openai/gpt-5-*, got %s", resGPT.Action)
	}

	// Case 2: DeepSeek Pattern
	reqDS := Request{Model: "deepseek/deepseek-coder-v2"}
	resDS := r.Resolve(reqDS)
	if resDS.Action != ActionAllow {
		t.Errorf("Expected Allow for known pattern deepseek/*, got %s", resDS.Action)
	}

	// Case 3: Unknown Model
	reqUnknown := Request{Model: "unknown/random-model"}
	resUnknown := r.Resolve(reqUnknown)
	// Current behavior for unknown is allow but warn
	if resUnknown.Action != ActionAllow {
		t.Errorf("Expected Allow (with warning) for unknown model, got %s", resUnknown.Action)
	}
	if resUnknown.Reason == "" {
		t.Errorf("Expected warning reason for unknown model")
	}
}
