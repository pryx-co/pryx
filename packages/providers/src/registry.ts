/**
 * Provider Registry Module
 *
 * Manages a collection of AI provider configurations with CRUD operations,
 * validation, connection testing, and JSON serialization support.
 */

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

/**
 * Registry for managing AI provider configurations
 */
export class ProviderRegistry {
  private _providers: Map<string, ProviderConfig> = new Map();
  private _defaultProvider: string | null = null;
  private _version = CURRENT_VERSION;

  /**
   * Creates a new ProviderRegistry with built-in presets loaded
   */
  constructor() {
    for (const preset of BUILTIN_PRESETS) {
      this._providers.set(preset.id, { ...preset });
    }
  }

  /**
   * Adds a new provider to the registry
   * @param config - The provider configuration to add
   * @throws ProviderAlreadyExistsError if provider with same ID exists
   */
  addProvider(config: ProviderConfig): void {
    if (this._providers.has(config.id)) {
      throw new ProviderAlreadyExistsError(config.id);
    }
    
    assertValidProviderConfig(config);
    this._providers.set(config.id, { ...config });
  }

  /**
   * Updates an existing provider configuration
   * @param id - The ID of the provider to update
   * @param updates - Partial configuration updates
   * @returns The updated provider configuration
   * @throws ProviderNotFoundError if provider doesn't exist
   */
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

  /**
   * Removes a provider from the registry
   * @param id - The ID of the provider to remove
   * @throws ProviderNotFoundError if provider doesn't exist
   */
  removeProvider(id: string): void {
    if (!this._providers.has(id)) {
      throw new ProviderNotFoundError(id);
    }
    
    this._providers.delete(id);
    
    if (this._defaultProvider === id) {
      this._defaultProvider = null;
    }
  }

  /**
   * Retrieves a provider by ID
   * @param id - The provider ID
   * @returns The provider configuration or undefined if not found
   */
  getProvider(id: string): ProviderConfig | undefined {
    return this._providers.get(id);
  }

  /**
   * Returns all registered providers
   * @returns Array of all provider configurations
   */
  getAllProviders(): ProviderConfig[] {
    return Array.from(this._providers.values());
  }

  /**
   * Returns only enabled providers
   * @returns Array of enabled provider configurations
   */
  getEnabledProviders(): ProviderConfig[] {
    return this.getAllProviders().filter((p) => p.enabled);
  }

  /**
   * Checks if a provider exists in the registry
   * @param id - The provider ID to check
   * @returns True if the provider exists
   */
  hasProvider(id: string): boolean {
    return this._providers.has(id);
  }

  /**
   * Sets the default provider
   * @param id - The ID of the provider to set as default
   * @throws ProviderNotFoundError if provider doesn't exist
   */
  setDefaultProvider(id: string): void {
    if (!this._providers.has(id)) {
      throw new ProviderNotFoundError(id);
    }
    
    this._defaultProvider = id;
  }

  /**
   * Gets the default provider, falling back to first enabled if none set
   * @returns The default provider configuration or undefined
   */
  getDefaultProvider(): ProviderConfig | undefined {
    if (this._defaultProvider) {
      return this._providers.get(this._defaultProvider);
    }
    
    const enabled = this.getEnabledProviders();
    return enabled.length > 0 ? enabled[0] : undefined;
  }

  /**
   * Gets the ID of the default provider
   * @returns The default provider ID or null
   */
  getDefaultProviderId(): string | null {
    return this._defaultProvider;
  }

  /**
   * Enables a provider
   * @param id - The ID of the provider to enable
   */
  enableProvider(id: string): void {
    this.updateProvider(id, { enabled: true });
  }

  /**
   * Disables a provider
   * @param id - The ID of the provider to disable
   */
  disableProvider(id: string): void {
    this.updateProvider(id, { enabled: false });
  }

  /**
   * Validates a provider's configuration
   * @param id - The ID of the provider to validate
   * @returns Validation result with errors if any
   */
  validateProvider(id: string): ReturnType<typeof validateProviderConfig> {
    const provider = this._providers.get(id);
    if (!provider) {
      return { valid: false, errors: [`Provider not found: ${id}`] };
    }
    
    return validateProviderConfig(provider);
  }

  /**
   * Tests connection to a provider
   * @param id - The ID of the provider to test
   * @returns Connection test result with success status and latency
   */
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

  /**
   * Serializes the registry to JSON configuration
   * @returns ProvidersConfig object for persistence
   */
  toJSON(): ProvidersConfig {
    return {
      version: this._version,
      defaultProvider: this._defaultProvider || undefined,
      providers: this.getAllProviders(),
    };
  }

  /**
   * Loads registry state from JSON configuration
   * @param data - ProvidersConfig object to load
   * @throws ProviderValidationError if version is unsupported
   */
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

  /**
   * Clears all providers and resets default
   */
  clear(): void {
    this._providers.clear();
    this._defaultProvider = null;
  }

  /**
   * Gets the number of registered providers
   * @returns The provider count
   */
  get size(): number {
    return this._providers.size;
  }
}

/**
 * Creates a new ProviderRegistry instance with built-in presets
 * @returns A new ProviderRegistry instance
 */
export function createRegistry(): ProviderRegistry {
  return new ProviderRegistry();
}
