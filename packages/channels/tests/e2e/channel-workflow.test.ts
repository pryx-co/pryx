import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { mkdtemp, rm } from 'fs/promises';
import { tmpdir } from 'os';
import { join } from 'path';
import { createRegistry, createStorage } from '../../src/index.js';

const defaultSettings = {
  allowCommands: true,
  autoReply: false,
  filterPatterns: [],
  allowedUsers: [],
  blockedUsers: [],
};

describe('Channel Workflow E2E', () => {
  let tempDir: string;
  let configPath: string;

  beforeEach(async () => {
    tempDir = await mkdtemp(join(tmpdir(), 'channels-e2e-test-'));
    configPath = join(tempDir, 'channels.json');
  });

  afterEach(async () => {
    await rm(tempDir, { recursive: true, force: true });
  });

  it('should configure Telegram bot', async () => {
    const registry = createRegistry();
    const storage = createStorage();

    registry.addChannel({
      id: 'telegram-bot',
      name: 'My Telegram Bot',
      type: 'telegram',
      enabled: true,
      config: {
        botToken: '123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11',
        chatId: '123456789',
        parseMode: 'Markdown' as const,
      },
      settings: defaultSettings,
    });

    const channel = registry.getChannel('telegram-bot');
    expect(channel?.type).toBe('telegram');
    expect(channel?.config.botToken).toBe('123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11');

    await storage.save(configPath, registry);
    expect(await storage.exists(configPath)).toBe(true);
  });

  it('should configure Discord bot', async () => {
    const registry = createRegistry();

    registry.addChannel({
      id: 'discord-bot',
      name: 'Discord Bot',
      type: 'discord',
      enabled: true,
      config: {
        botToken: 'discord-token',
        applicationId: '123456789',
        guildId: '987654321',
        intents: ['GUILDS', 'GUILD_MESSAGES'],
      },
      settings: defaultSettings,
    });

    expect(registry.hasChannel('discord-bot')).toBe(true);
  });

  it('should configure webhook endpoint', async () => {
    const registry = createRegistry();

    registry.addChannel({
      id: 'webhook-endpoint',
      name: 'Webhook Endpoint',
      type: 'webhook',
      enabled: true,
      config: {
        url: 'https://api.example.com/webhook',
        method: 'POST' as const,
        headers: {
          'Content-Type': 'application/json',
          'X-Custom-Header': 'value',
        },
      },
      settings: defaultSettings,
    });

    expect(registry.hasChannel('webhook-endpoint')).toBe(true);
  });

  it('should configure email channel', async () => {
    const registry = createRegistry();

    registry.addChannel({
      id: 'email-channel',
      name: 'Email Channel',
      type: 'email',
      enabled: true,
      config: {
        imap: {
          host: 'imap.gmail.com',
          port: 993,
          secure: true,
          username: 'user@gmail.com',
          password: 'app-password',
        },
        smtp: {
          host: 'smtp.gmail.com',
          port: 587,
          secure: true,
          username: 'user@gmail.com',
          password: 'app-password',
        },
        checkInterval: 60000,
        markAsRead: true,
      },
      settings: defaultSettings,
    });

    expect(registry.hasChannel('email-channel')).toBe(true);
  });

  it('should enable and disable channels', async () => {
    const registry = createRegistry();

    registry.addChannel({
      id: 'test-channel',
      name: 'Test Channel',
      type: 'telegram',
      enabled: true,
      config: { botToken: 'token' },
      settings: defaultSettings,
    });

    registry.disableChannel('test-channel');
    expect(registry.getChannel('test-channel')?.enabled).toBe(false);

    const enabledChannels = registry.getEnabledChannels();
    expect(enabledChannels.some((c) => c.id === 'test-channel')).toBe(false);

    registry.enableChannel('test-channel');
    expect(registry.getChannel('test-channel')?.enabled).toBe(true);
  });

  it('should persist configuration across restarts', async () => {
    const registry1 = createRegistry();
    const storage = createStorage();

    registry1.addChannel({
      id: 'persistent-channel',
      name: 'Persistent Channel',
      type: 'telegram',
      enabled: true,
      config: { botToken: 'persistent-token' },
      settings: defaultSettings,
    });

    await storage.save(configPath, registry1);

    const registry2 = await storage.load(configPath);

    expect(registry2.hasChannel('persistent-channel')).toBe(true);
    expect(registry2.getChannel('persistent-channel')?.config.botToken).toBe('persistent-token');
  });

  it('should filter channels by type', async () => {
    const registry = createRegistry();

    registry.addChannel({
      id: 'telegram-1',
      name: 'Telegram 1',
      type: 'telegram',
      enabled: true,
      config: { botToken: 'token1' },
      settings: defaultSettings,
    });
    registry.addChannel({
      id: 'discord-1',
      name: 'Discord 1',
      type: 'discord',
      enabled: true,
      config: { botToken: 'token2', applicationId: 'app' },
      settings: defaultSettings,
    });
    registry.addChannel({
      id: 'telegram-2',
      name: 'Telegram 2',
      type: 'telegram',
      enabled: true,
      config: { botToken: 'token3' },
      settings: defaultSettings,
    });

    const telegramChannels = registry.getChannelsByType('telegram');
    expect(telegramChannels.length).toBe(2);
  });

  it('should handle bulk operations', async () => {
    const registry = createRegistry();

    registry.addChannel({
      id: 'channel1',
      name: 'Channel 1',
      type: 'telegram',
      enabled: true,
      config: { botToken: 'token1' },
      settings: defaultSettings,
    });
    registry.addChannel({
      id: 'channel2',
      name: 'Channel 2',
      type: 'telegram',
      enabled: true,
      config: { botToken: 'token2' },
      settings: defaultSettings,
    });

    registry.disableAll();
    expect(registry.getAllChannels().every((c) => !c.enabled)).toBe(true);

    registry.enableAll();
    expect(registry.getAllChannels().every((c) => c.enabled)).toBe(true);
  });

  it('should update channel status', async () => {
    const registry = createRegistry();

    registry.addChannel({
      id: 'test',
      name: 'Test',
      type: 'telegram',
      enabled: true,
      config: { botToken: 'token' },
      settings: defaultSettings,
    });

    registry.updateChannelStatus('test', {
      connected: true,
      messageCount: 42,
    });

    const status = registry.getChannel('test')?.status;
    expect(status?.connected).toBe(true);
    expect(status?.messageCount).toBe(42);
  });
});
