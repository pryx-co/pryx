import { Effect, Context, Layer, ManagedRuntime } from "effect";
import fs from "node:fs";
import path from "node:path";
import yaml from "js-yaml";
import os from "node:os";

const CONFIG_PATH = path.join(os.homedir(), ".pryx", "config.yaml");

export interface AppConfig {
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

export class ConfigLoadError {
  readonly _tag = "ConfigLoadError";
  constructor(
    readonly message: string,
    readonly cause?: unknown
  ) {}
}

export class ConfigSaveError {
  readonly _tag = "ConfigSaveError";
  constructor(
    readonly message: string,
    readonly cause?: unknown
  ) {}
}

export interface ConfigService {
  readonly load: Effect.Effect<AppConfig, ConfigLoadError>;
  readonly save: (cfg: AppConfig) => Effect.Effect<void, ConfigSaveError>;
  readonly update: (updates: Partial<AppConfig>) => Effect.Effect<AppConfig, ConfigSaveError>;
  readonly getValue: <K extends keyof AppConfig>(
    key: K,
    defaultValue?: AppConfig[K]
  ) => Effect.Effect<AppConfig[K] | undefined, ConfigLoadError>;
}

export const ConfigService = Context.GenericTag<ConfigService>("@pryx/tui/ConfigService");

const makeConfigService = Effect.gen(function* () {
  const load = Effect.gen(function* () {
    const result = yield* Effect.try({
      try: () => {
        if (!fs.existsSync(CONFIG_PATH)) return {};
        const content = fs.readFileSync(CONFIG_PATH, "utf-8");
        return (yaml.load(content) as AppConfig) || {};
      },
      catch: error => new ConfigLoadError("Failed to load config", error),
    });
    return result;
  });

  const save = (cfg: AppConfig) =>
    Effect.gen(function* () {
      yield* Effect.try({
        try: () => {
          const dir = path.dirname(CONFIG_PATH);
          if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
          fs.writeFileSync(CONFIG_PATH, yaml.dump(cfg), "utf-8");
        },
        catch: error => new ConfigSaveError("Failed to save config", error),
      });
    });

  const update = (updates: Partial<AppConfig>) =>
    Effect.gen(function* () {
      const current = yield* load;
      const updated = { ...current, ...updates };
      yield* save(updated);
      return updated;
    });

  const getValue = <K extends keyof AppConfig>(key: K, defaultValue?: AppConfig[K]) =>
    Effect.gen(function* () {
      const config = yield* load;
      return config[key] ?? defaultValue;
    });

  return {
    load,
    save,
    update,
    getValue,
  } as ConfigService;
});

export const ConfigServiceLive = Layer.effect(ConfigService, makeConfigService);

const ConfigRuntime = ManagedRuntime.make(ConfigServiceLive);

export const loadConfig = (): AppConfig => {
  return ConfigRuntime.runSync(
    Effect.gen(function* () {
      const configService = yield* ConfigService;
      return yield* configService.load;
    })
  );
};

export const saveConfig = (cfg: AppConfig): void => {
  ConfigRuntime.runSync(
    Effect.gen(function* () {
      const configService = yield* ConfigService;
      yield* configService.save(cfg);
    })
  );
};

export const updateConfig = (updates: Partial<AppConfig>): AppConfig => {
  return ConfigRuntime.runSync(
    Effect.gen(function* () {
      const configService = yield* ConfigService;
      return yield* configService.update(updates);
    })
  );
};

export const getConfigValue = <K extends keyof AppConfig>(
  key: K,
  defaultValue?: AppConfig[K]
): AppConfig[K] | undefined => {
  return ConfigRuntime.runSync(
    Effect.gen(function* () {
      const configService = yield* ConfigService;
      return yield* configService.getValue(key, defaultValue);
    })
  );
};
