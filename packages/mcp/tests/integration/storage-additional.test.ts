import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { MCPStorage } from '../../src/storage.js';
import { MCPRegistry } from '../../src/registry.js';
import fs from 'fs';
import path from 'path';
import os from 'os';

describe('MCPStorage Additional Coverage', () => {
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

  describe('exists', () => {
    it('should return true when file exists', async () => {
      fs.writeFileSync(configPath, JSON.stringify({ version: 1, servers: [] }));
      const exists = await storage.exists(configPath);
      expect(exists).toBe(true);
    });

    it('should return false when file does not exist', async () => {
      const exists = await storage.exists(configPath);
      expect(exists).toBe(false);
    });
  });

  describe('save', () => {
    it('should create parent directories if they do not exist', async () => {
      const nestedPath = path.join(tempDir, 'nested', 'deep', 'mcp-servers.json');
      
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

      await storage.save(nestedPath, registry);
      expect(fs.existsSync(nestedPath)).toBe(true);
    });
  });

  describe('load', () => {
    it('should handle empty servers array', async () => {
      fs.writeFileSync(configPath, JSON.stringify({ version: 1, servers: [] }));
      const loaded = await storage.load(configPath);
      expect(loaded.size).toBe(0);
    });

    it('should handle JSON parse error', async () => {
      fs.writeFileSync(configPath, '{ invalid json');
      await expect(storage.load(configPath)).rejects.toThrow();
    });
  });
});
