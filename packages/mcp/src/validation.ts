/**
 * MCP Configuration Validation
 *
 * Provides validation functions for MCP server configurations,
 * including transport-specific validation and business rules.
 */

import {
  MCPServerConfig,
  MCPServerConfigSchema,
  ValidationResult,
  MCPValidationError,
  TransportType,
} from './types.js';

/**
 * Validates an MCP server configuration
 * @param config - The configuration object to validate
 * @returns Validation result with validity flag and error messages
 */
export function validateMCPServerConfig(config: unknown): ValidationResult {
  const result = MCPServerConfigSchema.safeParse(config);
  
  if (result.success) {
    const errors: string[] = [];
    const validated = result.data;
    
    switch (validated.transport.type) {
      case 'stdio': {
        if (!validated.transport.command) {
          errors.push('stdio transport requires command');
        }
        break;
      }
      case 'sse':
      case 'websocket': {
        if (!validated.transport.url) {
          errors.push(`${validated.transport.type} transport requires url`);
        }
        if (validated.transport.type === 'websocket' && 
            !validated.transport.url.match(/^wss?:\/\//)) {
          errors.push('websocket URL must start with ws:// or wss://');
        }
        break;
      }
    }
    
    if (validated.settings.fallbackServers.includes(validated.id)) {
      errors.push('Server cannot be its own fallback');
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
 * Validates an MCP server configuration and throws if invalid
 * @param config - The configuration object to validate
 * @returns The validated configuration
 * @throws MCPValidationError if validation fails
 */
export function assertValidMCPServerConfig(config: unknown): MCPServerConfig {
  const result = validateMCPServerConfig(config);
  
  if (!result.valid) {
    throw new MCPValidationError(result.errors);
  }
  
  return config as MCPServerConfig;
}

/**
 * Validates a server ID format
 * @param id - The server ID to validate
 * @returns True if the ID is valid
 */
export function isValidServerId(id: string): boolean {
  return /^[a-z0-9_-]+$/.test(id) && id.length >= 1 && id.length <= 64;
}

/**
 * Validates a transport type
 * @param type - The transport type to validate
 * @returns True if the transport type is valid
 */
export function isValidTransportType(type: string): type is TransportType {
  return ['stdio', 'sse', 'websocket'].includes(type);
}

/**
 * Validates a URL string
 * @param url - The URL to validate
 * @returns True if the URL is valid
 */
export function isValidUrl(url: string): boolean {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}

/**
 * Validates a WebSocket URL
 * @param url - The URL to validate
 * @returns True if the URL is a valid WebSocket URL
 */
export function isValidWebSocketUrl(url: string): boolean {
  return isValidUrl(url) && /^wss?:\/\//.test(url);
}

/**
 * Calculates exponential backoff delay
 * @param attempt - The retry attempt number
 * @param baseMs - The base delay in milliseconds (default: 1000)
 * @returns The calculated backoff delay
 */
export function calculateBackoff(attempt: number, baseMs: number = 1000): number {
  return Math.min(baseMs * Math.pow(2, attempt), 30000);
}
