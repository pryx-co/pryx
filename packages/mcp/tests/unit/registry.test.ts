import { describe, it, expect, beforeEach } from 'vitest';
import { MCPRegistry, createRegistry } from '../../src/registry.js';
import {
  MCPServerNotFoundError,
  MCPServerAlreadyExistsError,
  MCPValidationError,
} from '../../src/types.js';

const defaultSettings = {
  autoConnect: true,
  timeout: 30000,
  reconnect: true,
  maxReconnectAttempts: 3,
  fallbackServers: [],
};

describe('MCPRegistry', () => {
  let registry: MCPRegistry;

  beforeEach(() => {
    registry = new MCPRegistry();
  });

  describe('constructor', () => {
    it('should create empty registry', () => {
      expect(registry.size).toBe(0);
    });
  });

  describe('addServer', () => {
    it('should add stdio server', () => {
      const server = {
        id: 'filesystem',
        name: 'Filesystem Server',
        enabled: true,
        source: 'manual' as const,
        transport: {
          type: 'stdio' as const,
          command: 'npx',
          args: ['-y', '@modelcontextprotocol/server-filesystem'],
          env: {},
        },
        settings: defaultSettings,
      };

      registry.addServer(server);

      expect(registry.hasServer('filesystem')).toBe(true);
      expect(registry.getServer('filesystem')?.name).toBe('Filesystem Server');
    });

    it('should add sse server', () => {
      const server = {
        id: 'remote-api',
        name: 'Remote API',
        enabled: true,
        source: 'manual' as const,
        transport: {
          type: 'sse' as const,
          url: 'https://api.example.com/sse',
          headers: {},
        },
        settings: defaultSettings,
      };

      registry.addServer(server);

      expect(registry.hasServer('remote-api')).toBe(true);
    });

    it('should add websocket server', () => {
      const server = {
        id: 'ws-server',
        name: 'WebSocket Server',
        enabled: true,
        source: 'curated' as const,
        transport: {
          type: 'websocket' as const,
          url: 'wss://ws.example.com/mcp',
          headers: {},
        },
        settings: defaultSettings,
      };

      registry.addServer(server);

      expect(registry.hasServer('ws-server')).toBe(true);
    });

    it('should throw when server already exists', () => {
      const server = {
        id: 'test',
        name: 'Test',
        enabled: true,
        source: 'manual' as const,
        transport: {
          type: 'stdio' as const,
          command: 'test',
          args: [],
          env: {},
        },
        settings: defaultSettings,
      };

      registry.addServer(server);
      
      expect(() => registry.addServer(server)).toThrow(MCPServerAlreadyExistsError);
    });
  });

  describe('updateServer', () => {
    it('should update existing server', () => {
      registry.addServer({
        id: 'test',
        name: 'Test',
        enabled: true,
        source: 'manual' as const,
        transport: {
          type: 'stdio' as const,
          command: 'test',
          args: [],
          env: {},
        },
        settings: defaultSettings,
      });

      const updated = registry.updateServer('test', { name: 'Updated' });

      expect(updated.name).toBe('Updated');
      expect(registry.getServer('test')?.name).toBe('Updated');
    });

    it('should throw when server not found', () => {
      expect(() => registry.updateServer('nonexistent', { name: 'Test' })).toThrow(MCPServerNotFoundError);
    });
  });

  describe('removeServer', () => {
    it('should remove server', () => {
      registry.addServer({
        id: 'test',
        name: 'Test',
        enabled: true,
        source: 'manual' as const,
        transport: {
          type: 'stdio' as const,
          command: 'test',
          args: [],
          env: {},
        },
        settings: defaultSettings,
      });

      registry.removeServer('test');

      expect(registry.hasServer('test')).toBe(false);
    });

    it('should throw when server not found', () => {
      expect(() => registry.removeServer('nonexistent')).toThrow(MCPServerNotFoundError);
    });
  });

  describe('getAllServers', () => {
    it('should return all servers', () => {
      registry.addServer({
        id: 'server1',
        name: 'Server 1',
        enabled: true,
        source: 'manual' as const,
        transport: { type: 'stdio' as const, command: 'cmd1', args: [], env: {} },
        settings: defaultSettings,
      });
      registry.addServer({
        id: 'server2',
        name: 'Server 2',
        enabled: true,
        source: 'curated' as const,
        transport: { type: 'stdio' as const, command: 'cmd2', args: [], env: {} },
        settings: defaultSettings,
      });

      const servers = registry.getAllServers();

      expect(servers.length).toBe(2);
    });
  });

  describe('getEnabledServers', () => {
    it('should return only enabled servers', () => {
      registry.addServer({
        id: 'enabled',
        name: 'Enabled',
        enabled: true,
        source: 'manual' as const,
        transport: { type: 'stdio' as const, command: 'cmd', args: [], env: {} },
        settings: defaultSettings,
      });
      registry.addServer({
        id: 'disabled',
        name: 'Disabled',
        enabled: false,
        source: 'manual' as const,
        transport: { type: 'stdio' as const, command: 'cmd', args: [], env: {} },
        settings: defaultSettings,
      });

      const enabled = registry.getEnabledServers();

      expect(enabled.length).toBe(1);
      expect(enabled[0].id).toBe('enabled');
    });
  });

  describe('getServersByType', () => {
    it('should return servers by transport type', () => {
      registry.addServer({
        id: 'stdio1',
        name: 'Stdio 1',
        enabled: true,
        source: 'manual' as const,
        transport: { type: 'stdio' as const, command: 'cmd', args: [], env: {} },
        settings: defaultSettings,
      });
      registry.addServer({
        id: 'sse1',
        name: 'SSE 1',
        enabled: true,
        source: 'curated' as const,
        transport: { type: 'sse' as const, url: 'https://example.com', headers: {} },
        settings: defaultSettings,
      });

      const stdioServers = registry.getServersByType('stdio');

      expect(stdioServers.length).toBe(1);
      expect(stdioServers[0].id).toBe('stdio1');
    });
  });

  describe('fallback servers', () => {
    it('should return fallback servers', () => {
      registry.addServer({
        id: 'primary',
        name: 'Primary',
        enabled: true,
        source: 'manual' as const,
        transport: { type: 'stdio' as const, command: 'cmd', args: [], env: {} },
        settings: {
          ...defaultSettings,
          fallbackServers: ['fallback1', 'fallback2'],
        },
      });
      registry.addServer({
        id: 'fallback1',
        name: 'Fallback 1',
        enabled: true,
        source: 'manual' as const,
        transport: { type: 'stdio' as const, command: 'cmd', args: [], env: {} },
        settings: defaultSettings,
      });
      registry.addServer({
        id: 'fallback2',
        name: 'Fallback 2',
        enabled: true,
        source: 'manual' as const,
        transport: { type: 'stdio' as const, command: 'cmd', args: [], env: {} },
        settings: defaultSettings,
      });

      const fallbacks = registry.getFallbackServers('primary');

      expect(fallbacks.length).toBe(2);
    });
  });

  describe('getReconnectDelay', () => {
    it('should calculate reconnect delay', () => {
      registry.addServer({
        id: 'test',
        name: 'Test',
        enabled: true,
        source: 'manual' as const,
        transport: { type: 'stdio' as const, command: 'cmd', args: [], env: {} },
        settings: defaultSettings,
        status: {
          connected: false,
          reconnectAttempts: 2,
        },
      });

      const delay = registry.getReconnectDelay('test');

      expect(delay).toBe(4000);
    });
  });

  describe('toJSON / fromJSON', () => {
    it('should serialize to JSON', () => {
      registry.addServer({
        id: 'test',
        name: 'Test',
        enabled: true,
        source: 'manual' as const,
        transport: { type: 'stdio' as const, command: 'cmd', args: [], env: {} },
        settings: defaultSettings,
      });

      const json = registry.toJSON();

      expect(json.version).toBe(1);
      expect(json.servers.length).toBe(1);
    });

    it('should deserialize from JSON', () => {
      const json = {
        version: 1,
        servers: [{
          id: 'test',
          name: 'Test',
          enabled: true,
          source: 'manual' as const,
          transport: { type: 'stdio' as const, command: 'cmd', args: [], env: {} },
          settings: defaultSettings,
        }],
      };

      registry.fromJSON(json);

      expect(registry.size).toBe(1);
      expect(registry.hasServer('test')).toBe(true);
    });
  });
});

describe('createRegistry', () => {
  it('should create new registry', () => {
    const registry = createRegistry();

    expect(registry).toBeInstanceOf(MCPRegistry);
    expect(registry.size).toBe(0);
  });
});
