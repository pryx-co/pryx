import { describe, it, expect } from 'vitest';
import {
  validateChannelConfig,
  assertValidChannelConfig,
  isValidChannelId,
  isValidChannelType,
  matchesFilterPatterns,
  isUserAllowed,
} from '../../src/validation.js';
import { ChannelValidationError } from '../../src/types.js';

describe('validateChannelConfig', () => {
  const baseConfig = {
    id: 'test-channel',
    name: 'Test Channel',
    type: 'telegram' as const,
    enabled: true,
    config: {
      botToken: 'test-token',
    },
  };

  it('should validate correct telegram config', () => {
    const result = validateChannelConfig(baseConfig);
    
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('should validate correct discord config', () => {
    const config = {
      ...baseConfig,
      type: 'discord' as const,
      config: {
        botToken: 'discord-token',
        applicationId: 'app-id',
      },
    };
    
    const result = validateChannelConfig(config);
    expect(result.valid).toBe(true);
  });

  it('should validate correct slack config', () => {
    const config = {
      ...baseConfig,
      type: 'slack' as const,
      config: {
        botToken: 'slack-token',
      },
    };
    
    const result = validateChannelConfig(config);
    expect(result.valid).toBe(true);
  });

  it('should validate correct email config with imap', () => {
    const config = {
      ...baseConfig,
      type: 'email' as const,
      config: {
        imap: {
          host: 'imap.example.com',
          port: 993,
          secure: true,
          username: 'user',
          password: 'pass',
        },
      },
    };
    
    const result = validateChannelConfig(config);
    expect(result.valid).toBe(true);
  });

  it('should validate correct email config with smtp', () => {
    const config = {
      ...baseConfig,
      type: 'email' as const,
      config: {
        smtp: {
          host: 'smtp.example.com',
          port: 587,
          secure: true,
          username: 'user',
          password: 'pass',
        },
      },
    };
    
    const result = validateChannelConfig(config);
    expect(result.valid).toBe(true);
  });



  it('should validate correct whatsapp config', () => {
    const config = {
      ...baseConfig,
      type: 'whatsapp' as const,
      config: {
        sessionName: 'my-session',
      },
    };
    
    const result = validateChannelConfig(config);
    expect(result.valid).toBe(true);
  });

  it('should validate correct webhook config', () => {
    const config = {
      ...baseConfig,
      type: 'webhook' as const,
      config: {
        url: 'https://api.example.com/webhook',
        method: 'POST' as const,
      },
    };
    
    const result = validateChannelConfig(config);
    expect(result.valid).toBe(true);
  });

  it('should reject config with invalid id format', () => {
    const config = { ...baseConfig, id: 'Invalid ID!' };
    const result = validateChannelConfig(config);
    
    expect(result.valid).toBe(false);
  });

  it('should reject config with empty name', () => {
    const config = { ...baseConfig, name: '' };
    const result = validateChannelConfig(config);
    
    expect(result.valid).toBe(false);
  });

  it('should reject config with invalid type', () => {
    const config = { ...baseConfig, type: 'invalid' };
    const result = validateChannelConfig(config);
    
    expect(result.valid).toBe(false);
  });



  it('should reject webhook settings with enabled but no url', () => {
    const config = {
      ...baseConfig,
      webhook: {
        enabled: true,
      },
    };
    
    const result = validateChannelConfig(config);
    expect(result.valid).toBe(false);
  });
});

describe('assertValidChannelConfig', () => {
  const validConfig = {
    id: 'test',
    name: 'Test',
    type: 'telegram' as const,
    enabled: true,
    config: { botToken: 'token' },
  };

  it('should return config when valid', () => {
    const result = assertValidChannelConfig(validConfig);
    expect(result.id).toBe('test');
  });

  it('should throw when invalid', () => {
    const config = { ...validConfig, id: '' };
    
    expect(() => assertValidChannelConfig(config)).toThrow(ChannelValidationError);
  });
});

describe('isValidChannelId', () => {
  it('should return true for valid ids', () => {
    expect(isValidChannelId('telegram-bot')).toBe(true);
    expect(isValidChannelId('discord-server')).toBe(true);
    expect(isValidChannelId('a')).toBe(true);
  });

  it('should return false for invalid ids', () => {
    expect(isValidChannelId('')).toBe(false);
    expect(isValidChannelId('Invalid ID')).toBe(false);
    expect(isValidChannelId('test@channel')).toBe(false);
    expect(isValidChannelId('a'.repeat(65))).toBe(false);
  });
});

describe('isValidChannelType', () => {
  it('should return true for valid types', () => {
    expect(isValidChannelType('telegram')).toBe(true);
    expect(isValidChannelType('discord')).toBe(true);
    expect(isValidChannelType('slack')).toBe(true);
    expect(isValidChannelType('email')).toBe(true);
    expect(isValidChannelType('whatsapp')).toBe(true);
    expect(isValidChannelType('webhook')).toBe(true);
  });

  it('should return false for invalid types', () => {
    expect(isValidChannelType('invalid')).toBe(false);
    expect(isValidChannelType('signal')).toBe(false);
    expect(isValidChannelType('')).toBe(false);
  });
});

describe('matchesFilterPatterns', () => {
  it('should return true when no patterns', () => {
    expect(matchesFilterPatterns('hello', [])).toBe(true);
  });

  it('should match regex patterns', () => {
    expect(matchesFilterPatterns('hello world', ['hello'])).toBe(true);
    expect(matchesFilterPatterns('hello world', ['^hello'])).toBe(true);
    expect(matchesFilterPatterns('HELLO world', ['hello'])).toBe(true);
  });

  it('should match any pattern', () => {
    expect(matchesFilterPatterns('test', ['hello', 'test', 'world'])).toBe(true);
  });

  it('should return false when no match', () => {
    expect(matchesFilterPatterns('goodbye', ['hello'])).toBe(false);
  });

  it('should handle invalid regex gracefully', () => {
    expect(matchesFilterPatterns('hello', ['[invalid'])).toBe(false);
  });
});

describe('isUserAllowed', () => {
  it('should allow user when no restrictions', () => {
    expect(isUserAllowed('user1', [], [])).toBe(true);
  });

  it('should block user in blocked list', () => {
    expect(isUserAllowed('user1', [], ['user1'])).toBe(false);
  });

  it('should allow user in allowed list', () => {
    expect(isUserAllowed('user1', ['user1'], [])).toBe(true);
  });

  it('should block user not in allowed list', () => {
    expect(isUserAllowed('user2', ['user1'], [])).toBe(false);
  });

  it('should prioritize blocked over allowed', () => {
    expect(isUserAllowed('user1', ['user1'], ['user1'])).toBe(false);
  });
});
