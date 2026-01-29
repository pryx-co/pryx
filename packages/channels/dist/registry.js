import { ChannelNotFoundError, ChannelAlreadyExistsError, ChannelValidationError, CURRENT_VERSION, } from './types.js';
import { validateChannelConfig, assertValidChannelConfig } from './validation.js';
export class ChannelRegistry {
    _channels = new Map();
    _version = CURRENT_VERSION;
    addChannel(config) {
        if (this._channels.has(config.id)) {
            throw new ChannelAlreadyExistsError(config.id);
        }
        assertValidChannelConfig(config);
        this._channels.set(config.id, { ...config });
    }
    updateChannel(id, updates) {
        const existing = this._channels.get(id);
        if (!existing) {
            throw new ChannelNotFoundError(id);
        }
        const updated = { ...existing, ...updates };
        assertValidChannelConfig(updated);
        this._channels.set(id, updated);
        return updated;
    }
    removeChannel(id) {
        if (!this._channels.has(id)) {
            throw new ChannelNotFoundError(id);
        }
        this._channels.delete(id);
    }
    getChannel(id) {
        return this._channels.get(id);
    }
    getAllChannels() {
        return Array.from(this._channels.values());
    }
    getEnabledChannels() {
        return this.getAllChannels().filter((c) => c.enabled);
    }
    getChannelsByType(type) {
        return this.getAllChannels().filter((c) => c.type === type);
    }
    hasChannel(id) {
        return this._channels.has(id);
    }
    enableChannel(id) {
        this.updateChannel(id, { enabled: true });
    }
    disableChannel(id) {
        this.updateChannel(id, { enabled: false });
    }
    enableAll() {
        for (const channel of this._channels.values()) {
            channel.enabled = true;
        }
    }
    disableAll() {
        for (const channel of this._channels.values()) {
            channel.enabled = false;
        }
    }
    enableType(type) {
        for (const channel of this._channels.values()) {
            if (channel.type === type) {
                channel.enabled = true;
            }
        }
    }
    disableType(type) {
        for (const channel of this._channels.values()) {
            if (channel.type === type) {
                channel.enabled = false;
            }
        }
    }
    updateChannelStatus(id, status) {
        const channel = this._channels.get(id);
        if (!channel) {
            throw new ChannelNotFoundError(id);
        }
        channel.status = { ...channel.status, ...status };
    }
    validateChannel(id) {
        const channel = this._channels.get(id);
        if (!channel) {
            return { valid: false, errors: [`Channel not found: ${id}`] };
        }
        return validateChannelConfig(channel);
    }
    async testConnection(id) {
        const channel = this._channels.get(id);
        if (!channel) {
            return {
                success: false,
                error: `Channel not found: ${id}`,
            };
        }
        const start = performance.now();
        try {
            const latency = performance.now() - start;
            return {
                success: true,
                latency,
            };
        }
        catch (error) {
            return {
                success: false,
                error: error instanceof Error ? error.message : 'Unknown error',
            };
        }
    }
    toJSON() {
        return {
            version: this._version,
            channels: this.getAllChannels(),
        };
    }
    fromJSON(data) {
        if (data.version !== CURRENT_VERSION) {
            throw new ChannelValidationError([`Unsupported version: ${data.version}`]);
        }
        this._channels.clear();
        for (const channel of data.channels) {
            assertValidChannelConfig(channel);
            this._channels.set(channel.id, channel);
        }
    }
    clear() {
        this._channels.clear();
    }
    get size() {
        return this._channels.size;
    }
}
export function createRegistry() {
    return new ChannelRegistry();
}
//# sourceMappingURL=registry.js.map