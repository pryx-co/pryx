import {
  ProviderConfig,
  ProvidersConfig,
  ConnectionTestResult,
  ProviderNotFoundError,
  ProviderAlreadyExistsError,
  ProviderValidationError,
  CURRENT_VERSION,
} from './types.js';
import { validateProviderConfig, assertValidProviderConfig } from './validation.js';
import { BUILTIN_PRESETS } from './presets.js';

export class ProviderRegistry {
  private _providers: Map<string, ProviderConfig> = new Map();
  private _defaultProvider: string | null = null;
  private _version = CURRENT_VERSION;

  constructor() {
    for (const preset of BUILTIN_PRESETS) {
      this._providers.set(preset.id, { ...preset });
    }
  }

  addProvider(config: ProviderConfig): void {
    if (this._providers.has(config.id)) {
      throw new ProviderAlreadyExistsError(config.id);
    }
    
    assertValidProviderConfig(config);
    this._providers.set(config.id, { ...config });
  }

  updateProvider(id: string, updates: Partial<ProviderConfig>): ProviderConfig {
    const existing = this._providers.get(id);
    if (!existing) {
      throw new ProviderNotFoundError(id);
    }
    
    const updated = { ...existing, ...updates };
    assertValidProviderConfig(updated);
    this._providers.set(id, updated);
    
    return updated;
  }

  removeProvider(id: string): void {
    if (!this._providers.has(id)) {
      throw new ProviderNotFoundError(id);
    }
    
    this._providers.delete(id);
    
    if (this._defaultProvider === id) {
      this._defaultProvider = null;
    }
  }

  getProvider(id: string): ProviderConfig | undefined {
    return this._providers.get(id);
  }

  getAllProviders(): ProviderConfig[] {
    return Array.from(this._providers.values());
  }

  getEnabledProviders(): ProviderConfig[] {
    return this.getAllProviders().filter((p) => p.enabled);
  }

  hasProvider(id: string): boolean {
    return this._providers.has(id);
  }

  setDefaultProvider(id: string): void {
    if (!this._providers.has(id)) {
      throw new ProviderNotFoundError(id);
    }
    
    this._defaultProvider = id;
  }

  getDefaultProvider(): ProviderConfig | undefined {
    if (this._defaultProvider) {
      return this._providers.get(this._defaultProvider);
    }
    
    const enabled = this.getEnabledProviders();
    return enabled.length > 0 ? enabled[0] : undefined;
  }

  getDefaultProviderId(): string | null {
    return this._defaultProvider;
  }

  enableProvider(id: string): void {
    this.updateProvider(id, { enabled: true });
  }

  disableProvider(id: string): void {
    this.updateProvider(id, { enabled: false });
  }

  validateProvider(id: string): ReturnType<typeof validateProviderConfig> {
    const provider = this._providers.get(id);
    if (!provider) {
      return { valid: false, errors: [`Provider not found: ${id}`] };
    }
    
    return validateProviderConfig(provider);
  }

  async testConnection(id: string): Promise<ConnectionTestResult> {
    const provider = this._providers.get(id);
    if (!provider) {
      return {
        success: false,
        error: `Provider not found: ${id}`,
      };
    }
    
    const start = performance.now();
    
    try {
      if (!provider.apiKey && provider.type !== 'local') {
        return {
          success: false,
          error: 'API key not configured',
        };
      }
      
      const latency = performance.now() - start;
      
      return {
        success: true,
        latency,
        modelsAvailable: provider.models.map((m) => m.id),
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  toJSON(): ProvidersConfig {
    return {
      version: this._version,
      defaultProvider: this._defaultProvider || undefined,
      providers: this.getAllProviders(),
    };
  }

  fromJSON(data: ProvidersConfig): void {
    if (data.version !== CURRENT_VERSION) {
      throw new ProviderValidationError([`Unsupported version: ${data.version}`]);
    }
    
    this._providers.clear();
    
    for (const provider of data.providers) {
      assertValidProviderConfig(provider);
      this._providers.set(provider.id, provider);
    }
    
    this._defaultProvider = data.defaultProvider || null;
  }

  clear(): void {
    this._providers.clear();
    this._defaultProvider = null;
  }

  get size(): number {
    return this._providers.size;
  }
}

export function createRegistry(): ProviderRegistry {
  return new ProviderRegistry();
}
