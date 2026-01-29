import { ProviderRegistry } from './registry.js';
export declare class ProviderStorage {
    load(configPath: string): Promise<ProviderRegistry>;
    save(configPath: string, registry: ProviderRegistry): Promise<void>;
    exists(configPath: string): Promise<boolean>;
}
export declare function createStorage(): ProviderStorage;
//# sourceMappingURL=storage.d.ts.map