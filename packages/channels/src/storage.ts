import { readFile, writeFile, mkdir } from 'fs/promises';
import { dirname } from 'path';
import { ChannelRegistry } from './registry.js';
import { ChannelsConfigSchema } from './types.js';

/**
 * Manages persistent storage for channel configurations.
 * Handles loading and saving channel registry data to disk with JSON serialization.
 */
export class ChannelStorage {
  /**
   * Loads channel configuration from a file
   * @param configPath - Path to the configuration file
   * @returns A ChannelRegistry populated with the loaded configuration, or an empty registry if the file doesn't exist
   * @throws {Error} If the file exists but cannot be parsed or validated
   */
  async load(configPath: string): Promise<ChannelRegistry> {
    try {
      const data = await readFile(configPath, 'utf8');
      const parsed = JSON.parse(data);

      const validated = ChannelsConfigSchema.parse(parsed);

      const registry = new ChannelRegistry();
      registry.fromJSON(validated);

      return registry;
    } catch (error) {
      if ((error as NodeJS.ErrnoException).code === 'ENOENT') {
        return new ChannelRegistry();
      }
      throw error;
    }
  }

  /**
   * Saves channel configuration to a file
   * @param configPath - Path to save the configuration file
   * @param registry - The ChannelRegistry to save
   * @throws {Error} If the file cannot be written
   */
  async save(configPath: string, registry: ChannelRegistry): Promise<void> {
    const config = registry.toJSON();
    const data = JSON.stringify(config, null, 2);

    await mkdir(dirname(configPath), { recursive: true });
    await writeFile(configPath, data, { mode: 0o600 });
  }

  /**
   * Checks if a configuration file exists
   * @param configPath - Path to check
   * @returns True if the file exists and is readable
   */
  async exists(configPath: string): Promise<boolean> {
    try {
      await readFile(configPath);
      return true;
    } catch {
      return false;
    }
  }
}

/**
 * Creates a new ChannelStorage instance
 * @returns A new ChannelStorage instance
 */
export function createStorage(): ChannelStorage {
  return new ChannelStorage();
}
