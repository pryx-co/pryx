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

/**
 * Manages a registry of channel configurations with CRUD operations,
 * validation, and serialization support.
 */
export class ChannelRegistry {
  private _channels: Map<string, ChannelConfig> = new Map();
  private _version = CURRENT_VERSION;

  /**
   * Adds a new channel to the registry
   * @param config - The channel configuration to add
   * @throws {ChannelAlreadyExistsError} If a channel with the same ID already exists
   * @throws {ChannelValidationError} If the configuration is invalid
   */
  addChannel(config: ChannelConfig): void {
    if (this._channels.has(config.id)) {
      throw new ChannelAlreadyExistsError(config.id);
    }

    assertValidChannelConfig(config);
    this._channels.set(config.id, { ...config });
  }

  /**
   * Updates an existing channel's configuration
   * @param id - The ID of the channel to update
   * @param updates - Partial configuration updates
   * @returns The updated channel configuration
   * @throws {ChannelNotFoundError} If the channel doesn't exist
   * @throws {ChannelValidationError} If the updated configuration is invalid
   */
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

  /**
   * Removes a channel from the registry
   * @param id - The ID of the channel to remove
   * @throws {ChannelNotFoundError} If the channel doesn't exist
   */
  removeChannel(id: string): void {
    if (!this._channels.has(id)) {
      throw new ChannelNotFoundError(id);
    }

    this._channels.delete(id);
  }

  /**
   * Gets a channel by ID
   * @param id - The channel ID
   * @returns The channel configuration or undefined if not found
   */
  getChannel(id: string): ChannelConfig | undefined {
    return this._channels.get(id);
  }

  /**
   * Gets all channels in the registry
   * @returns Array of all channel configurations
   */
  getAllChannels(): ChannelConfig[] {
    return Array.from(this._channels.values());
  }

  /**
   * Gets all enabled channels
   * @returns Array of enabled channel configurations
   */
  getEnabledChannels(): ChannelConfig[] {
    return this.getAllChannels().filter((c) => c.enabled);
  }

  /**
   * Gets channels by type
   * @param type - The channel type to filter by
   * @returns Array of channel configurations of the specified type
   */
  getChannelsByType(type: ChannelType): ChannelConfig[] {
    return this.getAllChannels().filter((c) => c.type === type);
  }

  /**
   * Checks if a channel exists in the registry
   * @param id - The channel ID to check
   * @returns True if the channel exists
   */
  hasChannel(id: string): boolean {
    return this._channels.has(id);
  }

  /**
   * Enables a channel
   * @param id - The ID of the channel to enable
   * @throws {ChannelNotFoundError} If the channel doesn't exist
   */
  enableChannel(id: string): void {
    this.updateChannel(id, { enabled: true });
  }

  /**
   * Disables a channel
   * @param id - The ID of the channel to disable
   * @throws {ChannelNotFoundError} If the channel doesn't exist
   */
  disableChannel(id: string): void {
    this.updateChannel(id, { enabled: false });
  }

  /**
   * Enables all channels in the registry
   */
  enableAll(): void {
    for (const channel of this._channels.values()) {
      channel.enabled = true;
    }
  }

  /**
   * Disables all channels in the registry
   */
  disableAll(): void {
    for (const channel of this._channels.values()) {
      channel.enabled = false;
    }
  }

  /**
   * Enables all channels of a specific type
   * @param type - The channel type to enable
   */
  enableType(type: ChannelType): void {
    for (const channel of this._channels.values()) {
      if (channel.type === type) {
        channel.enabled = true;
      }
    }
  }

  /**
   * Disables all channels of a specific type
   * @param type - The channel type to disable
   */
  disableType(type: ChannelType): void {
    for (const channel of this._channels.values()) {
      if (channel.type === type) {
        channel.enabled = false;
      }
    }
  }

  /**
   * Updates a channel's connection status
   * @param id - The channel ID
   * @param status - Partial status updates
   * @throws {ChannelNotFoundError} If the channel doesn't exist
   */
  updateChannelStatus(id: string, status: Partial<ChannelConfig['status']>): void {
    const channel = this._channels.get(id);
    if (!channel) {
      throw new ChannelNotFoundError(id);
    }

    channel.status = { ...channel.status, ...status } as ChannelConfig['status'];
  }

  /**
   * Validates a channel's configuration
   * @param id - The channel ID
   * @returns Validation result
   */
  validateChannel(id: string): ReturnType<typeof validateChannelConfig> {
    const channel = this._channels.get(id);
    if (!channel) {
      return { valid: false, errors: [`Channel not found: ${id}`] };
    }

    return validateChannelConfig(channel);
  }

  /**
   * Tests connection to a channel (placeholder implementation)
   * @param id - The channel ID
   * @returns Connection test result
   */
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

  /**
   * Serializes the registry to JSON
   * @returns Channels configuration object
   */
  toJSON(): ChannelsConfig {
    return {
      version: this._version,
      channels: this.getAllChannels(),
    };
  }

  /**
   * Deserializes the registry from JSON
   * @param data - Channels configuration object
   * @throws {ChannelValidationError} If version is unsupported or data is invalid
   */
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

  /**
   * Clears all channels from the registry
   */
  clear(): void {
    this._channels.clear();
  }

  /**
   * Gets the number of channels in the registry
   */
  get size(): number {
    return this._channels.size;
  }
}

/**
 * Creates a new ChannelRegistry instance
 * @returns A new ChannelRegistry instance
 */
export function createRegistry(): ChannelRegistry {
  return new ChannelRegistry();
}
