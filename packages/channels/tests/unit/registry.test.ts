import { describe, it, expect, beforeEach } from 'vitest';
import { ChannelRegistry, createRegistry } from '../../src/registry.js';
import {
  ChannelNotFoundError,
  ChannelAlreadyExistsError,
  ChannelValidationError,
} from '../../src/types.js';

describe('ChannelRegistry', () => {
  let registry: ChannelRegistry;

  beforeEach(() => {
    registry = new ChannelRegistry();
  });

  describe('constructor', () => {
    it('should create empty registry', () => {
      expect(registry.size).toBe(0);
    });
  });

  describe('addChannel', () => {
    it('should add telegram channel', () => {
      const channel = {
        id: 'telegram-bot',
        name: 'My Telegram Bot',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: {
          botToken: 'test-token',
          chatId: '123456',
        },
      };

      registry.addChannel(channel);

      expect(registry.hasChannel('telegram-bot')).toBe(true);
      expect(registry.getChannel('telegram-bot')?.name).toBe('My Telegram Bot');
    });

    it('should add discord channel', () => {
      const channel = {
        id: 'discord-bot',
        name: 'Discord Bot',
        type: 'discord' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: {
          botToken: 'discord-token',
          applicationId: 'app-id',
          intents: ['GUILDS', 'GUILD_MESSAGES'],
        },
      };

      registry.addChannel(channel);

      expect(registry.hasChannel('discord-bot')).toBe(true);
    });

    it('should add webhook channel', () => {
      const channel = {
        id: 'webhook-endpoint',
        name: 'Webhook Endpoint',
        type: 'webhook' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: {
          url: 'https://api.example.com/webhook',
          method: 'POST' as const,
          headers: {
            'Content-Type': 'application/json',
          },
        },
      };

      registry.addChannel(channel);

      expect(registry.hasChannel('webhook-endpoint')).toBe(true);
    });

    it('should throw when channel already exists', () => {
      const channel = {
        id: 'test',
        name: 'Test',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      };

      registry.addChannel(channel);
      
      expect(() => registry.addChannel(channel)).toThrow(ChannelAlreadyExistsError);
    });

    it('should throw on invalid config', () => {
      const channel = {
        id: 'invalid',
        name: '',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: {},
      };

      expect(() => registry.addChannel(channel)).toThrow(ChannelValidationError);
    });
  });

  describe('updateChannel', () => {
    it('should update existing channel', () => {
      registry.addChannel({
        id: 'test',
        name: 'Test',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      const updated = registry.updateChannel('test', { name: 'Updated' });

      expect(updated.name).toBe('Updated');
      expect(registry.getChannel('test')?.name).toBe('Updated');
    });

    it('should throw when channel not found', () => {
      expect(() => registry.updateChannel('nonexistent', { name: 'Test' })).toThrow(ChannelNotFoundError);
    });
  });

  describe('removeChannel', () => {
    it('should remove channel', () => {
      registry.addChannel({
        id: 'test',
        name: 'Test',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      registry.removeChannel('test');

      expect(registry.hasChannel('test')).toBe(false);
    });

    it('should throw when channel not found', () => {
      expect(() => registry.removeChannel('nonexistent')).toThrow(ChannelNotFoundError);
    });
  });

  describe('getChannel', () => {
    it('should return channel by id', () => {
      registry.addChannel({
        id: 'test',
        name: 'Test',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      const channel = registry.getChannel('test');

      expect(channel).toBeDefined();
      expect(channel?.id).toBe('test');
    });

    it('should return undefined for nonexistent channel', () => {
      const channel = registry.getChannel('nonexistent');
      expect(channel).toBeUndefined();
    });
  });

  describe('getAllChannels', () => {
    it('should return all channels', () => {
      registry.addChannel({
        id: 'channel1',
        name: 'Channel 1',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token1' },
      });
      registry.addChannel({
        id: 'channel2',
        name: 'Channel 2',
        type: 'discord' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token2', applicationId: 'app' },
      });

      const channels = registry.getAllChannels();

      expect(channels.length).toBe(2);
    });
  });

  describe('getEnabledChannels', () => {
    it('should return only enabled channels', () => {
      registry.addChannel({
        id: 'enabled',
        name: 'Enabled',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });
      registry.addChannel({
        id: 'disabled',
        name: 'Disabled',
        type: 'telegram' as const,
        enabled: false, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      const enabled = registry.getEnabledChannels();

      expect(enabled.length).toBe(1);
      expect(enabled[0].id).toBe('enabled');
    });
  });

  describe('getChannelsByType', () => {
    it('should return channels by type', () => {
      registry.addChannel({
        id: 'telegram1',
        name: 'Telegram 1',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });
      registry.addChannel({
        id: 'discord1',
        name: 'Discord 1',
        type: 'discord' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token', applicationId: 'app' },
      });
      registry.addChannel({
        id: 'telegram2',
        name: 'Telegram 2',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      const telegramChannels = registry.getChannelsByType('telegram');

      expect(telegramChannels.length).toBe(2);
      expect(telegramChannels.every((c) => c.type === 'telegram')).toBe(true);
    });
  });

  describe('enableChannel / disableChannel', () => {
    it('should disable channel', () => {
      registry.addChannel({
        id: 'test',
        name: 'Test',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      registry.disableChannel('test');

      expect(registry.getChannel('test')?.enabled).toBe(false);
    });

    it('should enable channel', () => {
      registry.addChannel({
        id: 'test',
        name: 'Test',
        type: 'telegram' as const,
        enabled: false, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      registry.enableChannel('test');

      expect(registry.getChannel('test')?.enabled).toBe(true);
    });
  });

  describe('enableAll / disableAll', () => {
    it('should enable all channels', () => {
      registry.addChannel({
        id: 'test1',
        name: 'Test 1',
        type: 'telegram' as const,
        enabled: false, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });
      registry.addChannel({
        id: 'test2',
        name: 'Test 2',
        type: 'telegram' as const,
        enabled: false, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      registry.enableAll();

      expect(registry.getAllChannels().every((c) => c.enabled)).toBe(true);
    });

    it('should disable all channels', () => {
      registry.addChannel({
        id: 'test1',
        name: 'Test 1',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });
      registry.addChannel({
        id: 'test2',
        name: 'Test 2',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      registry.disableAll();

      expect(registry.getAllChannels().every((c) => !c.enabled)).toBe(true);
    });
  });

  describe('enableType / disableType', () => {
    it('should enable channels by type', () => {
      registry.addChannel({
        id: 'telegram1',
        name: 'Telegram 1',
        type: 'telegram' as const,
        enabled: false, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });
      registry.addChannel({
        id: 'discord1',
        name: 'Discord 1',
        type: 'discord' as const,
        enabled: false, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token', applicationId: 'app' },
      });

      registry.enableType('telegram');

      expect(registry.getChannel('telegram1')?.enabled).toBe(true);
      expect(registry.getChannel('discord1')?.enabled).toBe(false);
    });
  });

  describe('updateChannelStatus', () => {
    it('should update channel status', () => {
      registry.addChannel({
        id: 'test',
        name: 'Test',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      registry.updateChannelStatus('test', {
        connected: true,
        messageCount: 10,
      });

      const status = registry.getChannel('test')?.status;
      expect(status?.connected).toBe(true);
      expect(status?.messageCount).toBe(10);
    });

    it('should throw when channel not found', () => {
      expect(() => registry.updateChannelStatus('nonexistent', { connected: true })).toThrow(ChannelNotFoundError);
    });
  });

  describe('validateChannel', () => {
    it('should validate existing channel', () => {
      registry.addChannel({
        id: 'test',
        name: 'Test',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      const result = registry.validateChannel('test');

      expect(result.valid).toBe(true);
    });

    it('should return error for nonexistent channel', () => {
      const result = registry.validateChannel('nonexistent');

      expect(result.valid).toBe(false);
      expect(result.errors[0]).toContain('not found');
    });
  });

  describe('testConnection', () => {
    it('should fail for nonexistent channel', async () => {
      const result = await registry.testConnection('nonexistent');

      expect(result.success).toBe(false);
      expect(result.error).toContain('not found');
    });

    it('should return success for valid channel', async () => {
      registry.addChannel({
        id: 'test',
        name: 'Test',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      const result = await registry.testConnection('test');

      expect(result.success).toBe(true);
      expect(result.latency).toBeDefined();
    });
  });

  describe('toJSON / fromJSON', () => {
    it('should serialize to JSON', () => {
      registry.addChannel({
        id: 'test',
        name: 'Test',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      const json = registry.toJSON();

      expect(json.version).toBe(1);
      expect(json.channels.length).toBe(1);
    });

    it('should deserialize from JSON', () => {
      const json = {
        version: 1,
        channels: [{
          id: 'test',
          name: 'Test',
          type: 'telegram',
          enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
          config: { botToken: 'token' },
        }],
      };

      registry.fromJSON(json);

      expect(registry.size).toBe(1);
      expect(registry.hasChannel('test')).toBe(true);
    });

    it('should throw on unsupported version', () => {
      const json = { version: 999, channels: [] };

      expect(() => registry.fromJSON(json)).toThrow(ChannelValidationError);
    });
  });

  describe('clear', () => {
    it('should clear all channels', () => {
      registry.addChannel({
        id: 'test',
        name: 'Test',
        type: 'telegram' as const,
        enabled: true, settings: { allowCommands: true, autoReply: false, filterPatterns: [], allowedUsers: [], blockedUsers: [] },
        config: { botToken: 'token' },
      });

      registry.clear();

      expect(registry.size).toBe(0);
    });
  });
});

describe('createRegistry', () => {
  it('should create new registry', () => {
    const registry = createRegistry();

    expect(registry).toBeInstanceOf(ChannelRegistry);
    expect(registry.size).toBe(0);
  });
});
