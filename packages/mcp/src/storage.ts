import { readFile, writeFile, mkdir } from 'fs/promises';
import { dirname } from 'path';
import { MCPRegistry } from './registry.js';
import { MCPServersConfigSchema } from './types.js';

export class MCPStorage {
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

  async save(configPath: string, registry: MCPRegistry): Promise<void> {
    const config = registry.toJSON();
    const data = JSON.stringify(config, null, 2);
    
    await mkdir(dirname(configPath), { recursive: true });
    await writeFile(configPath, data, { mode: 0o600 });
  }

  async exists(configPath: string): Promise<boolean> {
    try {
      await readFile(configPath);
      return true;
    } catch {
      return false;
    }
  }
}

export function createStorage(): MCPStorage {
  return new MCPStorage();
}
