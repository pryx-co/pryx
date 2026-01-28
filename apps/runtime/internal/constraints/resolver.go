package constraints

import (
	"fmt"
)

type Resolver struct {
	catalog *Catalog
}

func NewResolver(catalog *Catalog) *Resolver {
	if catalog == nil {
		catalog = DefaultCatalog()
	}
	return &Resolver{catalog: catalog}
}

func (r *Resolver) CalculateCost(req Request, caps ModelCapabilities) CostEstimate {
	inputCost := (float64(req.PromptTokens) / 1000000.0) * caps.InputPrice1M
	outputCost := (float64(req.OutputTokens) / 1000000.0) * caps.OutputPrice1M
	thinkingCost := (float64(req.ThinkingTokens) / 1000000.0) * caps.InputPrice1M // Assuming thinking tokens cost same as input usually, or specific if needed. Often treated as output. Let's use Output assuming it's generated.
	// Actually thinking tokens are usually output tokens.
	thinkingCost = (float64(req.ThinkingTokens) / 1000000.0) * caps.OutputPrice1M

	total := inputCost + outputCost + thinkingCost + caps.RequestFixedCostUSD

	return CostEstimate{
		InputTokensCost:    inputCost,
		OutputTokensCost:   outputCost,
		ThinkingTokensCost: thinkingCost,
		FixedCost:          caps.RequestFixedCostUSD,
		TotalUSD:           total,
	}
}

func (r *Resolver) findFallbackModel(modelID string, req Request) (string, bool) {
	caps, ok := r.catalog.Get(modelID)
	if !ok {
		return "", false
	}

	for _, fallbackID := range caps.FallbackChain {
		fbCaps, ok := r.catalog.Get(fallbackID)
		if !ok {
			continue
		}

		totalTokens := req.PromptTokens + req.OutputTokens
		if totalTokens <= fbCaps.ContextWindow {
			if fbCaps.MaxOutputTokens == 0 || req.OutputTokens <= fbCaps.MaxOutputTokens {
				if req.ThinkingTokens == 0 || req.ThinkingTokens <= fbCaps.MaxThinkingTokens {
					return fallbackID, true
				}
			}
		}
	}

	return "", false
}

func (r *Resolver) Resolve(req Request) Resolution {
	caps, ok := r.catalog.Get(req.Model)
	if !ok {
		// Unknown model, allow but warn (or strictly deny? strictly deny safer for budget)
		// For MVP, allow with warning reasoning
		return Resolution{
			Action: ActionAllow,
			Reason: fmt.Sprintf("Unknown model %s, skipping constraints", req.Model),
		}
	}

	caps = caps.Effective(req.ProviderID)

	// Calculate Estimated Cost
	costEst := r.CalculateCost(req, caps)

	// 0. Check Max Cost Constraint
	if req.MaxCostUSD > 0 && costEst.TotalUSD > req.MaxCostUSD {
		return Resolution{
			Action:           ActionDeny,
			Reason:           fmt.Sprintf("Estimated cost $%.4f exceeds limit $%.4f", costEst.TotalUSD, req.MaxCostUSD),
			EstimatedCostUSD: costEst.TotalUSD,
		}
	}

	// 1. Check Context Window
	totalEstimatedTokens := req.PromptTokens + req.OutputTokens
	if totalEstimatedTokens > caps.ContextWindow {
		// Try fallback if defined in model capabilities
		if fallbackModel := caps.FallbackModel(); fallbackModel != "" {
			// Verify fallback exists (prevent infinite loops or bad targets)
			if _, ok := r.catalog.Get(fallbackModel); ok {
				return Resolution{
					Action:           ActionFallback,
					TargetModel:      fallbackModel,
					Reason:           fmt.Sprintf("Request exceeds context window (%d > %d)", totalEstimatedTokens, caps.ContextWindow),
					EstimatedCostUSD: costEst.TotalUSD,
				}
			}
		}

		return Resolution{
			Action:           ActionDeny,
			Reason:           fmt.Sprintf("Request exceeds context window (%d > %d) and no valid fallback available", totalEstimatedTokens, caps.ContextWindow),
			EstimatedCostUSD: costEst.TotalUSD,
		}
	}

	// 2. Check Output Token Limit
	if caps.MaxOutputTokens > 0 && req.OutputTokens > caps.MaxOutputTokens {
		if fallbackModel := caps.FallbackModel(); fallbackModel != "" {
			if fb, ok := r.catalog.Get(fallbackModel); ok {
				fb = fb.Effective(req.ProviderID)
				if fb.MaxOutputTokens == 0 || req.OutputTokens <= fb.MaxOutputTokens {
					return Resolution{
						Action:           ActionFallback,
						TargetModel:      fallbackModel,
						Reason:           fmt.Sprintf("Request exceeds max output tokens (%d > %d)", req.OutputTokens, caps.MaxOutputTokens),
						EstimatedCostUSD: costEst.TotalUSD,
					}
				}
			}
		}
		return Resolution{
			Action:           ActionDeny,
			Reason:           fmt.Sprintf("Request exceeds max output tokens (%d > %d)", req.OutputTokens, caps.MaxOutputTokens),
			EstimatedCostUSD: costEst.TotalUSD,
		}
	}

	// 3. Check Feature Capabilities
	if caps.MaxToolsPerRequest > 0 && len(req.Tools) > caps.MaxToolsPerRequest {
		return Resolution{
			Action:           ActionDeny,
			Reason:           fmt.Sprintf("Too many tools (%d > %d)", len(req.Tools), caps.MaxToolsPerRequest),
			EstimatedCostUSD: costEst.TotalUSD,
		}
	}

	if len(req.Tools) > 0 && !caps.SupportsTools {
		return Resolution{
			Action:           ActionDeny,
			Reason:           "Model does not support tools",
			EstimatedCostUSD: costEst.TotalUSD,
		}
	}

	if req.Images && !caps.SupportsVision {
		return Resolution{
			Action:           ActionDeny,
			Reason:           "Model does not support vision",
			EstimatedCostUSD: costEst.TotalUSD,
		}
	}

	if req.ThinkingTokens > 0 && !caps.SupportsThinking {
		return Resolution{
			Action:           ActionDeny,
			Reason:           "Model does not support thinking",
			EstimatedCostUSD: costEst.TotalUSD,
		}
	}

	if caps.MaxThinkingTokens > 0 && req.ThinkingTokens > caps.MaxThinkingTokens {
		if fallbackModel := caps.FallbackModel(); fallbackModel != "" {
			if fb, ok := r.catalog.Get(fallbackModel); ok {
				fb = fb.Effective(req.ProviderID)
				if fb.MaxThinkingTokens == 0 || req.ThinkingTokens <= fb.MaxThinkingTokens {
					return Resolution{
						Action:           ActionFallback,
						TargetModel:      fallbackModel,
						Reason:           fmt.Sprintf("Request exceeds thinking budget (%d > %d)", req.ThinkingTokens, caps.MaxThinkingTokens),
						EstimatedCostUSD: costEst.TotalUSD,
					}
				}
			}
		}
		return Resolution{
			Action:           ActionDeny,
			Reason:           fmt.Sprintf("Request exceeds thinking budget (%d > %d)", req.ThinkingTokens, caps.MaxThinkingTokens),
			EstimatedCostUSD: costEst.TotalUSD,
		}
	}

	return Resolution{
		Action:           ActionAllow,
		EstimatedCostUSD: costEst.TotalUSD,
	}
}
