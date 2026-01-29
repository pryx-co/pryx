import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { mkdtemp, rm } from 'fs/promises';
import { tmpdir } from 'os';
import { join } from 'path';
import { ProviderStorage, createStorage } from '../../src/storage.js';
import { ProviderRegistry, createRegistry } from '../../src/registry.js';

describe('ProviderStorage', () => {
  let tempDir: string;
  let configPath: string;
  let storage: ProviderStorage;

  beforeEach(async () => {
    tempDir = await mkdtemp(join(tmpdir(), 'providers-test-'));
    configPath = join(tempDir, 'providers.json');
    storage = new ProviderStorage();
  });

  afterEach(async () => {
    await rm(tempDir, { recursive: true, force: true });
  });

  describe('save', () => {
    it('should save registry to file', async () => {
      const registry = createRegistry();
      registry.updateProvider('openai', { apiKey: 'test-key' });

      await storage.save(configPath, registry);

      const exists = await storage.exists(configPath);
      expect(exists).toBe(true);
    });

    it('should create directory if not exists', async () => {
      const nestedPath = join(tempDir, 'nested', 'deep', 'providers.json');
      const registry = createRegistry();

      await storage.save(nestedPath, registry);

      const exists = await storage.exists(nestedPath);
      expect(exists).toBe(true);
    });
  });

  describe('load', () => {
    it('should load registry from file', async () => {
      const registry = createRegistry();
      registry.updateProvider('openai', { apiKey: 'test-key' });
      await storage.save(configPath, registry);

      const loaded = await storage.load(configPath);

      expect(loaded.hasProvider('openai')).toBe(true);
      expect(loaded.getProvider('openai')?.apiKey).toBe('test-key');
    });

    it('should return new registry when file not found', async () => {
      const nonExistentPath = join(tempDir, 'nonexistent.json');

      const registry = await storage.load(nonExistentPath);

      expect(registry).toBeInstanceOf(ProviderRegistry);
      expect(registry.size).toBeGreaterThan(0);
    });

    it('should throw on invalid JSON', async () => {
      const invalidPath = join(tempDir, 'invalid.json');
      await storage.save(invalidPath, createRegistry());
      
      const data = await import('fs/promises').then((m) => m.readFile(invalidPath, 'utf8'));
      await import('fs/promises').then((m) => m.writeFile(invalidPath, 'invalid json'));

      await expect(storage.load(invalidPath)).rejects.toThrow();
    });
  });

  describe('exists', () => {
    it('should return true when file exists', async () => {
      await storage.save(configPath, createRegistry());

      const exists = await storage.exists(configPath);
      expect(exists).toBe(true);
    });

    it('should return false when file does not exist', async () => {
      const exists = await storage.exists(join(tempDir, 'nonexistent.json'));
      expect(exists).toBe(false);
    });
  });
});

describe('createStorage', () => {
  it('should create new storage instance', () => {
    const storage = createStorage();
    expect(storage).toBeInstanceOf(ProviderStorage);
  });
});
