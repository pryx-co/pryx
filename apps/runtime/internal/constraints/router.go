package constraints

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

type RouteRequest struct {
	ProviderID     string
	PromptTokens   int
	OutputTokens   int
	ThinkingTokens int

	RequiresTools  bool
	RequiresVision bool

	Candidates    []string
	FallbackChain []string
	MaxCostUSD    float64
}

type Router struct {
	catalog *Catalog
}

func NewRouter(catalog *Catalog) *Router {
	if catalog == nil {
		catalog = DefaultCatalog()
	}
	return &Router{catalog: catalog}
}

func (r *Router) Select(req RouteRequest) (string, float64, Resolution) {
	candidates := req.Candidates
	if len(candidates) == 0 {
		candidates = r.catalog.AllModelIDs()
	}
	if len(candidates) == 0 {
		return "", 0, Resolution{Action: ActionDeny, Reason: "No models available"}
	}

	valid := make([]candidate, 0, len(candidates))
	for _, id := range candidates {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		caps, ok := r.catalog.Get(id)
		if !ok {
			continue
		}
		caps = caps.Effective(req.ProviderID)
		if err := validateCaps(caps, req); err != nil {
			continue
		}
		cost := estimateCostUSD(caps, req.PromptTokens, req.OutputTokens)
		if req.MaxCostUSD > 0 && cost > req.MaxCostUSD {
			continue
		}
		valid = append(valid, candidate{ID: id, CostUSD: cost})
	}

	if len(valid) == 0 && len(req.FallbackChain) > 0 {
		for _, id := range req.FallbackChain {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			caps, ok := r.catalog.Get(id)
			if !ok {
				continue
			}
			caps = caps.Effective(req.ProviderID)
			if err := validateCaps(caps, req); err != nil {
				continue
			}
			cost := estimateCostUSD(caps, req.PromptTokens, req.OutputTokens)
			if req.MaxCostUSD > 0 && cost > req.MaxCostUSD {
				continue
			}
			return id, cost, Resolution{Action: ActionFallback, TargetModel: id, Reason: "Selected fallback chain model"}
		}
	}

	if len(valid) == 0 {
		return "", 0, Resolution{Action: ActionDeny, Reason: "No candidate model satisfies constraints"}
	}

	sort.Slice(valid, func(i, j int) bool {
		if valid[i].CostUSD == valid[j].CostUSD {
			return valid[i].ID < valid[j].ID
		}
		return valid[i].CostUSD < valid[j].CostUSD
	})

	chosen := valid[0]
	if math.IsInf(chosen.CostUSD, 1) {
		return chosen.ID, chosen.CostUSD, Resolution{Action: ActionAllow, Reason: "Selected model without pricing metadata"}
	}
	return chosen.ID, chosen.CostUSD, Resolution{Action: ActionAllow, Reason: "Selected lowest cost model meeting constraints"}
}

type candidate struct {
	ID      string
	CostUSD float64
}

func estimateCostUSD(caps ModelCapabilities, promptTokens int, outputTokens int) float64 {
	if caps.InputPrice1M == 0 && caps.OutputPrice1M == 0 && caps.RequestFixedCostUSD == 0 {
		return math.Inf(1)
	}
	cost := caps.RequestFixedCostUSD
	cost += (float64(promptTokens) / 1_000_000.0) * caps.InputPrice1M
	cost += (float64(outputTokens) / 1_000_000.0) * caps.OutputPrice1M
	return cost
}

func validateCaps(caps ModelCapabilities, req RouteRequest) error {
	total := req.PromptTokens + req.OutputTokens
	if caps.ContextWindow > 0 && total > caps.ContextWindow {
		return fmt.Errorf("context window exceeded")
	}
	if caps.MaxOutputTokens > 0 && req.OutputTokens > caps.MaxOutputTokens {
		return fmt.Errorf("max output exceeded")
	}
	if req.RequiresTools && !caps.SupportsTools {
		return errors.New("tools unsupported")
	}
	if req.RequiresVision && !caps.SupportsVision {
		return errors.New("vision unsupported")
	}
	if req.ThinkingTokens > 0 && !caps.SupportsThinking {
		return errors.New("thinking unsupported")
	}
	if caps.MaxThinkingTokens > 0 && req.ThinkingTokens > caps.MaxThinkingTokens {
		return errors.New("thinking budget exceeded")
	}
	return nil
}
