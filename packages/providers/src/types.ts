/**
 * Provider Types and Schemas
 *
 * Defines Zod schemas and TypeScript types for AI provider configurations,
 * including model settings, rate limits, and validation results.
 */

import { z } from 'zod';

/**
 * Enum of supported AI provider types
 */
export const ProviderType = z.enum(['openai', 'anthropic', 'google', 'local', 'custom']);

/**
 * Schema for AI model configuration
 */
export const ModelConfigSchema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  maxTokens: z.number().int().positive(),
  supportsStreaming: z.boolean().default(true),
  supportsVision: z.boolean().default(false),
  supportsTools: z.boolean().default(false),
  costPer1KInput: z.number().positive().optional(),
  costPer1KOutput: z.number().positive().optional(),
});

/**
 * Schema for rate limiting configuration
 */
export const RateLimitConfigSchema = z.object({
  requestsPerMinute: z.number().int().positive().optional(),
  tokensPerMinute: z.number().int().positive().optional(),
  requestsPerDay: z.number().int().positive().optional(),
});

/**
 * Schema for AI provider configuration
 */
export const ProviderConfigSchema = z.object({
  id: z.string().min(1).max(64).regex(/^[a-z0-9_-]+$/),
  name: z.string().min(1).max(128),
  type: ProviderType,
  enabled: z.boolean().default(true),
  defaultModel: z.string().optional(),
  apiKey: z.string().optional(),
  baseUrl: z.string().url().optional(),
  organization: z.string().optional(),
  models: z.array(ModelConfigSchema).min(1),
  rateLimits: RateLimitConfigSchema.optional(),
  timeout: z.number().int().positive().default(30000),
  retries: z.number().int().min(0).max(10).default(3),
});

/**
 * Schema for complete providers configuration
 */
export const ProvidersConfigSchema = z.object({
  version: z.number().int().default(1),
  defaultProvider: z.string().optional(),
  providers: z.array(ProviderConfigSchema),
});

/**
 * Schema for validation results
 */
export const ValidationResultSchema = z.object({
  valid: z.boolean(),
  errors: z.array(z.string()),
});

/**
 * Schema for connection test results
 */
export const ConnectionTestResultSchema = z.object({
  success: z.boolean(),
  latency: z.number().optional(),
  error: z.string().optional(),
  modelsAvailable: z.array(z.string()).optional(),
});

/**
 * Inferred TypeScript type for ProviderType enum
 */
export type ProviderType = z.infer<typeof ProviderType>;

/**
 * Inferred TypeScript type for ModelConfig schema
 */
export type ModelConfig = z.infer<typeof ModelConfigSchema>;

/**
 * Inferred TypeScript type for RateLimitConfig schema
 */
export type RateLimitConfig = z.infer<typeof RateLimitConfigSchema>;

/**
 * Inferred TypeScript type for ProviderConfig schema
 */
export type ProviderConfig = z.infer<typeof ProviderConfigSchema>;

/**
 * Inferred TypeScript type for ProvidersConfig schema
 */
export type ProvidersConfig = z.infer<typeof ProvidersConfigSchema>;

/**
 * Inferred TypeScript type for ValidationResult schema
 */
export type ValidationResult = z.infer<typeof ValidationResultSchema>;

/**
 * Inferred TypeScript type for ConnectionTestResult schema
 */
export type ConnectionTestResult = z.infer<typeof ConnectionTestResultSchema>;

/**
 * Base error class for provider-related errors
 */
export class ProviderError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ProviderError';
  }
}

/**
 * Error thrown when a requested provider is not found
 */
export class ProviderNotFoundError extends ProviderError {
  constructor(id: string) {
    super(`Provider not found: ${id}`);
    this.name = 'ProviderNotFoundError';
  }
}

/**
 * Error thrown when provider configuration validation fails
 */
export class ProviderValidationError extends ProviderError {
  constructor(public errors: string[]) {
    super(`Validation failed: ${errors.join(', ')}`);
    this.name = 'ProviderValidationError';
  }
}

/**
 * Error thrown when attempting to add a provider that already exists
 */
export class ProviderAlreadyExistsError extends ProviderError {
  constructor(id: string) {
    super(`Provider already exists: ${id}`);
    this.name = 'ProviderAlreadyExistsError';
  }
}

/**
 * Current configuration version number
 */
export const CURRENT_VERSION = 1;
