import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { mkdtemp, rm } from 'fs/promises';
import { tmpdir } from 'os';
import { join } from 'path';
import { ProviderRegistry, createRegistry } from '../../src/registry.js';
import { ProviderStorage, createStorage } from '../../src/storage.js';

describe('Provider Registry Lifecycle', () => {
  let registry: ProviderRegistry;
  let storage: ProviderStorage;
  let tempDir: string;
  let configPath: string;

  beforeEach(async () => {
    registry = createRegistry();
    storage = createStorage();
    tempDir = await mkdtemp(join(tmpdir(), 'providers-lifecycle-test-'));
    configPath = join(tempDir, 'providers.json');
  });

  afterEach(async () => {
    await rm(tempDir, { recursive: true, force: true });
  });

  it('should complete full lifecycle: init → add → save → load → update → remove', async () => {
    const customProvider = {
      id: 'custom-ai',
      name: 'Custom AI Service',
      type: 'custom' as const,
      enabled: true,
      baseUrl: 'https://api.custom-ai.com/v1',
      apiKey: 'sk-custom-key',
      defaultModel: 'custom-model',
      models: [{
        id: 'custom-model',
        name: 'Custom Model',
        maxTokens: 8192,
        supportsStreaming: true,
        supportsVision: false,
        supportsTools: true,
      }],
      timeout: 30000,
      retries: 3,
    };

    registry.addProvider(customProvider);
    expect(registry.hasProvider('custom-ai')).toBe(true);

    await storage.save(configPath, registry);
    expect(await storage.exists(configPath)).toBe(true);

    const loadedRegistry = await storage.load(configPath);
    expect(loadedRegistry.hasProvider('custom-ai')).toBe(true);
    expect(loadedRegistry.getProvider('custom-ai')?.apiKey).toBe('sk-custom-key');

    loadedRegistry.updateProvider('custom-ai', { apiKey: 'sk-updated-key' });
    expect(loadedRegistry.getProvider('custom-ai')?.apiKey).toBe('sk-updated-key');

    loadedRegistry.removeProvider('custom-ai');
    expect(loadedRegistry.hasProvider('custom-ai')).toBe(false);
  });

  it('should handle multiple providers', async () => {
    const providers = [
      {
        id: 'provider-1',
        name: 'Provider 1',
        type: 'openai' as const,
        enabled: true,
        apiKey: 'key-1',
        models: [{ id: 'model-1', name: 'Model 1', maxTokens: 1000, supportsStreaming: true, supportsVision: false, supportsTools: false }],
        timeout: 30000,
        retries: 3,
      },
      {
        id: 'provider-2',
        name: 'Provider 2',
        type: 'anthropic' as const,
        enabled: true,
        apiKey: 'key-2',
        models: [{ id: 'model-2', name: 'Model 2', maxTokens: 2000, supportsStreaming: true, supportsVision: false, supportsTools: false }],
        timeout: 30000,
        retries: 3,
      },
    ];

    for (const provider of providers) {
      registry.addProvider(provider);
    }

    expect(registry.size).toBeGreaterThanOrEqual(7);

    await storage.save(configPath, registry);

    const loaded = await storage.load(configPath);
    expect(loaded.hasProvider('provider-1')).toBe(true);
    expect(loaded.hasProvider('provider-2')).toBe(true);
  });

  it('should persist default provider', async () => {
    registry.setDefaultProvider('anthropic');
    expect(registry.getDefaultProviderId()).toBe('anthropic');

    await storage.save(configPath, registry);

    const loaded = await storage.load(configPath);
    expect(loaded.getDefaultProviderId()).toBe('anthropic');
  });

  it('should handle enable/disable providers', async () => {
    registry.disableProvider('openai');
    expect(registry.getProvider('openai')?.enabled).toBe(false);

    await storage.save(configPath, registry);

    const loaded = await storage.load(configPath);
    expect(loaded.getProvider('openai')?.enabled).toBe(false);

    loaded.enableProvider('openai');
    expect(loaded.getProvider('openai')?.enabled).toBe(true);
  });

  it('should handle provider updates', async () => {
    registry.updateProvider('openai', {
      apiKey: 'sk-test',
      defaultModel: 'gpt-4o',
    });

    const provider = registry.getProvider('openai');
    expect(provider?.apiKey).toBe('sk-test');
    expect(provider?.defaultModel).toBe('gpt-4o');
  });

  it('should handle registry clear', async () => {
    registry.clear();
    expect(registry.size).toBe(0);
    expect(registry.getDefaultProviderId()).toBeNull();
  });
});
