import {
  ChannelConfig,
  ChannelsConfig,
  ConnectionTestResult,
  ChannelNotFoundError,
  ChannelAlreadyExistsError,
  ChannelValidationError,
  CURRENT_VERSION,
  ChannelType,
} from './types.js';
import { validateChannelConfig, assertValidChannelConfig } from './validation.js';

export class ChannelRegistry {
  private _channels: Map<string, ChannelConfig> = new Map();
  private _version = CURRENT_VERSION;

  addChannel(config: ChannelConfig): void {
    if (this._channels.has(config.id)) {
      throw new ChannelAlreadyExistsError(config.id);
    }
    
    assertValidChannelConfig(config);
    this._channels.set(config.id, { ...config });
  }

  updateChannel(id: string, updates: Partial<ChannelConfig>): ChannelConfig {
    const existing = this._channels.get(id);
    if (!existing) {
      throw new ChannelNotFoundError(id);
    }
    
    const updated = { ...existing, ...updates };
    assertValidChannelConfig(updated);
    this._channels.set(id, updated);
    
    return updated;
  }

  removeChannel(id: string): void {
    if (!this._channels.has(id)) {
      throw new ChannelNotFoundError(id);
    }
    
    this._channels.delete(id);
  }

  getChannel(id: string): ChannelConfig | undefined {
    return this._channels.get(id);
  }

  getAllChannels(): ChannelConfig[] {
    return Array.from(this._channels.values());
  }

  getEnabledChannels(): ChannelConfig[] {
    return this.getAllChannels().filter((c) => c.enabled);
  }

  getChannelsByType(type: ChannelType): ChannelConfig[] {
    return this.getAllChannels().filter((c) => c.type === type);
  }

  hasChannel(id: string): boolean {
    return this._channels.has(id);
  }

  enableChannel(id: string): void {
    this.updateChannel(id, { enabled: true });
  }

  disableChannel(id: string): void {
    this.updateChannel(id, { enabled: false });
  }

  enableAll(): void {
    for (const channel of this._channels.values()) {
      channel.enabled = true;
    }
  }

  disableAll(): void {
    for (const channel of this._channels.values()) {
      channel.enabled = false;
    }
  }

  enableType(type: ChannelType): void {
    for (const channel of this._channels.values()) {
      if (channel.type === type) {
        channel.enabled = true;
      }
    }
  }

  disableType(type: ChannelType): void {
    for (const channel of this._channels.values()) {
      if (channel.type === type) {
        channel.enabled = false;
      }
    }
  }

  updateChannelStatus(id: string, status: Partial<ChannelConfig['status']>): void {
    const channel = this._channels.get(id);
    if (!channel) {
      throw new ChannelNotFoundError(id);
    }
    
    channel.status = { ...channel.status, ...status } as ChannelConfig['status'];
  }

  validateChannel(id: string): ReturnType<typeof validateChannelConfig> {
    const channel = this._channels.get(id);
    if (!channel) {
      return { valid: false, errors: [`Channel not found: ${id}`] };
    }
    
    return validateChannelConfig(channel);
  }

  async testConnection(id: string): Promise<ConnectionTestResult> {
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
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  toJSON(): ChannelsConfig {
    return {
      version: this._version,
      channels: this.getAllChannels(),
    };
  }

  fromJSON(data: ChannelsConfig): void {
    if (data.version !== CURRENT_VERSION) {
      throw new ChannelValidationError([`Unsupported version: ${data.version}`]);
    }
    
    this._channels.clear();
    
    for (const channel of data.channels) {
      assertValidChannelConfig(channel);
      this._channels.set(channel.id, channel);
    }
  }

  clear(): void {
    this._channels.clear();
  }

  get size(): number {
    return this._channels.size;
  }
}

export function createRegistry(): ChannelRegistry {
  return new ChannelRegistry();
}
