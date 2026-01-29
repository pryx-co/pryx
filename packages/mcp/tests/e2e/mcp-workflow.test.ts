import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { MCPRegistry } from '../../src/registry.js';
import { MCPStorage } from '../../src/storage.js';
import { validateMCPServerConfig } from '../../src/validation.js';
import { MCPServerConfig } from '../../src/types.js';
import fs from 'fs';
import path from 'path';
import os from 'os';

describe('MCP End-to-End Workflow', () => {
  let tempDir: string;
  let configPath: string;
  let storage: MCPStorage;

  beforeEach(() => {
    tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'mcp-e2e-test-'));
    configPath = path.join(tempDir, 'mcp-servers.json');
    storage = new MCPStorage();
  });

  afterEach(() => {
    fs.rmSync(tempDir, { recursive: true, force: true });
  });

  describe('complete workflow', () => {
    it('should handle full server lifecycle', async () => {
      const registry = new MCPRegistry();

      const server1: MCPServerConfig = {
        id: 'filesystem',
        name: 'Filesystem Server',
        enabled: true,
        source: 'curated',
        transport: {
          type: 'stdio',
          command: 'npx',
          args: ['-y', '@modelcontextprotocol/server-filesystem', '/tmp'],
          env: {},
        },
        capabilities: {
          tools: [
            {
              name: 'read_file',
              description: 'Read a file',
              inputSchema: {
                type: 'object',
                properties: {
                  path: { type: 'string' },
                },
              },
            },
          ],
          resources: [],
          prompts: [],
        },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      };

      const server2: MCPServerConfig = {
        id: 'remote-api',
        name: 'Remote API Server',
        enabled: true,
        source: 'manual',
        transport: {
          type: 'sse',
          url: 'https://api.example.com/mcp/sse',
          headers: {
            Authorization: 'Bearer test-token',
          },
        },
        settings: {
          autoConnect: true,
          timeout: 60000,
          reconnect: true,
          maxReconnectAttempts: 5,
          fallbackServers: ['filesystem'],
        },
      };

      registry.addServer(server1);
      registry.addServer(server2);

      expect(registry.size).toBe(2);
      expect(registry.getEnabledServers()).toHaveLength(2);

      const fallbacks = registry.getFallbackServers('remote-api');
      expect(fallbacks).toHaveLength(1);
      expect(fallbacks[0].id).toBe('filesystem');

      await storage.save(configPath, registry);
      expect(fs.existsSync(configPath)).toBe(true);

      const loadedRegistry = await storage.load(configPath);
      expect(loadedRegistry.size).toBe(2);

      const loadedServer1 = loadedRegistry.getServer('filesystem');
      expect(loadedServer1?.capabilities?.tools).toHaveLength(1);
      expect(loadedServer1?.transport.type).toBe('stdio');

      const loadedServer2 = loadedRegistry.getServer('remote-api');
      expect(loadedServer2?.transport.type).toBe('sse');
      expect(loadedServer2?.settings.timeout).toBe(60000);

      loadedRegistry.updateServer('filesystem', { enabled: false });
      expect(loadedRegistry.getEnabledServers()).toHaveLength(1);

      loadedRegistry.removeServer('remote-api');
      expect(loadedRegistry.size).toBe(1);
      expect(loadedRegistry.hasServer('remote-api')).toBe(false);

      await storage.save(configPath, loadedRegistry);
      const finalRegistry = await storage.load(configPath);
      expect(finalRegistry.size).toBe(1);
      expect(finalRegistry.getServer('filesystem')?.enabled).toBe(false);
    });

    it('should validate server configs correctly', () => {
      const validStdio = {
        id: 'valid-stdio',
        name: 'Valid Stdio',
        enabled: true,
        source: 'manual',
        transport: {
          type: 'stdio',
          command: 'npx',
          args: ['-y', '@modelcontextprotocol/server-filesystem'],
          env: {},
        },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      };

      const result = validateMCPServerConfig(validStdio);
      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject invalid server configs', () => {
      const invalidServer = {
        id: 'invalid',
        name: '',
        enabled: true,
        source: 'manual',
        transport: {
          type: 'sse',
          url: 'not-a-url',
          headers: {},
        },
        settings: {
          autoConnect: true,
          timeout: -1,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      };

      const result = validateMCPServerConfig(invalidServer);
      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });

    it('should handle all transport types', async () => {
      const registry = new MCPRegistry();

      const stdioServer: MCPServerConfig = {
        id: 'stdio-server',
        name: 'Stdio Server',
        enabled: true,
        source: 'manual',
        transport: {
          type: 'stdio',
          command: 'python',
          args: ['-m', 'mcp.server'],
          env: { PYTHONPATH: '/app' },
          cwd: '/home/user',
        },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      };

      const sseServer: MCPServerConfig = {
        id: 'sse-server',
        name: 'SSE Server',
        enabled: true,
        source: 'curated',
        transport: {
          type: 'sse',
          url: 'https://sse.example.com/events',
          headers: { 'X-API-Key': 'secret' },
        },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      };

      const wsServer: MCPServerConfig = {
        id: 'ws-server',
        name: 'WebSocket Server',
        enabled: true,
        source: 'marketplace',
        transport: {
          type: 'websocket',
          url: 'wss://ws.example.com/mcp',
          headers: {},
        },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 3,
          fallbackServers: [],
        },
      };

      registry.addServer(stdioServer);
      registry.addServer(sseServer);
      registry.addServer(wsServer);

      expect(registry.getServersByType('stdio')).toHaveLength(1);
      expect(registry.getServersByType('sse')).toHaveLength(1);
      expect(registry.getServersByType('websocket')).toHaveLength(1);

      await storage.save(configPath, registry);
      const loaded = await storage.load(configPath);

      const loadedStdio = loaded.getServer('stdio-server');
      expect(loadedStdio?.transport.type).toBe('stdio');
      if (loadedStdio?.transport.type === 'stdio') {
        expect(loadedStdio.transport.cwd).toBe('/home/user');
        expect(loadedStdio.transport.env.PYTHONPATH).toBe('/app');
      }

      const loadedSse = loaded.getServer('sse-server');
      expect(loadedSse?.transport.type).toBe('sse');
      if (loadedSse?.transport.type === 'sse') {
        expect(loadedSse.transport.headers['X-API-Key']).toBe('secret');
      }

      const loadedWs = loaded.getServer('ws-server');
      expect(loadedWs?.transport.type).toBe('websocket');
    });

    it('should handle server status and reconnection', () => {
      const registry = new MCPRegistry();

      registry.addServer({
        id: 'reconnect-test',
        name: 'Reconnect Test',
        enabled: true,
        source: 'manual',
        transport: { type: 'stdio', command: 'test', args: [], env: {} },
        settings: {
          autoConnect: true,
          timeout: 30000,
          reconnect: true,
          maxReconnectAttempts: 5,
          fallbackServers: [],
        },
        status: {
          connected: false,
          reconnectAttempts: 3,
          lastError: 'Connection timeout',
        },
      });

      const delay1 = registry.getReconnectDelay('reconnect-test');
      expect(delay1).toBe(8000);

      registry.updateServer('reconnect-test', {
        status: {
          connected: false,
          reconnectAttempts: 5,
        },
      });

      const delay2 = registry.getReconnectDelay('reconnect-test');
      expect(delay2).toBe(30000);
    });

    it('should persist and restore complete server state', async () => {
      const registry = new MCPRegistry();
      const now = new Date().toISOString();

      registry.addServer({
        id: 'complete-server',
        name: 'Complete Server',
        enabled: false,
        source: 'curated',
        transport: {
          type: 'stdio',
          command: 'node',
          args: ['server.js'],
          env: { NODE_ENV: 'production' },
        },
        capabilities: {
          tools: [
            { name: 'tool1', description: 'Tool 1', inputSchema: {} },
            { name: 'tool2', description: 'Tool 2', inputSchema: {} },
          ],
          resources: [
            { uri: 'file:///data', name: 'Data', mimeType: 'application/json' },
          ],
          prompts: [
            { name: 'prompt1', description: 'Prompt 1' },
          ],
        },
        settings: {
          autoConnect: false,
          timeout: 45000,
          reconnect: false,
          maxReconnectAttempts: 10,
          fallbackServers: ['backup1', 'backup2'],
        },
        status: {
          connected: true,
          lastConnected: now,
          reconnectAttempts: 0,
        },
      });

      await storage.save(configPath, registry);
      const loaded = await storage.load(configPath);
      const server = loaded.getServer('complete-server');

      expect(server?.enabled).toBe(false);
      expect(server?.source).toBe('curated');
      expect(server?.capabilities?.tools).toHaveLength(2);
      expect(server?.capabilities?.resources).toHaveLength(1);
      expect(server?.capabilities?.prompts).toHaveLength(1);
      expect(server?.settings.autoConnect).toBe(false);
      expect(server?.settings.timeout).toBe(45000);
      expect(server?.settings.fallbackServers).toEqual(['backup1', 'backup2']);
      expect(server?.status?.connected).toBe(true);
      expect(server?.status?.lastConnected).toBe(now);
    });
  });
});
