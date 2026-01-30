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
    tools: Array<{ name: string; description: string }>;
    resources?: Array<{ name: string; description: string }>;
    prompts?: Array<{ name: string; description: string }>;
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

export class MCPServerDiscovery {
  private _database: CuratedServersDatabase | null = null;
  private _databasePath: string;

  constructor(databasePath?: string) {
    this._databasePath = databasePath || this._getDefaultDatabasePath();
  }

  async loadDatabase(): Promise<void> {
    try {
      const fs = await import('fs/promises');
      const data = await fs.readFile(this._databasePath, 'utf-8');
      this._database = JSON.parse(data) as CuratedServersDatabase;
    } catch (error) {
      throw new Error(`Failed to load curated servers database: ${error}`);
    }
  }

  async search(filters: SearchFilters = {}): Promise<CuratedServer[]> {
    if (!this._database) {
      await this.loadDatabase();
    }

    let results = this._database!.servers;

    if (filters.category) {
      results = results.filter(s => s.category === filters.category);
    }

    if (filters.author) {
      results = results.filter(s => 
        s.author.toLowerCase().includes(filters.author!.toLowerCase())
      );
    }

    if (filters.query) {
      const query = filters.query.toLowerCase();
      results = results.filter(s => 
        s.name.toLowerCase().includes(query) ||
        s.description.toLowerCase().includes(query) ||
        s.id.toLowerCase().includes(query)
      );
    }

    return results;
  }

  getCategories(): CuratedCategory[] {
    if (!this._database) {
      throw new Error('Database not loaded. Call loadDatabase() first.');
    }
    return this._database.categories;
  }

  getServerById(id: string): CuratedServer | undefined {
    if (!this._database) {
      throw new Error('Database not loaded. Call loadDatabase() first.');
    }
    return this._database.servers.find(s => s.id === id);
  }

  getServersByCategory(categoryId: string): CuratedServer[] {
    if (!this._database) {
      throw new Error('Database not loaded. Call loadDatabase() first.');
    }
    return this._database.servers.filter(s => s.category === categoryId);
  }

  async validateCustomUrl(url: string): Promise<ValidationResult> {
    const errors: string[] = [];
    const warnings: string[] = [];

    try {
      const parsed = new URL(url);

      if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
        errors.push('URL must use HTTP or HTTPS protocol');
      }

      if (parsed.hostname === 'localhost' || parsed.hostname === '127.0.0.1') {
        warnings.push('Localhost URLs are only accessible on this machine');
      }

      if (parsed.protocol === 'http:' && parsed.hostname !== 'localhost') {
        warnings.push('Non-localhost HTTP URLs are insecure. Consider using HTTPS.');
      }

      if (!parsed.pathname || parsed.pathname === '/') {
        warnings.push('URL has no path component. Make sure this is the correct endpoint.');
      }
    } catch {
      errors.push('Invalid URL format');
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
    };
  }

  validateServerId(id: string): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    if (!id) {
      errors.push('Server ID is required');
    } else if (!/^[a-z0-9_-]+$/.test(id)) {
      errors.push('Server ID must contain only lowercase letters, numbers, underscores, and hyphens');
    }

    if (id.length > 64) {
      errors.push('Server ID must be 64 characters or less');
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
    };
  }

  toMCPServerConfig(curated: CuratedServer, customArgs?: string[]): MCPServerConfig {
    const transport = { ...curated.transport };
    
    if (customArgs && transport.type === 'stdio') {
      transport.args = [...(transport.args || []), ...customArgs];
    }

    return {
      id: curated.id,
      name: curated.name,
      enabled: true,
      transport: transport as MCPServerConfig['transport'],
      capabilities: {
        tools: curated.capabilities.tools.map(t => ({
          name: t.name,
          description: t.description,
          inputSchema: {},
        })),
        resources: [],
        prompts: [],
      },
      source: 'curated',
      settings: {
        autoConnect: true,
        timeout: 30000,
        reconnect: true,
        maxReconnectAttempts: 3,
        fallbackServers: [],
      },
    };
  }

  getStats(): {
    totalServers: number;
    totalCategories: number;
    serversByCategory: Record<string, number>;
  } {
    if (!this._database) {
      throw new Error('Database not loaded. Call loadDatabase() first.');
    }

    const serversByCategory: Record<string, number> = {};
    
    for (const server of this._database.servers) {
      serversByCategory[server.category] = (serversByCategory[server.category] || 0) + 1;
    }

    return {
      totalServers: this._database.servers.length,
      totalCategories: this._database.categories.length,
      serversByCategory,
    };
  }

  private _getDefaultDatabasePath(): string {
    const path = require('path');
    return path.join(__dirname, '..', 'data', 'curated-servers.json');
  }
}

export function createServerDiscovery(databasePath?: string): MCPServerDiscovery {
  return new MCPServerDiscovery(databasePath);
}
