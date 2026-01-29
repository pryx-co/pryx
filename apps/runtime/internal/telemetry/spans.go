package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Span helpers for common Pryx operations

// LLMSpan creates a span for LLM operations
func (p *Provider) LLMSpan(ctx context.Context, provider, model string) (context.Context, trace.Span) {
	ctx, span := p.StartSpan(ctx, "llm.request",
		attribute.String("llm.provider", provider),
		attribute.String("llm.model", model),
	)
	return ctx, span
}

// ToolSpan creates a span for tool execution
func (p *Provider) ToolSpan(ctx context.Context, toolName string, args map[string]interface{}) (context.Context, trace.Span) {
	ctx, span := p.StartSpan(ctx, "tool.execute",
		attribute.String("tool.name", toolName),
	)

	// Add args as attributes (careful: may contain PII - filter in redaction layer)
	for k, v := range args {
		span.SetAttributes(attribute.String(fmt.Sprintf("tool.arg.%s", k), fmt.Sprintf("%v", v)))
	}

	return ctx, span
}

// SessionSpan creates a span for session operations
func (p *Provider) SessionSpan(ctx context.Context, sessionID, operation string) (context.Context, trace.Span) {
	ctx, span := p.StartSpan(ctx, fmt.Sprintf("session.%s", operation),
		attribute.String("session.id", sessionID),
	)
	return ctx, span
}

// ChannelSpan creates a span for channel operations
func (p *Provider) ChannelSpan(ctx context.Context, channelType, channelID, operation string) (context.Context, trace.Span) {
	ctx, span := p.StartSpan(ctx, fmt.Sprintf("channel.%s", operation),
		attribute.String("channel.type", channelType),
		attribute.String("channel.id", channelID),
	)
	return ctx, span
}

// MCPSpan creates a span for MCP server calls
func (p *Provider) MCPSpan(ctx context.Context, serverName, toolName string) (context.Context, trace.Span) {
	ctx, span := p.StartSpan(ctx, "mcp.call",
		attribute.String("mcp.server", serverName),
		attribute.String("mcp.tool", toolName),
	)
	return ctx, span
}

// RecordError records an error on the span
func RecordError(span trace.Span, err error) {
	if span == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// RecordSuccess marks a span as successful
func RecordSuccess(span trace.Span) {
	if span == nil {
		return
	}
	span.SetStatus(codes.Ok, "")
}

// RecordDuration records the duration of an operation
func RecordDuration(span trace.Span, start time.Time) {
	if span == nil {
		return
	}
	span.SetAttributes(attribute.Int64("duration_ms", time.Since(start).Milliseconds()))
}

// RecordTokenUsage records LLM token usage
func RecordTokenUsage(span trace.Span, inputTokens, outputTokens int) {
	if span == nil {
		return
	}
	span.SetAttributes(
		attribute.Int("llm.tokens.input", inputTokens),
		attribute.Int("llm.tokens.output", outputTokens),
		attribute.Int("llm.tokens.total", inputTokens+outputTokens),
	)
}

// RecordCost records the cost of an operation
func RecordCost(span trace.Span, cost float64, currency string) {
	if span == nil {
		return
	}
	span.SetAttributes(
		attribute.Float64("cost.amount", cost),
		attribute.String("cost.currency", currency),
	)
}
