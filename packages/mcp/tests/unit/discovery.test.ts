import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import {
  MCPServerDiscovery,
  createServerDiscovery,
  CuratedServer,
} from '../../src/discovery.js';
import { writeFileSync, mkdtempSync, rmSync } from 'fs';
import { join } from 'path';
import { tmpdir } from 'os';

describe('MCPServerDiscovery', () => {
  let discovery: MCPServerDiscovery;
  let tempDir: string;
  let dbPath: string;

  const mockDatabase = {
    version: 1,
    lastUpdated: '2026-01-30',
    categories: [
      { id: 'filesystem', name: 'File System', description: 'File operations' },
      { id: 'web', name: 'Web', description: 'Web tools' },
    ],
    servers: [
      {
        id: 'filesystem',
        name: 'File System',
        description: 'Access local files',
        category: 'filesystem',
        author: 'Anthropic',
        repository: 'https://github.com/test',
        transport: { type: 'stdio', command: 'npx', args: ['-y', '@test/filesystem'] },
        capabilities: { tools: [{ name: 'read', description: 'Read file' }] },
      },
      {
        id: 'github',
        name: 'GitHub',
        description: 'GitHub integration',
        category: 'web',
        author: 'Anthropic',
        repository: 'https://github.com/test',
        transport: { type: 'stdio', command: 'npx', args: ['-y', '@test/github'] },
        capabilities: { tools: [{ name: 'search', description: 'Search repos' }] },
      },
      {
        id: 'brave-search',
        name: 'Brave Search',
        description: 'Web search',
        category: 'web',
        author: 'Community',
        repository: 'https://github.com/test',
        transport: { type: 'stdio', command: 'npx', args: ['-y', '@test/brave'] },
        capabilities: { tools: [{ name: 'search', description: 'Search web' }] },
      },
    ] as CuratedServer[],
  };

  beforeEach(() => {
    tempDir = mkdtempSync(join(tmpdir(), 'mcp-test-'));
    dbPath = join(tempDir, 'test-servers.json');
    writeFileSync(dbPath, JSON.stringify(mockDatabase, null, 2));
    discovery = createServerDiscovery(dbPath);
  });

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true });
  });

  describe('loadDatabase', () => {
    it('should load database from file', async () => {
      await discovery.loadDatabase();
      const categories = discovery.getCategories();
      expect(categories).toHaveLength(2);
    });

    it('should throw on missing file', async () => {
      const badDiscovery = createServerDiscovery('/nonexistent/path.json');
      await expect(badDiscovery.loadDatabase()).rejects.toThrow();
    });
  });

  describe('search', () => {
    beforeEach(async () => {
      await discovery.loadDatabase();
    });

    it('should return all servers when no filters', async () => {
      const results = await discovery.search();
      expect(results).toHaveLength(3);
    });

    it('should filter by category', async () => {
      const results = await discovery.search({ category: 'web' });
      expect(results).toHaveLength(2);
      expect(results.every(s => s.category === 'web')).toBe(true);
    });

    it('should filter by query', async () => {
      const results = await discovery.search({ query: 'brave' });
      expect(results).toHaveLength(1);
      expect(results[0].id).toBe('brave-search');
    });

    it('should filter by author', async () => {
      const results = await discovery.search({ author: 'community' });
      expect(results).toHaveLength(1);
      expect(results[0].id).toBe('brave-search');
    });

    it('should combine filters', async () => {
      const results = await discovery.search({ 
        category: 'web', 
        author: 'Anthropic' 
      });
      expect(results).toHaveLength(1);
      expect(results[0].id).toBe('github');
    });
  });

  describe('getServerById', () => {
    beforeEach(async () => {
      await discovery.loadDatabase();
    });

    it('should find server by id', () => {
      const server = discovery.getServerById('filesystem');
      expect(server).toBeDefined();
      expect(server?.name).toBe('File System');
    });

    it('should return undefined for unknown id', () => {
      const server = discovery.getServerById('nonexistent');
      expect(server).toBeUndefined();
    });
  });

  describe('getServersByCategory', () => {
    beforeEach(async () => {
      await discovery.loadDatabase();
    });

    it('should return servers in category', () => {
      const servers = discovery.getServersByCategory('web');
      expect(servers).toHaveLength(2);
    });

    it('should return empty array for unknown category', () => {
      const servers = discovery.getServersByCategory('unknown');
      expect(servers).toHaveLength(0);
    });
  });

  describe('validateCustomUrl', () => {
    it('should validate HTTPS URL', async () => {
      const result = await discovery.validateCustomUrl('https://api.example.com/mcp');
      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject invalid protocol', async () => {
      const result = await discovery.validateCustomUrl('ftp://example.com');
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('URL must use HTTP or HTTPS protocol');
    });

    it('should warn about localhost', async () => {
      const result = await discovery.validateCustomUrl('http://localhost:3000');
      expect(result.valid).toBe(true);
      expect(result.warnings.length).toBeGreaterThan(0);
    });

    it('should warn about insecure HTTP', async () => {
      const result = await discovery.validateCustomUrl('http://api.example.com');
      expect(result.valid).toBe(true);
      expect(result.warnings.some(w => w.includes('insecure'))).toBe(true);
    });

    it('should reject invalid URL format', async () => {
      const result = await discovery.validateCustomUrl('not-a-url');
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Invalid URL format');
    });
  });

  describe('validateServerId', () => {
    it('should validate correct server id', () => {
      const result = discovery.validateServerId('valid-server_id123');
      expect(result.valid).toBe(true);
    });

    it('should reject empty id', () => {
      const result = discovery.validateServerId('');
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Server ID is required');
    });

    it('should reject invalid characters', () => {
      const result = discovery.validateServerId('Invalid Server!');
      expect(result.valid).toBe(false);
    });

    it('should reject too long id', () => {
      const result = discovery.validateServerId('a'.repeat(65));
      expect(result.valid).toBe(false);
      expect(result.errors.some(e => e.includes('64'))).toBe(true);
    });
  });

  describe('toMCPServerConfig', () => {
    beforeEach(async () => {
      await discovery.loadDatabase();
    });

    it('should convert curated server to config', () => {
      const server = discovery.getServerById('filesystem')!;
      const config = discovery.toMCPServerConfig(server);

      expect(config.id).toBe('filesystem');
      expect(config.name).toBe('File System');
      expect(config.enabled).toBe(true);
      expect(config.source).toBe('curated');
      expect(config.transport.type).toBe('stdio');
    });

    it('should add custom args', () => {
      const server = discovery.getServerById('filesystem')!;
      const config = discovery.toMCPServerConfig(server, ['/home/user']);

      expect(config.transport.type).toBe('stdio');
      if (config.transport.type === 'stdio') {
        expect(config.transport.args).toContain('/home/user');
      }
    });
  });

  describe('getStats', () => {
    beforeEach(async () => {
      await discovery.loadDatabase();
    });

    it('should return database stats', () => {
      const stats = discovery.getStats();

      expect(stats.totalServers).toBe(3);
      expect(stats.totalCategories).toBe(2);
      expect(stats.serversByCategory['filesystem']).toBe(1);
      expect(stats.serversByCategory['web']).toBe(2);
    });
  });
});

describe('createServerDiscovery', () => {
  it('should create discovery with custom path', () => {
    const discovery = createServerDiscovery('/custom/path.json');
    expect(discovery).toBeInstanceOf(MCPServerDiscovery);
  });

  it('should create discovery with default path', () => {
    const discovery = createServerDiscovery();
    expect(discovery).toBeInstanceOf(MCPServerDiscovery);
  });
});
