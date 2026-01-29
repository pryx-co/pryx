export const OPENAI_MODELS = [
    {
        id: 'gpt-4o',
        name: 'GPT-4o',
        maxTokens: 128000,
        supportsStreaming: true,
        supportsVision: true,
        supportsTools: true,
        costPer1KInput: 0.005,
        costPer1KOutput: 0.015,
    },
    {
        id: 'gpt-4o-mini',
        name: 'GPT-4o Mini',
        maxTokens: 128000,
        supportsStreaming: true,
        supportsVision: true,
        supportsTools: true,
        costPer1KInput: 0.00015,
        costPer1KOutput: 0.0006,
    },
    {
        id: 'gpt-4-turbo',
        name: 'GPT-4 Turbo',
        maxTokens: 128000,
        supportsStreaming: true,
        supportsVision: true,
        supportsTools: true,
        costPer1KInput: 0.01,
        costPer1KOutput: 0.03,
    },
    {
        id: 'gpt-3.5-turbo',
        name: 'GPT-3.5 Turbo',
        maxTokens: 16385,
        supportsStreaming: true,
        supportsVision: false,
        supportsTools: true,
        costPer1KInput: 0.0005,
        costPer1KOutput: 0.0015,
    },
];
export const ANTHROPIC_MODELS = [
    {
        id: 'claude-3-opus-20240229',
        name: 'Claude 3 Opus',
        maxTokens: 200000,
        supportsStreaming: true,
        supportsVision: true,
        supportsTools: true,
        costPer1KInput: 0.015,
        costPer1KOutput: 0.075,
    },
    {
        id: 'claude-3-sonnet-20240229',
        name: 'Claude 3 Sonnet',
        maxTokens: 200000,
        supportsStreaming: true,
        supportsVision: true,
        supportsTools: true,
        costPer1KInput: 0.003,
        costPer1KOutput: 0.015,
    },
    {
        id: 'claude-3-haiku-20240307',
        name: 'Claude 3 Haiku',
        maxTokens: 200000,
        supportsStreaming: true,
        supportsVision: true,
        supportsTools: true,
        costPer1KInput: 0.00025,
        costPer1KOutput: 0.00125,
    },
];
export const GOOGLE_MODELS = [
    {
        id: 'gemini-1.5-pro',
        name: 'Gemini 1.5 Pro',
        maxTokens: 1048576,
        supportsStreaming: true,
        supportsVision: true,
        supportsTools: true,
        costPer1KInput: 0.0035,
        costPer1KOutput: 0.0105,
    },
    {
        id: 'gemini-1.5-flash',
        name: 'Gemini 1.5 Flash',
        maxTokens: 1048576,
        supportsStreaming: true,
        supportsVision: true,
        supportsTools: true,
        costPer1KInput: 0.00035,
        costPer1KOutput: 0.00105,
    },
];
export const LOCAL_MODELS = [
    {
        id: 'llama2',
        name: 'Llama 2',
        maxTokens: 4096,
        supportsStreaming: true,
        supportsVision: false,
        supportsTools: false,
    },
    {
        id: 'codellama',
        name: 'CodeLlama',
        maxTokens: 16384,
        supportsStreaming: true,
        supportsVision: false,
        supportsTools: false,
    },
    {
        id: 'mistral',
        name: 'Mistral',
        maxTokens: 8192,
        supportsStreaming: true,
        supportsVision: false,
        supportsTools: false,
    },
];
export const OPENAI_PRESET = {
    id: 'openai',
    name: 'OpenAI',
    type: 'openai',
    enabled: true,
    defaultModel: 'gpt-4o',
    models: OPENAI_MODELS,
    timeout: 30000,
    retries: 3,
};
export const ANTHROPIC_PRESET = {
    id: 'anthropic',
    name: 'Anthropic',
    type: 'anthropic',
    enabled: true,
    defaultModel: 'claude-3-sonnet-20240229',
    models: ANTHROPIC_MODELS,
    timeout: 30000,
    retries: 3,
};
export const GOOGLE_PRESET = {
    id: 'google',
    name: 'Google',
    type: 'google',
    enabled: true,
    defaultModel: 'gemini-1.5-pro',
    models: GOOGLE_MODELS,
    timeout: 30000,
    retries: 3,
};
export const OLLAMA_PRESET = {
    id: 'ollama',
    name: 'Ollama (Local)',
    type: 'local',
    enabled: false,
    defaultModel: 'llama2',
    baseUrl: 'http://localhost:11434',
    models: LOCAL_MODELS,
    timeout: 60000,
    retries: 1,
};
export const LMSTUDIO_PRESET = {
    id: 'lmstudio',
    name: 'LM Studio (Local)',
    type: 'local',
    enabled: false,
    baseUrl: 'http://localhost:1234',
    models: LOCAL_MODELS,
    timeout: 60000,
    retries: 1,
};
export const BUILTIN_PRESETS = [
    OPENAI_PRESET,
    ANTHROPIC_PRESET,
    GOOGLE_PRESET,
    OLLAMA_PRESET,
    LMSTUDIO_PRESET,
];
export function getPreset(id) {
    return BUILTIN_PRESETS.find((p) => p.id === id);
}
export function getAllPresets() {
    return [...BUILTIN_PRESETS];
}
export function getPresetIds() {
    return BUILTIN_PRESETS.map((p) => p.id);
}
//# sourceMappingURL=presets.js.map