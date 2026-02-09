import {
  ChannelConfig,
  ChannelConfigSchema,
  ValidationResult,
  ChannelValidationError,
  ChannelType,
} from './types.js';

/**
 * Validates a channel configuration object against the schema
 * @param config - The configuration object to validate
 * @returns Validation result with success status and any error messages
 */
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

/**
 * Asserts that a configuration is valid, throwing an error if not
 * @param config - The configuration object to validate
 * @returns The validated ChannelConfig
 * @throws {ChannelValidationError} If validation fails
 */
export function assertValidChannelConfig(config: unknown): ChannelConfig {
  const result = validateChannelConfig(config);

  if (!result.valid) {
    throw new ChannelValidationError(result.errors);
  }

  return config as ChannelConfig;
}

/**
 * Validates a channel ID format
 * @param id - The channel ID to validate
 * @returns True if the ID is valid (1-64 chars, alphanumeric, hyphen, underscore)
 */
export function isValidChannelId(id: string): boolean {
  return /^[a-z0-9_-]+$/.test(id) && id.length >= 1 && id.length <= 64;
}

/**
 * Type guard to check if a string is a valid channel type
 * @param type - The type string to check
 * @returns True if the type is valid
 */
export function isValidChannelType(type: string): type is ChannelType {
  return ['telegram', 'discord', 'slack', 'email', 'whatsapp', 'webhook'].includes(type);
}

/**
 * Checks if a message matches any of the provided filter patterns
 * @param message - The message to check
 * @param patterns - Array of regex patterns or substrings to match
 * @returns True if message matches any pattern (or if patterns array is empty)
 */
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

/**
 * Checks if a user is allowed to access a channel
 * @param userId - The user ID to check
 * @param allowedList - List of allowed user IDs (empty means allow all)
 * @param blockedList - List of blocked user IDs
 * @returns True if the user is allowed (not blocked and in allowed list if specified)
 */
export function isUserAllowed(userId: string, allowedList: string[], blockedList: string[]): boolean {
  if (blockedList.includes(userId)) {
    return false;
  }

  if (allowedList.length > 0 && !allowedList.includes(userId)) {
    return false;
  }

  return true;
}
