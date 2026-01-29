import { MCPServerNotFoundError, MCPServerAlreadyExistsError, MCPValidationError, CURRENT_VERSION, } from './types.js';
import { validateMCPServerConfig, assertValidMCPServerConfig, calculateBackoff } from './validation.js';
export class MCPRegistry {
    _servers = new Map();
    _version = CURRENT_VERSION;
    addServer(config) {
        if (this._servers.has(config.id)) {
            throw new MCPServerAlreadyExistsError(config.id);
        }
        assertValidMCPServerConfig(config);
        this._servers.set(config.id, { ...config });
    }
    updateServer(id, updates) {
        const existing = this._servers.get(id);
        if (!existing) {
            throw new MCPServerNotFoundError(id);
        }
        const updated = { ...existing, ...updates };
        assertValidMCPServerConfig(updated);
        this._servers.set(id, updated);
        return updated;
    }
    removeServer(id) {
        if (!this._servers.has(id)) {
            throw new MCPServerNotFoundError(id);
        }
        this._servers.delete(id);
    }
    getServer(id) {
        return this._servers.get(id);
    }
    getAllServers() {
        return Array.from(this._servers.values());
    }
    getEnabledServers() {
        return this.getAllServers().filter((s) => s.enabled);
    }
    getServersByType(type) {
        return this.getAllServers().filter((s) => s.transport.type === type);
    }
    hasServer(id) {
        return this._servers.has(id);
    }
    enableServer(id) {
        this.updateServer(id, { enabled: true });
    }
    disableServer(id) {
        this.updateServer(id, { enabled: false });
    }
    enableAll() {
        for (const server of this._servers.values()) {
            server.enabled = true;
        }
    }
    disableAll() {
        for (const server of this._servers.values()) {
            server.enabled = false;
        }
    }
    enableType(type) {
        for (const server of this._servers.values()) {
            if (server.transport.type === type) {
                server.enabled = true;
            }
        }
    }
    disableType(type) {
        for (const server of this._servers.values()) {
            if (server.transport.type === type) {
                server.enabled = false;
            }
        }
    }
    updateServerStatus(id, status) {
        const server = this._servers.get(id);
        if (!server) {
            throw new MCPServerNotFoundError(id);
        }
        server.status = { ...server.status, ...status };
    }
    getFallbackServers(id) {
        const server = this._servers.get(id);
        if (!server) {
            return [];
        }
        return server.settings.fallbackServers
            .map((fallbackId) => this._servers.get(fallbackId))
            .filter((s) => s !== undefined);
    }
    validateServer(id) {
        const server = this._servers.get(id);
        if (!server) {
            return { valid: false, errors: [`Server not found: ${id}`] };
        }
        return validateMCPServerConfig(server);
    }
    async testConnection(id) {
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
        }
        catch (error) {
            return {
                success: false,
                error: error instanceof Error ? error.message : 'Unknown error',
            };
        }
    }
    getReconnectDelay(id) {
        const server = this._servers.get(id);
        if (!server || !server.status) {
            return 0;
        }
        return calculateBackoff(server.status.reconnectAttempts);
    }
    toJSON() {
        return {
            version: this._version,
            servers: this.getAllServers(),
        };
    }
    fromJSON(data) {
        if (data.version !== CURRENT_VERSION) {
            throw new MCPValidationError([`Unsupported version: ${data.version}`]);
        }
        this._servers.clear();
        for (const server of data.servers) {
            assertValidMCPServerConfig(server);
            this._servers.set(server.id, server);
        }
    }
    clear() {
        this._servers.clear();
    }
    get size() {
        return this._servers.size;
    }
}
export function createRegistry() {
    return new MCPRegistry();
}
//# sourceMappingURL=registry.js.map