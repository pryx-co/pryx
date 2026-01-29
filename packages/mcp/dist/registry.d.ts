import { MCPServerConfig, MCPServersConfig, ConnectionTestResult, TransportType } from './types.js';
import { validateMCPServerConfig } from './validation.js';
export declare class MCPRegistry {
    private _servers;
    private _version;
    addServer(config: MCPServerConfig): void;
    updateServer(id: string, updates: Partial<MCPServerConfig>): MCPServerConfig;
    removeServer(id: string): void;
    getServer(id: string): MCPServerConfig | undefined;
    getAllServers(): MCPServerConfig[];
    getEnabledServers(): MCPServerConfig[];
    getServersByType(type: TransportType): MCPServerConfig[];
    hasServer(id: string): boolean;
    enableServer(id: string): void;
    disableServer(id: string): void;
    enableAll(): void;
    disableAll(): void;
    enableType(type: TransportType): void;
    disableType(type: TransportType): void;
    updateServerStatus(id: string, status: Partial<MCPServerConfig['status']>): void;
    getFallbackServers(id: string): MCPServerConfig[];
    validateServer(id: string): ReturnType<typeof validateMCPServerConfig>;
    testConnection(id: string): Promise<ConnectionTestResult>;
    getReconnectDelay(id: string): number;
    toJSON(): MCPServersConfig;
    fromJSON(data: MCPServersConfig): void;
    clear(): void;
    get size(): number;
}
export declare function createRegistry(): MCPRegistry;
//# sourceMappingURL=registry.d.ts.map