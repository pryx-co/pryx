import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { mkdtemp, rm } from 'fs/promises';
import { tmpdir } from 'os';
import { join } from 'path';
import { ChannelStorage, createStorage } from '../../src/storage.js';
import { ChannelRegistry, createRegistry } from '../../src/registry.js';

const defaultSettings = {
  allowCommands: true,
  autoReply: false,
  filterPatterns: [],
  allowedUsers: [],
  blockedUsers: [],
};

describe('ChannelStorage', () => {
  let tempDir: string;
  let configPath: string;
  let storage: ChannelStorage;

  beforeEach(async () => {
    tempDir = await mkdtemp(join(tmpdir(), 'channels-test-'));
    configPath = join(tempDir, 'channels.json');
    storage = new ChannelStorage();
  });

  afterEach(async () => {
    await rm(tempDir, { recursive: true, force: true });
  });

  describe('save', () => {
    it('should save registry to file', async () => {
      const registry = createRegistry();
      registry.addChannel({
        id: 'telegram-bot',
        name: 'Telegram Bot',
        type: 'telegram',
        enabled: true,
        config: { botToken: 'test-token' },
        settings: defaultSettings,
      });

      await storage.save(configPath, registry);

      const exists = await storage.exists(configPath);
      expect(exists).toBe(true);
    });

    it('should create directory if not exists', async () => {
      const nestedPath = join(tempDir, 'nested', 'deep', 'channels.json');
      const registry = createRegistry();

      await storage.save(nestedPath, registry);

      const exists = await storage.exists(nestedPath);
      expect(exists).toBe(true);
    });
  });

  describe('load', () => {
    it('should load registry from file', async () => {
      const registry = createRegistry();
      registry.addChannel({
        id: 'telegram-bot',
        name: 'Telegram Bot',
        type: 'telegram',
        enabled: true,
        config: { botToken: 'test-token' },
        settings: defaultSettings,
      });
      await storage.save(configPath, registry);

      const loaded = await storage.load(configPath);

      expect(loaded.hasChannel('telegram-bot')).toBe(true);
    });

    it('should return new registry when file not found', async () => {
      const nonExistentPath = join(tempDir, 'nonexistent.json');

      const registry = await storage.load(nonExistentPath);

      expect(registry).toBeInstanceOf(ChannelRegistry);
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
    expect(storage).toBeInstanceOf(ChannelStorage);
  });
});
