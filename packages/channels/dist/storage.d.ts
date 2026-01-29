import { ChannelRegistry } from './registry.js';
export declare class ChannelStorage {
    load(configPath: string): Promise<ChannelRegistry>;
    save(configPath: string, registry: ChannelRegistry): Promise<void>;
    exists(configPath: string): Promise<boolean>;
}
export declare function createStorage(): ChannelStorage;
//# sourceMappingURL=storage.d.ts.map