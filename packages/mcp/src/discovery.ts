/**
 * MCP Server Discovery Module
 *
 * Provides functionality for discovering and searching curated MCP servers.
 */

import { MCPServerConfig } from './types.js';
import { dirname, join } from 'node:path';

/**
 * Interface representing a curated MCP server
 */
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

/**
 * Interface representing a curated server category
 */
export interface CuratedCategory {
  id: string;
  name: string;
  description: string;
}

/**
 * Interface representing the curated servers database
 */
export interface CuratedServersDatabase {
  version: number;
  lastUpdated: string;
  categories: CuratedCategory[];
  servers: CuratedServer[];
}

/**
 * Interface for search filters
 */
export interface SearchFilters {
  category?: string;
  query?: string;
  author?: string;
}

/**
 * Interface for validation results
 */
export interface ValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
}

/**
 * Class for discovering and managing curated MCP servers
 */
export class MCPServerDiscovery {
  private _database: CuratedServersDatabase | null = null;
  private _databasePath: string;

  /**
   * Creates a new MCPServerDiscovery instance
   * @param databasePath - Optional path to the database file
   */
  constructor(databasePath?: string) {
    this._databasePath = databasePath || this._getDefaultDatabasePath();
  }

  /**
   * Loads the curated servers database
   */
  async loadDatabase(): Promise<void> {
    try {
      const fs = await import('fs/promises');
      const data = await fs.readFile(this._databasePath, 'utf-8');
      this._database = JSON.parse(data) as CuratedServersDatabase;
    } catch (error) {
      throw new Error(`Failed to load curated servers database: ${error}`);
    }
  }

  /**
   * Searches for curated servers based on filters
   * @param filters - Search filters
   * @returns Array of matching curated servers
   */
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

  /**
   * Gets all categories
   * @returns Array of categories
   */
  getCategories(): CuratedCategory[] {
    if (!this._database) {
      throw new Error('Database not loaded. Call loadDatabase() first.');
    }
    return this._database.categories;
  }

  /**
   * Gets a server by ID
   * @param id - The server ID
   * @returns The server or undefined
   */
  getServerById(id: string): CuratedServer | undefined {
    if (!this._database) {
      throw new Error('Database not loaded. Call loadDatabase() first.');
    }
    return this._database.servers.find(s => s.id === id);
  }

  /**
   * Gets servers by category
   * @param categoryId - The category ID
   * @returns Array of servers in the category
   */
  getServersByCategory(categoryId: string): CuratedServer[] {
    if (!this._database) {
      throw new Error('Database not loaded. Call loadDatabase() first.');
    }
    return this._database.servers.filter(s => s.category === categoryId);
  }

  /**
   * Validates a custom URL
   * @param url - The URL to validate
   * @returns Validation result
   */
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

  /**
   * Validates a server ID
   * @param id - The server ID to validate
   * @returns Validation result
   */
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

  /**
   * Converts a curated server to an MCP server config
   * @param curated - The curated server
   * @param customArgs - Optional custom arguments
   * @returns The MCP server configuration
   */
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

  /**
   * Gets statistics about the database
   * @returns Database statistics
   */
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

  /**
   * Gets the default database path
   * @returns The default database path
   */
  private _getDefaultDatabasePath(): string {
    const moduleDir = typeof __dirname === 'string'
      ? __dirname
      : dirname(process.argv[1] || process.cwd());
    return join(moduleDir, '..', 'data', 'curated-servers.json');
  }
}

/**
 * Creates a new MCPServerDiscovery instance
 * @param databasePath - Optional path to the database file
 * @returns A new MCPServerDiscovery
 */
export function createServerDiscovery(databasePath?: string): MCPServerDiscovery {
  return new MCPServerDiscovery(databasePath);
}
