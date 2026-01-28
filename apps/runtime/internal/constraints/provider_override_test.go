package constraints

import "testing"

func TestModelCapabilities_Effective_ProviderOverride(t *testing.T) {
	caps := ModelCapabilities{
		ContextWindow:     100,
		MaxOutputTokens:   10,
		MaxThinkingTokens: 5,
		ProviderOverrides: map[string]ProviderOverride{
			"openrouter": {ContextWindow: 200, MaxOutputTokens: 20, MaxThinkingTokens: 15},
		},
	}

	e := caps.Effective("openrouter")
	if e.ContextWindow != 200 || e.MaxOutputTokens != 20 || e.MaxThinkingTokens != 15 {
		t.Fatalf("unexpected effective caps: %#v", e)
	}
}
