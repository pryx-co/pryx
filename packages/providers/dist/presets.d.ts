import type { ProviderConfig, ModelConfig } from './types.js';
export declare const OPENAI_MODELS: ModelConfig[];
export declare const ANTHROPIC_MODELS: ModelConfig[];
export declare const GOOGLE_MODELS: ModelConfig[];
export declare const LOCAL_MODELS: ModelConfig[];
export declare const OPENAI_PRESET: ProviderConfig;
export declare const ANTHROPIC_PRESET: ProviderConfig;
export declare const GOOGLE_PRESET: ProviderConfig;
export declare const OLLAMA_PRESET: ProviderConfig;
export declare const LMSTUDIO_PRESET: ProviderConfig;
export declare const BUILTIN_PRESETS: ProviderConfig[];
export declare function getPreset(id: string): ProviderConfig | undefined;
export declare function getAllPresets(): ProviderConfig[];
export declare function getPresetIds(): string[];
//# sourceMappingURL=presets.d.ts.map