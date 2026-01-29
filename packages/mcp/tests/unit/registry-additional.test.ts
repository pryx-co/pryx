import { describe, it, expect, beforeEach } from 'vitest';
import { MCPRegistry } from '../../src/registry.js';
import {
  MCPServerNotFoundError,
  MCPServerAlreadyExistsError,
  MCPValidationError,
} from '../../src/types.js';

describe('MCPRegistry Additional Coverage', () => {
  let registry: MCPRegistry;

  beforeEach(() => {
    registry = new MCPRegistry();
  });

  describe('enable/disable operations', () => {
    it('should enable server', () => {
      registry.addServer({
        id: 'test',
        name: 'Test',
        enabled: false,
        source: 'manual',
        transport: { type: 'stdio', command: 'test', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      registry.enableServer('test');
      expect(registry.getServer('test')?.enabled).toBe(true);
    });

    it('should disable server', () => {
      registry.addServer({
        id: 'test',
        name: 'Test',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'test', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      registry.disableServer('test');
      expect(registry.getServer('test')?.enabled).toBe(false);
    });

    it('should enable all servers', () => {
      registry.addServer({
        id: 'server1',
        name: 'Server 1',
        enabled: false,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd1', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });
      registry.addServer({
        id: 'server2',
        name: 'Server 2',
        enabled: false,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd2', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      registry.enableAll();
      expect(registry.getServer('server1')?.enabled).toBe(true);
      expect(registry.getServer('server2')?.enabled).toBe(true);
    });

    it('should disable all servers', () => {
      registry.addServer({
        id: 'server1',
        name: 'Server 1',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd1', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });
      registry.addServer({
        id: 'server2',
        name: 'Server 2',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd2', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      registry.disableAll();
      expect(registry.getServer('server1')?.enabled).toBe(false);
      expect(registry.getServer('server2')?.enabled).toBe(false);
    });

    it('should enable servers by type', () => {
      registry.addServer({
        id: 'stdio-server',
        name: 'Stdio Server',
        enabled: false,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });
      registry.addServer({
        id: 'sse-server',
        name: 'SSE Server',
        enabled: false,
        source: 'manual',
        transport: { type: 'sse', url: 'https://example.com', headers: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      registry.enableType('stdio');
      expect(registry.getServer('stdio-server')?.enabled).toBe(true);
      expect(registry.getServer('sse-server')?.enabled).toBe(false);
    });

    it('should disable servers by type', () => {
      registry.addServer({
        id: 'stdio-server',
        name: 'Stdio Server',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });
      registry.addServer({
        id: 'sse-server',
        name: 'SSE Server',
        enabled: true,
        source: 'manual',
        transport: { type: 'sse', url: 'https://example.com', headers: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      registry.disableType('stdio');
      expect(registry.getServer('stdio-server')?.enabled).toBe(false);
      expect(registry.getServer('sse-server')?.enabled).toBe(true);
    });
  });

  describe('updateServerStatus', () => {
    it('should update server status', () => {
      registry.addServer({
        id: 'test',
        name: 'Test',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      registry.updateServerStatus('test', {
        connected: true,
        lastConnected: new Date().toISOString(),
      });

      const server = registry.getServer('test');
      expect(server?.status?.connected).toBe(true);
      expect(server?.status?.lastConnected).toBeDefined();
    });

    it('should throw when updating status of non-existent server', () => {
      expect(() => {
        registry.updateServerStatus('nonexistent', { connected: true });
      }).toThrow(MCPServerNotFoundError);
    });
  });

  describe('validateServer', () => {
    it('should return valid for valid server', () => {
      registry.addServer({
        id: 'test',
        name: 'Test',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      const result = registry.validateServer('test');
      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should return invalid for non-existent server', () => {
      const result = registry.validateServer('nonexistent');
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Server not found: nonexistent');
    });
  });

  describe('testConnection', () => {
    it('should return success for existing server', async () => {
      registry.addServer({
        id: 'test',
        name: 'Test',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
        capabilities: {
          tools: [{ name: 'tool1', description: 'Tool 1', inputSchema: {} }],
          resources: [],
          prompts: [],
        },
      });

      const result = await registry.testConnection('test');
      expect(result.success).toBe(true);
      expect(result.latency).toBeDefined();
      expect(result.capabilities?.tools).toHaveLength(1);
    });

    it('should return error for non-existent server', async () => {
      const result = await registry.testConnection('nonexistent');
      expect(result.success).toBe(false);
      expect(result.error).toBe('Server not found: nonexistent');
    });
  });

  describe('getReconnectDelay', () => {
    it('should return 0 for non-existent server', () => {
      const delay = registry.getReconnectDelay('nonexistent');
      expect(delay).toBe(0);
    });

    it('should return 0 for server without status', () => {
      registry.addServer({
        id: 'test',
        name: 'Test',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      const delay = registry.getReconnectDelay('test');
      expect(delay).toBe(0);
    });
  });

  describe('fromJSON', () => {
    it('should throw for unsupported version', () => {
      expect(() => {
        registry.fromJSON({
          version: 999,
          servers: [],
        });
      }).toThrow(MCPValidationError);
    });

    it('should clear existing servers when loading', () => {
      registry.addServer({
        id: 'old',
        name: 'Old',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      registry.fromJSON({
        version: 1,
        servers: [{
          id: 'new',
          name: 'New',
          enabled: true,
          source: 'manual',
          transport: { type: 'stdio', command: 'cmd', args: [], env: {} },
          settings: {
            autoConnect: true,
            timeout: 30000,
            reconnect: true,
            maxReconnectAttempts: 3,
            fallbackServers: [],
          },
        }],
      });

      expect(registry.hasServer('old')).toBe(false);
      expect(registry.hasServer('new')).toBe(true);
    });
  });

  describe('clear', () => {
    it('should clear all servers', () => {
      registry.addServer({
        id: 'test',
        name: 'Test',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      registry.clear();
      expect(registry.size).toBe(0);
      expect(registry.hasServer('test')).toBe(false);
    });
  });

  describe('getFallbackServers', () => {
    it('should return empty array for non-existent server', () => {
      const fallbacks = registry.getFallbackServers('nonexistent');
      expect(fallbacks).toEqual([]);
    });

    it('should filter out non-existent fallback servers', () => {
      registry.addServer({
        id: 'primary',
        name: 'Primary',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: ['exists', 'missing'],
        },
      });
      registry.addServer({
        id: 'exists',
        name: 'Exists',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'cmd', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      });

      const fallbacks = registry.getFallbackServers('primary');
      expect(fallbacks).toHaveLength(1);
      expect(fallbacks[0].id).toBe('exists');
    });
  });
});
