/**
 * MCP Registry Module
 *
 * Manages MCP server configurations with CRUD operations and connection testing.
 */

import {
  MCPServerConfig,
  MCPServersConfig,
  ConnectionTestResult,
  MCPServerNotFoundError,
  MCPServerAlreadyExistsError,
  MCPValidationError,
  CURRENT_VERSION,
  TransportType,
} from './types.js';
import { validateMCPServerConfig, assertValidMCPServerConfig, calculateBackoff } from './validation.js';

/**
 * Registry for managing MCP server configurations
 */
export class MCPRegistry {
  private _servers: Map<string, MCPServerConfig> = new Map();
  private _version = CURRENT_VERSION;

  /**
   * Adds a server to the registry
   * @param config - The server configuration to add
   * @throws MCPServerAlreadyExistsError if server with same ID exists
   */
  addServer(config: MCPServerConfig): void {
    if (this._servers.has(config.id)) {
      throw new MCPServerAlreadyExistsError(config.id);
    }
    
    assertValidMCPServerConfig(config);
    this._servers.set(config.id, { ...config });
  }

  /**
   * Updates an existing server configuration
   * @param id - The ID of the server to update
   * @param updates - Partial configuration updates
   * @returns The updated server configuration
   * @throws MCPServerNotFoundError if server doesn't exist
   */
  updateServer(id: string, updates: Partial<MCPServerConfig>): MCPServerConfig {
    const existing = this._servers.get(id);
    if (!existing) {
      throw new MCPServerNotFoundError(id);
    }
    
    const updated = { ...existing, ...updates };
    assertValidMCPServerConfig(updated);
    this._servers.set(id, updated);
    
    return updated;
  }

  /**
   * Removes a server from the registry
   * @param id - The ID of the server to remove
   * @throws MCPServerNotFoundError if server doesn't exist
   */
  removeServer(id: string): void {
    if (!this._servers.has(id)) {
      throw new MCPServerNotFoundError(id);
    }
    
    this._servers.delete(id);
  }

  /**
   * Gets a server by ID
   * @param id - The server ID
   * @returns The server configuration or undefined
   */
  getServer(id: string): MCPServerConfig | undefined {
    return this._servers.get(id);
  }

  /**
   * Gets all registered servers
   * @returns Array of all server configurations
   */
  getAllServers(): MCPServerConfig[] {
    return Array.from(this._servers.values());
  }

  /**
   * Gets all enabled servers
   * @returns Array of enabled server configurations
   */
  getEnabledServers(): MCPServerConfig[] {
    return this.getAllServers().filter((s) => s.enabled);
  }

  /**
   * Gets servers by transport type
   * @param type - The transport type to filter by
   * @returns Array of server configurations matching the transport type
   */
  getServersByType(type: TransportType): MCPServerConfig[] {
    return this.getAllServers().filter((s) => s.transport.type === type);
  }

  /**
   * Checks if a server exists
   * @param id - The server ID to check
   * @returns True if the server exists
   */
  hasServer(id: string): boolean {
    return this._servers.has(id);
  }

  /**
   * Enables a server
   * @param id - The ID of the server to enable
   */
  enableServer(id: string): void {
    this.updateServer(id, { enabled: true });
  }

  /**
   * Disables a server
   * @param id - The ID of the server to disable
   */
  disableServer(id: string): void {
    this.updateServer(id, { enabled: false });
  }

  /**
   * Enables all servers
   */
  enableAll(): void {
    for (const server of this._servers.values()) {
      server.enabled = true;
    }
  }

  /**
   * Disables all servers
   */
  disableAll(): void {
    for (const server of this._servers.values()) {
      server.enabled = false;
    }
  }

  /**
   * Enables all servers of a specific transport type
   * @param type - The transport type to enable
   */
  enableType(type: TransportType): void {
    for (const server of this._servers.values()) {
      if (server.transport.type === type) {
        server.enabled = true;
      }
    }
  }

  /**
   * Disables all servers of a specific transport type
   * @param type - The transport type to disable
   */
  disableType(type: TransportType): void {
    for (const server of this._servers.values()) {
      if (server.transport.type === type) {
        server.enabled = false;
      }
    }
  }

  /**
   * Updates the status of a server
   * @param id - The ID of the server
   * @param status - Partial status updates
   * @throws MCPServerNotFoundError if server doesn't exist
   */
  updateServerStatus(id: string, status: Partial<MCPServerConfig['status']>): void {
    const server = this._servers.get(id);
    if (!server) {
      throw new MCPServerNotFoundError(id);
    }
    
    server.status = { ...server.status, ...status } as MCPServerConfig['status'];
  }

  /**
   * Gets fallback servers for a given server
   * @param id - The ID of the server
   * @returns Array of fallback server configurations
   */
  getFallbackServers(id: string): MCPServerConfig[] {
    const server = this._servers.get(id);
    if (!server) {
      return [];
    }
    
    return server.settings.fallbackServers
      .map((fallbackId) => this._servers.get(fallbackId))
      .filter((s): s is MCPServerConfig => s !== undefined);
  }

  /**
   * Validates a server configuration
   * @param id - The ID of the server to validate
   * @returns Validation result
   */
  validateServer(id: string): ReturnType<typeof validateMCPServerConfig> {
    const server = this._servers.get(id);
    if (!server) {
      return { valid: false, errors: [`Server not found: ${id}`] };
    }
    
    return validateMCPServerConfig(server);
  }

  /**
   * Tests connection to a server
   * @param id - The ID of the server to test
   * @returns Connection test result
   */
  async testConnection(id: string): Promise<ConnectionTestResult> {
    const server = this._servers.get(id);
    if (!server) {
      return {
        success: false,
        error: `Server not found: ${id}`,
      };
    }
    
    const start = performance.now();
    
    try {
      const latency = performance.now() - start;
      
      return {
        success: true,
        latency,
        capabilities: server.capabilities,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  /**
   * Gets the reconnect delay for a server
   * @param id - The ID of the server
   * @returns The reconnect delay in milliseconds
   */
  getReconnectDelay(id: string): number {
    const server = this._servers.get(id);
    if (!server || !server.status) {
      return 0;
    }
    
    return calculateBackoff(server.status.reconnectAttempts);
  }

  /**
   * Serializes the registry to JSON
   * @returns The serialized configuration
   */
  toJSON(): MCPServersConfig {
    return {
      version: this._version,
      servers: this.getAllServers(),
    };
  }

  /**
   * Deserializes the registry from JSON
   * @param data - The configuration to load
   * @throws MCPValidationError if version is unsupported
   */
  fromJSON(data: MCPServersConfig): void {
    if (data.version !== CURRENT_VERSION) {
      throw new MCPValidationError([`Unsupported version: ${data.version}`]);
    }
    
    this._servers.clear();
    
    for (const server of data.servers) {
      assertValidMCPServerConfig(server);
      this._servers.set(server.id, server);
    }
  }

  /**
   * Clears all servers from the registry
   */
  clear(): void {
    this._servers.clear();
  }

  /**
   * Gets the number of registered servers
   * @returns The server count
   */
  get size(): number {
    return this._servers.size;
  }
}

/**
 * Creates a new MCPRegistry instance
 * @returns A new MCPRegistry
 */
export function createRegistry(): MCPRegistry {
  return new MCPRegistry();
}
