import fs from 'node:fs';
import path from 'node:path';
import yaml from 'js-yaml';
import os from 'node:os';

const CONFIG_PATH = path.join(os.homedir(), '.pryx', 'config.yaml');

export interface Config {
    model_provider?: string;
    model_name?: string;
    openai_key?: string;
    anthropic_key?: string;
    ollama_endpoint?: string;
    telegram_token?: string;
    telegram_enabled?: boolean;
    webhook_enabled?: boolean;
    [key: string]: any;
}

export const loadConfig = (): Config => {
    try {
        if (!fs.existsSync(CONFIG_PATH)) return {};
        const content = fs.readFileSync(CONFIG_PATH, 'utf-8');
        return (yaml.load(content) as Config) || {};
    } catch (e) {
        console.error("[Config] Failed to load config:", e);
        return {};
    }
};

export const saveConfig = (cfg: Config): void => {
    try {
        const dir = path.dirname(CONFIG_PATH);
        if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
        fs.writeFileSync(CONFIG_PATH, yaml.dump(cfg), 'utf-8');
        console.log("[Config] Saved successfully");
    } catch (e) {
        console.error("[Config] Failed to save config:", e);
        throw e;
    }
};

export const updateConfig = (updates: Partial<Config>): Config => {
    const current = loadConfig();
    const updated = { ...current, ...updates };
    saveConfig(updated);
    return updated;
};

export const getConfigValue = <K extends keyof Config>(
    key: K,
    defaultValue?: Config[K]
): Config[K] | undefined => {
    const config = loadConfig();
    return config[key] ?? defaultValue;
};
