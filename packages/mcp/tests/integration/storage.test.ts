import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { MCPStorage } from '../../src/storage.js';
import { MCPRegistry } from '../../src/registry.js';
import { MCPServerConfig } from '../../src/types.js';
import fs from 'fs';
import path from 'path';
import os from 'os';

describe('MCPRegistryStorage Integration', () => {
  let tempDir: string;
  let storage: MCPStorage;
  let registry: MCPRegistry;
  let configPath: string;

  beforeEach(() => {
    tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'mcp-storage-test-'));
    configPath = path.join(tempDir, 'mcp-servers.json');
    storage = new MCPStorage();
    registry = new MCPRegistry();
  });

  afterEach(() => {
    fs.rmSync(tempDir, { recursive: true, force: true });
  });

  describe('save and load', () => {
    it('should save registry to file', async () => {
      const server: MCPServerConfig = {
        id: 'filesystem',
        name: 'Filesystem Server',
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

      registry.addServer(server);
      await storage.save(configPath, registry);

      expect(fs.existsSync(configPath)).toBe(true);

      const content = fs.readFileSync(configPath, 'utf-8');
      const parsed = JSON.parse(content);
      expect(parsed.version).toBe(1);
      expect(parsed.servers).toHaveLength(1);
      expect(parsed.servers[0].id).toBe('filesystem');
    });

    it('should load registry from file', async () => {
      const config = {
        version: 1,
        servers: [{
          id: 'test-server',
          name: 'Test Server',
          enabled: true,
          source: 'curated',
          transport: {
            type: 'sse',
            url: 'https://api.example.com/sse',
            headers: {},
          },
          settings: {
            autoConnect: true,
            timeout: 30000,
            reconnect: true,
            maxReconnectAttempts: 3,
            fallbackServers: [],
          },
        }],
      };

      fs.writeFileSync(configPath, JSON.stringify(config, null, 2));

      const loaded = await storage.load(configPath);
      expect(loaded.size).toBe(1);
      expect(loaded.hasServer('test-server')).toBe(true);
      expect(loaded.getServer('test-server')?.name).toBe('Test Server');
    });

    it('should handle non-existent config file', async () => {
      const loaded = await storage.load(configPath);
      expect(loaded.size).toBe(0);
    });

    it('should handle invalid JSON gracefully', async () => {
      fs.writeFileSync(configPath, 'invalid json {');

      await expect(storage.load(configPath)).rejects.toThrow();
    });

    it('should preserve all server properties on save/load', async () => {
      const server: MCPServerConfig = {
        id: 'complex-server',
        name: 'Complex Server',
        enabled: false,
        source: 'marketplace',
        transport: {
          type: 'websocket',
          url: 'wss://ws.example.com/mcp',
          headers: { Authorization: 'Bearer token123' },
        },
        capabilities: {
          tools: [
            {
              name: 'test-tool',
              description: 'A test tool',
              inputSchema: { type: 'object' },
            },
          ],
          resources: [
            {
              uri: 'file:///test',
              name: 'Test Resource',
              mimeType: 'text/plain',
            },
          ],
          prompts: [
            {
              name: 'test-prompt',
              description: 'A test prompt',
            },
          ],
        },
        settings: {
          autoConnect: false,
          timeout: 60000,
          reconnect: false,
          maxReconnectAttempts: 5,
          fallbackServers: ['fallback1', 'fallback2'],
        },
        status: {
          connected: true,
          lastConnected: new Date().toISOString(),
          lastError: undefined,
          reconnectAttempts: 0,
        },
      };

      registry.addServer(server);
      await storage.save(configPath, registry);

      const loaded = await storage.load(configPath);
      const loadedServer = loaded.getServer('complex-server');

      expect(loadedServer).toBeDefined();
      expect(loadedServer?.enabled).toBe(false);
      expect(loadedServer?.source).toBe('marketplace');
      expect(loadedServer?.capabilities?.tools).toHaveLength(1);
      expect(loadedServer?.capabilities?.resources).toHaveLength(1);
      expect(loadedServer?.settings.timeout).toBe(60000);
      expect(loadedServer?.status?.connected).toBe(true);
    });
  });

  describe('multiple servers', () => {
    it('should handle multiple servers', async () => {
      const servers: MCPServerConfig[] = [
        {
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
        },
        {
          id: 'server2',
          name: 'Server 2',
          enabled: true,
          source: 'curated',
          transport: { type: 'sse', url: 'https://example.com', headers: {} },
          settings: {
            autoConnect: true,
            timeout: 30000,
            reconnect: true,
            maxReconnectAttempts: 3,
            fallbackServers: [],
          },
        },
        {
          id: 'server3',
          name: 'Server 3',
          enabled: false,
          source: 'manual',
          transport: { type: 'websocket', url: 'wss://example.com', headers: {} },
          settings: {
            autoConnect: true,
            timeout: 30000,
            reconnect: true,
            maxReconnectAttempts: 3,
            fallbackServers: [],
          },
        },
      ];

      servers.forEach(s => { registry.addServer(s); });
      await storage.save(configPath, registry);

      const loaded = await storage.load(configPath);
      expect(loaded.size).toBe(3);
      expect(loaded.getEnabledServers()).toHaveLength(2);
    });
  });

  describe('file permissions', () => {
    it('should create file with secure permissions', async () => {
      const server: MCPServerConfig = {
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
      };

      registry.addServer(server);
      await storage.save(configPath, registry);

      const stats = fs.statSync(configPath);
      
      const mode = stats.mode & 0o777;
      expect(mode).toBe(0o600);
    });
  });
});
