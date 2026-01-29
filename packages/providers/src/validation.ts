import {
  ProviderConfig,
  ProviderConfigSchema,
  ValidationResult,
  ProviderValidationError,
} from './types.js';

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

export function assertValidProviderConfig(config: unknown): ProviderConfig {
  const result = validateProviderConfig(config);
  
  if (!result.valid) {
    throw new ProviderValidationError(result.errors);
  }
  
  return config as ProviderConfig;
}

export function isValidProviderId(id: string): boolean {
  return /^[a-z0-9_-]+$/.test(id) && id.length >= 1 && id.length <= 64;
}

export function isValidProviderType(type: string): type is ProviderConfig['type'] {
  return ['openai', 'anthropic', 'google', 'local', 'custom'].includes(type);
}

export function isValidUrl(url: string): boolean {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}
