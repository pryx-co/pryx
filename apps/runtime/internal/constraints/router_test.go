package constraints

import "testing"

func TestRouter_Select_LowestCost(t *testing.T) {
	c := NewCatalog()
	c.RegisterExact("a", ModelCapabilities{
		ContextWindow:   1000,
		MaxOutputTokens: 1000,
		SupportsTools:   true,
		InputPrice1M:    1.0,
		OutputPrice1M:   1.0,
	})
	c.RegisterExact("b", ModelCapabilities{
		ContextWindow:   1000,
		MaxOutputTokens: 1000,
		SupportsTools:   true,
		InputPrice1M:    0.1,
		OutputPrice1M:   0.1,
	})

	r := NewRouter(c)
	id, cost, res := r.Select(RouteRequest{
		PromptTokens:  100,
		OutputTokens:  100,
		RequiresTools: true,
		Candidates:    []string{"a", "b"},
	})

	if res.Action != ActionAllow {
		t.Fatalf("expected allow, got %s (%s)", res.Action, res.Reason)
	}
	if id != "b" {
		t.Fatalf("expected b, got %s", id)
	}
	if cost <= 0 {
		t.Fatalf("expected positive cost, got %f", cost)
	}
}

func TestRouter_Select_FallbackChain(t *testing.T) {
	c := NewCatalog()
	c.RegisterExact("primary", ModelCapabilities{
		ContextWindow:   50,
		MaxOutputTokens: 50,
		SupportsTools:   true,
	})
	c.RegisterExact("fallback", ModelCapabilities{
		ContextWindow:   1000,
		MaxOutputTokens: 1000,
		SupportsTools:   true,
	})

	r := NewRouter(c)
	id, _, res := r.Select(RouteRequest{
		PromptTokens:  900,
		OutputTokens:  50,
		RequiresTools: true,
		Candidates:    []string{"primary"},
		FallbackChain: []string{"fallback"},
	})

	if res.Action != ActionFallback {
		t.Fatalf("expected fallback, got %s (%s)", res.Action, res.Reason)
	}
	if id != "fallback" {
		t.Fatalf("expected fallback, got %s", id)
	}
}
