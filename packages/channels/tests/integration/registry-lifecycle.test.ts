import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { mkdtemp, rm } from 'fs/promises';
import { tmpdir } from 'os';
import { join } from 'path';
import { ChannelRegistry, createRegistry } from '../../src/registry.js';
import { ChannelStorage, createStorage } from '../../src/storage.js';

const defaultSettings = {
  allowCommands: true,
  autoReply: false,
  filterPatterns: [],
  allowedUsers: [],
  blockedUsers: [],
};

describe('Channel Registry Lifecycle', () => {
  let registry: ChannelRegistry;
  let storage: ChannelStorage;
  let tempDir: string;
  let configPath: string;

  beforeEach(async () => {
    registry = createRegistry();
    storage = createStorage();
    tempDir = await mkdtemp(join(tmpdir(), 'channels-lifecycle-test-'));
    configPath = join(tempDir, 'channels.json');
  });

  afterEach(async () => {
    await rm(tempDir, { recursive: true, force: true });
  });

  it('should complete full lifecycle: init → add → save → load → update → remove', async () => {
    const channel = {
      id: 'telegram-bot',
      name: 'Telegram Bot',
      type: 'telegram' as const,
      enabled: true,
      config: { botToken: 'test-token' },
      settings: defaultSettings,
    };

    registry.addChannel(channel);
    expect(registry.hasChannel('telegram-bot')).toBe(true);

    await storage.save(configPath, registry);
    expect(await storage.exists(configPath)).toBe(true);

    const loaded = await storage.load(configPath);
    expect(loaded.hasChannel('telegram-bot')).toBe(true);

    loaded.updateChannel('telegram-bot', { name: 'Updated Bot' });
    expect(loaded.getChannel('telegram-bot')?.name).toBe('Updated Bot');

    loaded.removeChannel('telegram-bot');
    expect(loaded.hasChannel('telegram-bot')).toBe(false);
  });

  it('should handle multiple channels', async () => {
    const channels = [
      {
        id: 'telegram-1',
        name: 'Telegram 1',
        type: 'telegram' as const,
        enabled: true,
        config: { botToken: 'token1' },
        settings: defaultSettings,
      },
      {
        id: 'discord-1',
        name: 'Discord 1',
        type: 'discord' as const,
        enabled: true,
        config: { botToken: 'token2', applicationId: 'app-id' },
        settings: defaultSettings,
      },
    ];

    for (const channel of channels) {
      registry.addChannel(channel);
    }

    expect(registry.size).toBe(2);

    await storage.save(configPath, registry);

    const loaded = await storage.load(configPath);
    expect(loaded.hasChannel('telegram-1')).toBe(true);
    expect(loaded.hasChannel('discord-1')).toBe(true);
  });

  it('should handle enable/disable channels', async () => {
    registry.addChannel({
      id: 'test',
      name: 'Test',
      type: 'telegram',
      enabled: true,
      config: { botToken: 'token' },
      settings: defaultSettings,
    });

    registry.disableChannel('test');
    expect(registry.getChannel('test')?.enabled).toBe(false);

    await storage.save(configPath, registry);

    const loaded = await storage.load(configPath);
    expect(loaded.getChannel('test')?.enabled).toBe(false);

    loaded.enableChannel('test');
    expect(loaded.getChannel('test')?.enabled).toBe(true);
  });
});
