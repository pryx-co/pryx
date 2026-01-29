import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { mkdtemp, rm } from 'fs/promises';
import { tmpdir } from 'os';
import { join } from 'path';
import { createRegistry, createStorage } from '../../src/index.js';

describe('Provider Workflow E2E', () => {
  let tempDir: string;
  let configPath: string;

  beforeEach(async () => {
    tempDir = await mkdtemp(join(tmpdir(), 'providers-e2e-test-'));
    configPath = join(tempDir, 'providers.json');
  });

  afterEach(async () => {
    await rm(tempDir, { recursive: true, force: true });
  });

  it('should add and configure OpenAI provider', async () => {
    const registry = createRegistry();
    const storage = createStorage();

    registry.updateProvider('openai', {
      apiKey: 'sk-openai-test-key',
      enabled: true,
    });

    const provider = registry.getProvider('openai');
    expect(provider?.apiKey).toBe('sk-openai-test-key');
    expect(provider?.enabled).toBe(true);
    expect(provider?.models.some((m) => m.id === 'gpt-4o')).toBe(true);

    await storage.save(configPath, registry);
    expect(await storage.exists(configPath)).toBe(true);
  });

  it('should add and configure Anthropic provider', async () => {
    const registry = createRegistry();
    const storage = createStorage();

    registry.updateProvider('anthropic', {
      apiKey: 'sk-ant-test-key',
      defaultModel: 'claude-3-opus-20240229',
    });

    const provider = registry.getProvider('anthropic');
    expect(provider?.apiKey).toBe('sk-ant-test-key');
    expect(provider?.defaultModel).toBe('claude-3-opus-20240229');

    await storage.save(configPath, registry);
  });

  it('should add custom provider', async () => {
    const registry = createRegistry();

    const customProvider = {
      id: 'my-custom-provider',
      name: 'My Custom AI',
      type: 'custom' as const,
      enabled: true,
      baseUrl: 'https://api.my-ai.com/v1',
      apiKey: 'my-api-key',
      defaultModel: 'custom-v1',
      models: [{
        id: 'custom-v1',
        name: 'Custom V1',
        maxTokens: 4096,
        supportsStreaming: true,
        supportsVision: false,
        supportsTools: true,
      }],
      timeout: 30000,
      retries: 3,
    };

    registry.addProvider(customProvider);

    expect(registry.hasProvider('my-custom-provider')).toBe(true);
    expect(registry.getProvider('my-custom-provider')?.baseUrl).toBe('https://api.my-ai.com/v1');
  });

  it('should switch between providers', async () => {
    const registry = createRegistry();

    registry.setDefaultProvider('openai');
    expect(registry.getDefaultProvider()?.id).toBe('openai');

    registry.setDefaultProvider('anthropic');
    expect(registry.getDefaultProvider()?.id).toBe('anthropic');
  });

  it('should enable and disable providers', async () => {
    const registry = createRegistry();

    registry.disableProvider('openai');
    expect(registry.getProvider('openai')?.enabled).toBe(false);

    const enabledProviders = registry.getEnabledProviders();
    expect(enabledProviders.some((p) => p.id === 'openai')).toBe(false);

    registry.enableProvider('openai');
    expect(registry.getProvider('openai')?.enabled).toBe(true);
  });

  it('should persist configuration across restarts', async () => {
    const registry1 = createRegistry();
    const storage = createStorage();

    registry1.updateProvider('openai', { apiKey: 'persistent-key' });
    registry1.setDefaultProvider('anthropic');
    registry1.disableProvider('google');

    await storage.save(configPath, registry1);

    const registry2 = await storage.load(configPath);

    expect(registry2.getProvider('openai')?.apiKey).toBe('persistent-key');
    expect(registry2.getDefaultProviderId()).toBe('anthropic');
    expect(registry2.getProvider('google')?.enabled).toBe(false);
  });

  it('should validate provider configuration', async () => {
    const registry = createRegistry();

    const result = registry.validateProvider('openai');
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('should test connection to provider', async () => {
    const registry = createRegistry();

    const result = await registry.testConnection('openai');
    expect(result.success).toBe(false);
    expect(result.error).toContain('API key');
  });

  it('should handle provider removal', async () => {
    const registry = createRegistry();

    registry.setDefaultProvider('openai');
    registry.removeProvider('openai');

    expect(registry.hasProvider('openai')).toBe(false);
    expect(registry.getDefaultProviderId()).toBeNull();
  });

  it('should handle multiple provider operations', async () => {
    const registry = createRegistry();

    registry.updateProvider('openai', { apiKey: 'openai-key' });
    registry.updateProvider('anthropic', { apiKey: 'anthropic-key' });
    registry.setDefaultProvider('anthropic');
    registry.disableProvider('google');

    const allProviders = registry.getAllProviders();
    const enabledProviders = registry.getEnabledProviders();

    expect(allProviders.length).toBeGreaterThanOrEqual(5);
    expect(enabledProviders.some((p) => p.id === 'google')).toBe(false);
    expect(registry.getDefaultProvider()?.id).toBe('anthropic');
  });
});
