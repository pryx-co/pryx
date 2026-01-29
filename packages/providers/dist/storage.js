import { readFile, writeFile, mkdir } from 'fs/promises';
import { dirname } from 'path';
import { ProviderRegistry } from './registry.js';
import { ProvidersConfigSchema } from './types.js';
export class ProviderStorage {
    async load(configPath) {
        try {
            const data = await readFile(configPath, 'utf8');
            const parsed = JSON.parse(data);
            const validated = ProvidersConfigSchema.parse(parsed);
            const registry = new ProviderRegistry();
            registry.fromJSON(validated);
            return registry;
        }
        catch (error) {
            if (error.code === 'ENOENT') {
                return new ProviderRegistry();
            }
            throw error;
        }
    }
    async save(configPath, registry) {
        const config = registry.toJSON();
        const data = JSON.stringify(config, null, 2);
        await mkdir(dirname(configPath), { recursive: true });
        await writeFile(configPath, data, { mode: 0o600 });
    }
    async exists(configPath) {
        try {
            await readFile(configPath);
            return true;
        }
        catch {
            return false;
        }
    }
}
export function createStorage() {
    return new ProviderStorage();
}
//# sourceMappingURL=storage.js.map