/**
 * MCP Storage Module
 *
 * Handles persistence of MCP server configurations to the file system.
 */

import { readFile, writeFile, mkdir } from 'fs/promises';
import { dirname } from 'path';
import { MCPRegistry } from './registry.js';
import { MCPServersConfigSchema } from './types.js';

/**
 * Storage handler for MCP server configurations
 */
export class MCPStorage {
  /**
   * Loads the MCP registry from a file
   * @param configPath - Path to the configuration file
   * @returns The loaded MCP registry
   */
  async load(configPath: string): Promise<MCPRegistry> {
    try {
      const data = await readFile(configPath, 'utf8');
      const parsed = JSON.parse(data);
      
      const validated = MCPServersConfigSchema.parse(parsed);
      
      const registry = new MCPRegistry();
      registry.fromJSON(validated);
      
      return registry;
    } catch (error) {
      if ((error as NodeJS.ErrnoException).code === 'ENOENT') {
        return new MCPRegistry();
      }
      throw error;
    }
  }

  /**
   * Saves the MCP registry to a file
   * @param configPath - Path to the configuration file
   * @param registry - The MCP registry to save
   */
  async save(configPath: string, registry: MCPRegistry): Promise<void> {
    const config = registry.toJSON();
    const data = JSON.stringify(config, null, 2);
    
    await mkdir(dirname(configPath), { recursive: true });
    await writeFile(configPath, data, { mode: 0o600 });
  }

  /**
   * Checks if a configuration file exists
   * @param configPath - Path to the configuration file
   * @returns True if the file exists
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
 * Creates a new MCPStorage instance
 * @returns A new MCPStorage
 */
export function createStorage(): MCPStorage {
  return new MCPStorage();
}
