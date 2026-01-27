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

	// 1. Check Context Window
	totalEstimatedTokens := req.PromptTokens + req.OutputTokens
	if totalEstimatedTokens > caps.ContextWindow {
		// Try fallback if defined in model capabilities
		if caps.FallbackModel != "" {
			// Verify fallback exists (prevent infinite loops or bad targets)
			if _, ok := r.catalog.Get(caps.FallbackModel); ok {
				return Resolution{
					Action:      ActionFallback,
					TargetModel: caps.FallbackModel,
					Reason:      fmt.Sprintf("Request exceeds context window (%d > %d)", totalEstimatedTokens, caps.ContextWindow),
				}
			}
		}

		return Resolution{
			Action: ActionDeny,
			Reason: fmt.Sprintf("Request exceeds context window (%d > %d) and no valid fallback available", totalEstimatedTokens, caps.ContextWindow),
		}
	}

	// 2. Check Feature Capabilities
	if len(req.Tools) > 0 && !caps.SupportsTools {
		return Resolution{
			Action: ActionDeny,
			Reason: "Model does not support tools",
		}
	}

	if req.Images && !caps.SupportsVision {
		return Resolution{
			Action: ActionDeny,
			Reason: "Model does not support vision",
		}
	}

	return Resolution{
		Action: ActionAllow,
	}
}
