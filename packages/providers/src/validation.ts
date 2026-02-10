/**
 * Provider Configuration Validation
 *
 * Provides validation functions for AI provider configurations,
 * including comprehensive checks for provider settings and constraints.
 */

import {
  ProviderConfig,
  ProviderConfigSchema,
  ValidationResult,
  ProviderValidationError,
} from './types.js';

/**
 * Validates a provider configuration against the schema and business rules
 * @param config - The configuration object to validate
 * @returns Validation result with validity flag and error messages
 */
export function validateProviderConfig(config: unknown): ValidationResult {
  const result = ProviderConfigSchema.safeParse(config);
  
  if (result.success) {
    const errors: string[] = [];
    const validated = result.data;
    
    if (validated.defaultModel) {
      const modelExists = validated.models.some((m) => m.id === validated.defaultModel);
      if (!modelExists) {
        errors.push(`Default model "${validated.defaultModel}" not found in models list`);
      }
    }
    
    if (validated.type === 'custom' && !validated.baseUrl) {
      errors.push('Custom providers require a baseUrl');
    }
    
    if (validated.type === 'local' && validated.apiKey) {
      errors.push('Local providers should not have an API key');
    }
    

    
    return {
      valid: errors.length === 0,
      errors,
    };
  }
  
  return {
    valid: false,
    errors: result.error.errors.map((e) => `${e.path.join('.')}: ${e.message}`),
  };
}

/**
 * Validates a provider configuration and throws an error if invalid
 * @param config - The configuration object to validate
 * @returns The validated ProviderConfig object
 * @throws ProviderValidationError if validation fails
 */
export function assertValidProviderConfig(config: unknown): ProviderConfig {
  const result = validateProviderConfig(config);
  
  if (!result.valid) {
    throw new ProviderValidationError(result.errors);
  }
  
  return config as ProviderConfig;
}

/**
 * Checks if a provider ID is valid (lowercase alphanumeric with hyphens/underscores)
 * @param id - The provider ID to validate
 * @returns True if the ID is valid, false otherwise
 */
export function isValidProviderId(id: string): boolean {
  return /^[a-z0-9_-]+$/.test(id) && id.length >= 1 && id.length <= 64;
}

/**
 * Checks if a string is a valid provider type
 * @param type - The type string to validate
 * @returns Type predicate indicating if the type is valid
 */
export function isValidProviderType(type: string): type is ProviderConfig['type'] {
  return ['openai', 'anthropic', 'google', 'local', 'custom'].includes(type);
}

/**
 * Validates if a string is a valid URL
 * @param url - The URL string to validate
 * @returns True if the URL is valid, false otherwise
 */
export function isValidUrl(url: string): boolean {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}
