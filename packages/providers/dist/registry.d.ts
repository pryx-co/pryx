import { ProviderConfig, ProvidersConfig, ConnectionTestResult } from './types.js';
import { validateProviderConfig } from './validation.js';
export declare class ProviderRegistry {
    private _providers;
    private _defaultProvider;
    private _version;
    constructor();
    addProvider(config: ProviderConfig): void;
    updateProvider(id: string, updates: Partial<ProviderConfig>): ProviderConfig;
    removeProvider(id: string): void;
    getProvider(id: string): ProviderConfig | undefined;
    getAllProviders(): ProviderConfig[];
    getEnabledProviders(): ProviderConfig[];
    hasProvider(id: string): boolean;
    setDefaultProvider(id: string): void;
    getDefaultProvider(): ProviderConfig | undefined;
    getDefaultProviderId(): string | null;
    enableProvider(id: string): void;
    disableProvider(id: string): void;
    validateProvider(id: string): ReturnType<typeof validateProviderConfig>;
    testConnection(id: string): Promise<ConnectionTestResult>;
    toJSON(): ProvidersConfig;
    fromJSON(data: ProvidersConfig): void;
    clear(): void;
    get size(): number;
}
export declare function createRegistry(): ProviderRegistry;
//# sourceMappingURL=registry.d.ts.map