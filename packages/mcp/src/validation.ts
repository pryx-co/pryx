import {
  MCPServerConfig,
  MCPServerConfigSchema,
  ValidationResult,
  MCPValidationError,
  TransportType,
} from './types.js';

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

export function assertValidMCPServerConfig(config: unknown): MCPServerConfig {
  const result = validateMCPServerConfig(config);
  
  if (!result.valid) {
    throw new MCPValidationError(result.errors);
  }
  
  return config as MCPServerConfig;
}

export function isValidServerId(id: string): boolean {
  return /^[a-z0-9_-]+$/.test(id) && id.length >= 1 && id.length <= 64;
}

export function isValidTransportType(type: string): type is TransportType {
  return ['stdio', 'sse', 'websocket'].includes(type);
}

export function isValidUrl(url: string): boolean {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}

export function isValidWebSocketUrl(url: string): boolean {
  return isValidUrl(url) && /^wss?:\/\//.test(url);
}

export function calculateBackoff(attempt: number, baseMs: number = 1000): number {
  return Math.min(baseMs * Math.pow(2, attempt), 30000);
}
