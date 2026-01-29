import {
  ChannelConfig,
  ChannelConfigSchema,
  ValidationResult,
  ChannelValidationError,
  ChannelType,
} from './types.js';

export function validateChannelConfig(config: unknown): ValidationResult {
  const result = ChannelConfigSchema.safeParse(config);
  
  if (result.success) {
    const errors: string[] = [];
    const validated = result.data;
    
    if (validated.webhook?.enabled && !validated.webhook.url) {
      errors.push('Webhook settings enabled but URL is missing');
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

export function assertValidChannelConfig(config: unknown): ChannelConfig {
  const result = validateChannelConfig(config);
  
  if (!result.valid) {
    throw new ChannelValidationError(result.errors);
  }
  
  return config as ChannelConfig;
}

export function isValidChannelId(id: string): boolean {
  return /^[a-z0-9_-]+$/.test(id) && id.length >= 1 && id.length <= 64;
}

export function isValidChannelType(type: string): type is ChannelType {
  return ['telegram', 'discord', 'slack', 'email', 'whatsapp', 'webhook'].includes(type);
}

export function matchesFilterPatterns(message: string, patterns: string[]): boolean {
  if (patterns.length === 0) return true;
  
  return patterns.some((pattern) => {
    try {
      const regex = new RegExp(pattern, 'i');
      return regex.test(message);
    } catch {
      return message.toLowerCase().includes(pattern.toLowerCase());
    }
  });
}

export function isUserAllowed(userId: string, allowedList: string[], blockedList: string[]): boolean {
  if (blockedList.includes(userId)) {
    return false;
  }
  
  if (allowedList.length > 0 && !allowedList.includes(userId)) {
    return false;
  }
  
  return true;
}
