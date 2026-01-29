import { describe, it, expect } from 'vitest';
import {
  validateMCPServerConfig,
  assertValidMCPServerConfig,
  isValidServerId,
  isValidTransportType,
  isValidUrl,
  isValidWebSocketUrl,
  calculateBackoff,
} from '../../src/validation.js';
import { MCPValidationError } from '../../src/types.js';

describe('validateMCPServerConfig', () => {
  const baseConfig = {
    id: 'test-server',
    name: 'Test Server',
    enabled: true,
    transport: {
      type: 'stdio' as const,
      command: 'npx',
      args: ['-y', '@modelcontextprotocol/server-filesystem'],
    },
    settings: {
      autoConnect: true,
      timeout: 30000,
      reconnect: true,
      maxReconnectAttempts: 3,
      fallbackServers: [],
    },
  };

  it('should validate correct stdio config', () => {
    const result = validateMCPServerConfig(baseConfig);
    
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('should validate correct sse config', () => {
    const config = {
      ...baseConfig,
      transport: {
        type: 'sse' as const,
        url: 'https://api.example.com/sse',
      },
    };
    
    const result = validateMCPServerConfig(config);
    expect(result.valid).toBe(true);
  });

  it('should validate correct websocket config', () => {
    const config = {
      ...baseConfig,
      transport: {
        type: 'websocket' as const,
        url: 'wss://api.example.com/ws',
      },
    };
    
    const result = validateMCPServerConfig(config);
    expect(result.valid).toBe(true);
  });

  it('should reject config with invalid id format', () => {
    const config = { ...baseConfig, id: 'Invalid ID!' };
    const result = validateMCPServerConfig(config);
    
    expect(result.valid).toBe(false);
  });

  it('should reject config with empty name', () => {
    const config = { ...baseConfig, name: '' };
    const result = validateMCPServerConfig(config);
    
    expect(result.valid).toBe(false);
  });

  it('should reject stdio config without command', () => {
    const config = {
      ...baseConfig,
      transport: {
        type: 'stdio' as const,
        command: '',
      },
    };
    
    const result = validateMCPServerConfig(config);
    expect(result.valid).toBe(false);
  });

  it('should reject sse config without url', () => {
    const config = {
      ...baseConfig,
      transport: {
        type: 'sse' as const,
        url: '',
      },
    };
    
    const result = validateMCPServerConfig(config);
    expect(result.valid).toBe(false);
  });

  it('should reject websocket config with invalid url', () => {
    const config = {
      ...baseConfig,
      transport: {
        type: 'websocket' as const,
        url: 'https://example.com',
      },
    };
    
    const result = validateMCPServerConfig(config);
    expect(result.valid).toBe(false);
  });

  it('should reject config with self as fallback', () => {
    const config = {
      ...baseConfig,
      settings: {
        ...baseConfig.settings,
        fallbackServers: ['test-server'],
      },
    };
    
    const result = validateMCPServerConfig(config);
    expect(result.valid).toBe(false);
  });
});

describe('assertValidMCPServerConfig', () => {
  const validConfig = {
    id: 'test',
    name: 'Test',
    enabled: true,
    transport: {
      type: 'stdio' as const,
      command: 'test',
    },
    settings: {
      autoConnect: true,
      timeout: 30000,
      reconnect: true,
      maxReconnectAttempts: 3,
      fallbackServers: [],
    },
  };

  it('should return config when valid', () => {
    const result = assertValidMCPServerConfig(validConfig);
    expect(result.id).toBe('test');
  });

  it('should throw when invalid', () => {
    const config = { ...validConfig, id: '' };
    
    expect(() => assertValidMCPServerConfig(config)).toThrow(MCPValidationError);
  });
});

describe('isValidServerId', () => {
  it('should return true for valid ids', () => {
    expect(isValidServerId('filesystem')).toBe(true);
    expect(isValidServerId('github-server')).toBe(true);
    expect(isValidServerId('a')).toBe(true);
  });

  it('should return false for invalid ids', () => {
    expect(isValidServerId('')).toBe(false);
    expect(isValidServerId('Invalid ID')).toBe(false);
    expect(isValidServerId('test@server')).toBe(false);
    expect(isValidServerId('a'.repeat(65))).toBe(false);
  });
});

describe('isValidTransportType', () => {
  it('should return true for valid types', () => {
    expect(isValidTransportType('stdio')).toBe(true);
    expect(isValidTransportType('sse')).toBe(true);
    expect(isValidTransportType('websocket')).toBe(true);
  });

  it('should return false for invalid types', () => {
    expect(isValidTransportType('http')).toBe(false);
    expect(isValidTransportType('grpc')).toBe(false);
    expect(isValidTransportType('')).toBe(false);
  });
});

describe('isValidUrl', () => {
  it('should return true for valid URLs', () => {
    expect(isValidUrl('https://api.example.com')).toBe(true);
    expect(isValidUrl('http://localhost:3000')).toBe(true);
    expect(isValidUrl('wss://ws.example.com')).toBe(true);
  });

  it('should return false for invalid URLs', () => {
    expect(isValidUrl('not-a-url')).toBe(false);
    expect(isValidUrl('')).toBe(false);
  });
});

describe('isValidWebSocketUrl', () => {
  it('should return true for valid WebSocket URLs', () => {
    expect(isValidWebSocketUrl('ws://localhost:3000')).toBe(true);
    expect(isValidWebSocketUrl('wss://api.example.com/ws')).toBe(true);
  });

  it('should return false for non-WebSocket URLs', () => {
    expect(isValidWebSocketUrl('https://example.com')).toBe(false);
    expect(isValidWebSocketUrl('http://localhost')).toBe(false);
  });
});

describe('calculateBackoff', () => {
  it('should calculate exponential backoff', () => {
    expect(calculateBackoff(0, 1000)).toBe(1000);
    expect(calculateBackoff(1, 1000)).toBe(2000);
    expect(calculateBackoff(2, 1000)).toBe(4000);
    expect(calculateBackoff(3, 1000)).toBe(8000);
  });

  it('should cap at 30 seconds', () => {
    expect(calculateBackoff(10, 1000)).toBe(30000);
  });

  it('should use default base of 1000ms', () => {
    expect(calculateBackoff(0)).toBe(1000);
  });
});
