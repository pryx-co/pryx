import { z } from 'zod';

export const ProviderType = z.enum(['openai', 'anthropic', 'google', 'local', 'custom']);

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

export const RateLimitConfigSchema = z.object({
  requestsPerMinute: z.number().int().positive().optional(),
  tokensPerMinute: z.number().int().positive().optional(),
  requestsPerDay: z.number().int().positive().optional(),
});

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

export const ProvidersConfigSchema = z.object({
  version: z.number().int().default(1),
  defaultProvider: z.string().optional(),
  providers: z.array(ProviderConfigSchema),
});

export const ValidationResultSchema = z.object({
  valid: z.boolean(),
  errors: z.array(z.string()),
});

export const ConnectionTestResultSchema = z.object({
  success: z.boolean(),
  latency: z.number().optional(),
  error: z.string().optional(),
  modelsAvailable: z.array(z.string()).optional(),
});

export type ProviderType = z.infer<typeof ProviderType>;
export type ModelConfig = z.infer<typeof ModelConfigSchema>;
export type RateLimitConfig = z.infer<typeof RateLimitConfigSchema>;
export type ProviderConfig = z.infer<typeof ProviderConfigSchema>;
export type ProvidersConfig = z.infer<typeof ProvidersConfigSchema>;
export type ValidationResult = z.infer<typeof ValidationResultSchema>;
export type ConnectionTestResult = z.infer<typeof ConnectionTestResultSchema>;

export class ProviderError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ProviderError';
  }
}

export class ProviderNotFoundError extends ProviderError {
  constructor(id: string) {
    super(`Provider not found: ${id}`);
    this.name = 'ProviderNotFoundError';
  }
}

export class ProviderValidationError extends ProviderError {
  constructor(public errors: string[]) {
    super(`Validation failed: ${errors.join(', ')}`);
    this.name = 'ProviderValidationError';
  }
}

export class ProviderAlreadyExistsError extends ProviderError {
  constructor(id: string) {
    super(`Provider already exists: ${id}`);
    this.name = 'ProviderAlreadyExistsError';
  }
}

export const CURRENT_VERSION = 1;
