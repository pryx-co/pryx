import { MCPServerConfig } from './types.js';
export interface CuratedServer {
    id: string;
    name: string;
    description: string;
    category: string;
    author: string;
    repository: string;
    transport: {
        type: 'stdio' | 'sse' | 'websocket';
        command?: string;
        args?: string[];
        url?: string;
    };
    capabilities: {
        tools: Array<{
            name: string;
            description: string;
        }>;
        resources?: Array<{
            name: string;
            description: string;
        }>;
        prompts?: Array<{
            name: string;
            description: string;
        }>;
    };
    settings?: {
        requiresArgs?: boolean;
        argDescription?: string;
        requiresEnv?: boolean;
        envDescription?: string;
    };
}
export interface CuratedCategory {
    id: string;
    name: string;
    description: string;
}
export interface CuratedServersDatabase {
    version: number;
    lastUpdated: string;
    categories: CuratedCategory[];
    servers: CuratedServer[];
}
export interface SearchFilters {
    category?: string;
    query?: string;
    author?: string;
}
export interface ValidationResult {
    valid: boolean;
    errors: string[];
    warnings: string[];
}
export declare class MCPServerDiscovery {
    private _database;
    private _databasePath;
    constructor(databasePath?: string);
    loadDatabase(): Promise<void>;
    search(filters?: SearchFilters): Promise<CuratedServer[]>;
    getCategories(): CuratedCategory[];
    getServerById(id: string): CuratedServer | undefined;
    getServersByCategory(categoryId: string): CuratedServer[];
    validateCustomUrl(url: string): Promise<ValidationResult>;
    validateServerId(id: string): ValidationResult;
    toMCPServerConfig(curated: CuratedServer, customArgs?: string[]): MCPServerConfig;
    getStats(): {
        totalServers: number;
        totalCategories: number;
        serversByCategory: Record<string, number>;
    };
    private _getDefaultDatabasePath;
}
export declare function createServerDiscovery(databasePath?: string): MCPServerDiscovery;
//# sourceMappingURL=discovery.d.ts.map