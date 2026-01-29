import { ProviderNotFoundError, ProviderAlreadyExistsError, ProviderValidationError, CURRENT_VERSION, } from './types.js';
import { validateProviderConfig, assertValidProviderConfig } from './validation.js';
import { BUILTIN_PRESETS } from './presets.js';
export class ProviderRegistry {
    _providers = new Map();
    _defaultProvider = null;
    _version = CURRENT_VERSION;
    constructor() {
        for (const preset of BUILTIN_PRESETS) {
            this._providers.set(preset.id, { ...preset });
        }
    }
    addProvider(config) {
        if (this._providers.has(config.id)) {
            throw new ProviderAlreadyExistsError(config.id);
        }
        assertValidProviderConfig(config);
        this._providers.set(config.id, { ...config });
    }
    updateProvider(id, updates) {
        const existing = this._providers.get(id);
        if (!existing) {
            throw new ProviderNotFoundError(id);
        }
        const updated = { ...existing, ...updates };
        assertValidProviderConfig(updated);
        this._providers.set(id, updated);
        return updated;
    }
    removeProvider(id) {
        if (!this._providers.has(id)) {
            throw new ProviderNotFoundError(id);
        }
        this._providers.delete(id);
        if (this._defaultProvider === id) {
            this._defaultProvider = null;
        }
    }
    getProvider(id) {
        return this._providers.get(id);
    }
    getAllProviders() {
        return Array.from(this._providers.values());
    }
    getEnabledProviders() {
        return this.getAllProviders().filter((p) => p.enabled);
    }
    hasProvider(id) {
        return this._providers.has(id);
    }
    setDefaultProvider(id) {
        if (!this._providers.has(id)) {
            throw new ProviderNotFoundError(id);
        }
        this._defaultProvider = id;
    }
    getDefaultProvider() {
        if (this._defaultProvider) {
            return this._providers.get(this._defaultProvider);
        }
        const enabled = this.getEnabledProviders();
        return enabled.length > 0 ? enabled[0] : undefined;
    }
    getDefaultProviderId() {
        return this._defaultProvider;
    }
    enableProvider(id) {
        this.updateProvider(id, { enabled: true });
    }
    disableProvider(id) {
        this.updateProvider(id, { enabled: false });
    }
    validateProvider(id) {
        const provider = this._providers.get(id);
        if (!provider) {
            return { valid: false, errors: [`Provider not found: ${id}`] };
        }
        return validateProviderConfig(provider);
    }
    async testConnection(id) {
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
        }
        catch (error) {
            return {
                success: false,
                error: error instanceof Error ? error.message : 'Unknown error',
            };
        }
    }
    toJSON() {
        return {
            version: this._version,
            defaultProvider: this._defaultProvider || undefined,
            providers: this.getAllProviders(),
        };
    }
    fromJSON(data) {
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
    clear() {
        this._providers.clear();
        this._defaultProvider = null;
    }
    get size() {
        return this._providers.size;
    }
}
export function createRegistry() {
    return new ProviderRegistry();
}
//# sourceMappingURL=registry.js.map