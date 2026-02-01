import { describe, it, expect, beforeEach } from 'vitest';
import { ProviderRegistry, createRegistry } from '../../src/registry.js';
import {
  ProviderNotFoundError,
  ProviderAlreadyExistsError,
  ProviderValidationError,
} from '../../src/types.js';
import { OPENAI_PRESET, ANTHROPIC_PRESET } from '../../src/presets.js';

describe('ProviderRegistry', () => {
  let registry: ProviderRegistry;

  beforeEach(() => {
    registry = new ProviderRegistry();
  });

  describe('constructor', () => {
    it('should initialize with builtin presets', () => {
      expect(registry.size).toBeGreaterThan(0);
      expect(registry.hasProvider('openai')).toBe(true);
      expect(registry.hasProvider('anthropic')).toBe(true);
    });
  });

  describe('addProvider', () => {
    it('should add new provider', () => {
      const newProvider = {
        id: 'custom',
        name: 'Custom Provider',
        type: 'custom' as const,
        enabled: true,
        baseUrl: 'https://api.custom.com',
        apiKey: 'test-key',
        models: [{
          id: 'model-1',
          name: 'Model 1',
          maxTokens: 4096,
          supportsStreaming: true,
          supportsVision: false,
          supportsTools: false,
        }],
        timeout: 30000,
        retries: 3,
      };

      registry.addProvider(newProvider);

      expect(registry.hasProvider('custom')).toBe(true);
      expect(registry.getProvider('custom')?.name).toBe('Custom Provider');
    });

    it('should throw when provider already exists', () => {
      const provider = {
        id: 'openai',
        name: 'Duplicate',
        type: 'openai' as const,
        enabled: true,
        apiKey: 'key',
        models: [{
          id: 'model',
          name: 'Model',
          maxTokens: 1000,
          supportsStreaming: true,
          supportsVision: false,
          supportsTools: false,
        }],
        timeout: 30000,
        retries: 3,
      };

      expect(() => registry.addProvider(provider)).toThrow(ProviderAlreadyExistsError);
    });

    it('should throw on invalid config', () => {
      const invalidProvider = {
        id: 'invalid',
        name: '',
        type: 'openai' as const,
        enabled: true,
        apiKey: 'key',
        models: [],
        timeout: 30000,
        retries: 3,
      };

      expect(() => registry.addProvider(invalidProvider)).toThrow(ProviderValidationError);
    });
  });

  describe('updateProvider', () => {
    it('should update existing provider', () => {
      const updated = registry.updateProvider('openai', { name: 'Updated OpenAI' });

      expect(updated.name).toBe('Updated OpenAI');
      expect(registry.getProvider('openai')?.name).toBe('Updated OpenAI');
    });

    it('should throw when provider not found', () => {
      expect(() => registry.updateProvider('nonexistent', { name: 'Test' })).toThrow(ProviderNotFoundError);
    });
  });

  describe('removeProvider', () => {
    it('should remove provider', () => {
      registry.removeProvider('openai');

      expect(registry.hasProvider('openai')).toBe(false);
    });

    it('should throw when provider not found', () => {
      expect(() => registry.removeProvider('nonexistent')).toThrow(ProviderNotFoundError);
    });

    it('should clear default provider when removed', () => {
      registry.setDefaultProvider('openai');
      registry.removeProvider('openai');

      expect(registry.getDefaultProviderId()).toBeNull();
    });
  });

  describe('getProvider', () => {
    it('should return provider by id', () => {
      const provider = registry.getProvider('openai');

      expect(provider).toBeDefined();
      expect(provider?.id).toBe('openai');
    });

    it('should return undefined for nonexistent provider', () => {
      const provider = registry.getProvider('nonexistent');

      expect(provider).toBeUndefined();
    });
  });

  describe('getAllProviders', () => {
    it('should return all providers', () => {
      const providers = registry.getAllProviders();

      expect(providers.length).toBe(registry.size);
      expect(providers.some((p) => p.id === 'openai')).toBe(true);
    });
  });

  describe('getEnabledProviders', () => {
    it('should return only enabled providers', () => {
      registry.disableProvider('openai');
      const enabled = registry.getEnabledProviders();

      expect(enabled.some((p) => p.id === 'openai')).toBe(false);
      expect(enabled.every((p) => p.enabled)).toBe(true);
    });
  });

  describe('setDefaultProvider', () => {
    it('should set default provider', () => {
      registry.setDefaultProvider('anthropic');

      expect(registry.getDefaultProviderId()).toBe('anthropic');
    });

    it('should throw when provider not found', () => {
      expect(() => registry.setDefaultProvider('nonexistent')).toThrow(ProviderNotFoundError);
    });
  });

  describe('getDefaultProvider', () => {
    it('should return explicitly set default', () => {
      registry.setDefaultProvider('anthropic');
      const provider = registry.getDefaultProvider();

      expect(provider?.id).toBe('anthropic');
    });

    it('should return first enabled provider when no default set', () => {
      const provider = registry.getDefaultProvider();

      expect(provider).toBeDefined();
      expect(provider?.enabled).toBe(true);
    });

    it('should return undefined when no providers', () => {
      registry.clear();
      const provider = registry.getDefaultProvider();

      expect(provider).toBeUndefined();
    });
  });

  describe('enableProvider / disableProvider', () => {
    it('should disable provider', () => {
      registry.disableProvider('openai');

      expect(registry.getProvider('openai')?.enabled).toBe(false);
    });

    it('should enable provider', () => {
      registry.disableProvider('openai');
      registry.enableProvider('openai');

      expect(registry.getProvider('openai')?.enabled).toBe(true);
    });
  });

  describe('validateProvider', () => {
    it('should validate existing provider', () => {
      const result = registry.validateProvider('openai');

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should return error for nonexistent provider', () => {
      const result = registry.validateProvider('nonexistent');

      expect(result.valid).toBe(false);
      expect(result.errors[0]).toContain('not found');
    });
  });

  describe('testConnection', () => {
    it('should fail for nonexistent provider', async () => {
      const result = await registry.testConnection('nonexistent');

      expect(result.success).toBe(false);
      expect(result.error).toContain('not found');
    });

    it('should fail when api key not configured', async () => {
      const result = await registry.testConnection('openai');

      expect(result.success).toBe(false);
      expect(result.error).toContain('API key');
    });
  });

  describe('toJSON / fromJSON', () => {
    it('should serialize to JSON', () => {
      const json = registry.toJSON();

      expect(json.version).toBe(1);
      expect(json.providers.length).toBe(registry.size);
    });

    it('should deserialize from JSON', () => {
      const json = registry.toJSON();
      const newRegistry = new ProviderRegistry();
      newRegistry.clear();

      newRegistry.fromJSON(json);

      expect(newRegistry.size).toBe(registry.size);
      expect(newRegistry.hasProvider('openai')).toBe(true);
    });

    it('should throw on unsupported version', () => {
      const json = { version: 999, providers: [] };

      expect(() => registry.fromJSON(json)).toThrow(ProviderValidationError);
    });
  });

  describe('clear', () => {
    it('should clear all providers', () => {
      registry.clear();

      expect(registry.size).toBe(0);
      expect(registry.getDefaultProviderId()).toBeNull();
    });
  });
});

describe('createRegistry', () => {
  it('should create new registry', () => {
    const registry = createRegistry();

    expect(registry).toBeInstanceOf(ProviderRegistry);
    expect(registry.size).toBeGreaterThan(0);
  });
});
