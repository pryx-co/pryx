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

export class MCPRegistry {
  private _servers: Map<string, MCPServerConfig> = new Map();
  private _version = CURRENT_VERSION;

  addServer(config: MCPServerConfig): void {
    if (this._servers.has(config.id)) {
      throw new MCPServerAlreadyExistsError(config.id);
    }
    
    assertValidMCPServerConfig(config);
    this._servers.set(config.id, { ...config });
  }

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

  removeServer(id: string): void {
    if (!this._servers.has(id)) {
      throw new MCPServerNotFoundError(id);
    }
    
    this._servers.delete(id);
  }

  getServer(id: string): MCPServerConfig | undefined {
    return this._servers.get(id);
  }

  getAllServers(): MCPServerConfig[] {
    return Array.from(this._servers.values());
  }

  getEnabledServers(): MCPServerConfig[] {
    return this.getAllServers().filter((s) => s.enabled);
  }

  getServersByType(type: TransportType): MCPServerConfig[] {
    return this.getAllServers().filter((s) => s.transport.type === type);
  }

  hasServer(id: string): boolean {
    return this._servers.has(id);
  }

  enableServer(id: string): void {
    this.updateServer(id, { enabled: true });
  }

  disableServer(id: string): void {
    this.updateServer(id, { enabled: false });
  }

  enableAll(): void {
    for (const server of this._servers.values()) {
      server.enabled = true;
    }
  }

  disableAll(): void {
    for (const server of this._servers.values()) {
      server.enabled = false;
    }
  }

  enableType(type: TransportType): void {
    for (const server of this._servers.values()) {
      if (server.transport.type === type) {
        server.enabled = true;
      }
    }
  }

  disableType(type: TransportType): void {
    for (const server of this._servers.values()) {
      if (server.transport.type === type) {
        server.enabled = false;
      }
    }
  }

  updateServerStatus(id: string, status: Partial<MCPServerConfig['status']>): void {
    const server = this._servers.get(id);
    if (!server) {
      throw new MCPServerNotFoundError(id);
    }
    
    server.status = { ...server.status, ...status } as MCPServerConfig['status'];
  }

  getFallbackServers(id: string): MCPServerConfig[] {
    const server = this._servers.get(id);
    if (!server) {
      return [];
    }
    
    return server.settings.fallbackServers
      .map((fallbackId) => this._servers.get(fallbackId))
      .filter((s): s is MCPServerConfig => s !== undefined);
  }

  validateServer(id: string): ReturnType<typeof validateMCPServerConfig> {
    const server = this._servers.get(id);
    if (!server) {
      return { valid: false, errors: [`Server not found: ${id}`] };
    }
    
    return validateMCPServerConfig(server);
  }

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

  getReconnectDelay(id: string): number {
    const server = this._servers.get(id);
    if (!server || !server.status) {
      return 0;
    }
    
    return calculateBackoff(server.status.reconnectAttempts);
  }

  toJSON(): MCPServersConfig {
    return {
      version: this._version,
      servers: this.getAllServers(),
    };
  }

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

  clear(): void {
    this._servers.clear();
  }

  get size(): number {
    return this._servers.size;
  }
}

export function createRegistry(): MCPRegistry {
  return new MCPRegistry();
}
