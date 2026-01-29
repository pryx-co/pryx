import { MCPRegistry } from './registry.js';
export declare class MCPStorage {
    load(configPath: string): Promise<MCPRegistry>;
    save(configPath: string, registry: MCPRegistry): Promise<void>;
    exists(configPath: string): Promise<boolean>;
}
export declare function createStorage(): MCPStorage;
//# sourceMappingURL=storage.d.ts.map