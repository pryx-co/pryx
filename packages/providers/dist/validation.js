import { ProviderConfigSchema, ProviderValidationError, } from './types.js';
export function validateProviderConfig(config) {
    const result = ProviderConfigSchema.safeParse(config);
    if (result.success) {
        const errors = [];
        const validated = result.data;
        if (validated.defaultModel) {
            const modelExists = validated.models.some((m) => m.id === validated.defaultModel);
            if (!modelExists) {
                errors.push(`Default model "${validated.defaultModel}" not found in models list`);
            }
        }
        if (validated.type === 'custom' && !validated.baseUrl) {
            errors.push('Custom providers require a baseUrl');
        }
        if (validated.type === 'local' && validated.apiKey) {
            errors.push('Local providers should not have an API key');
        }
        if (validated.type !== 'local' && !validated.apiKey) {
            errors.push(`Provider type "${validated.type}" requires an API key`);
        }
        return {
            valid: errors.length === 0,
            errors,
        };
    }
    return {
        valid: false,
        errors: result.error.errors.map((e) => `${e.path.join('.')}: ${e.message}`),
    };
}
export function assertValidProviderConfig(config) {
    const result = validateProviderConfig(config);
    if (!result.valid) {
        throw new ProviderValidationError(result.errors);
    }
    return config;
}
export function isValidProviderId(id) {
    return /^[a-z0-9_-]+$/.test(id) && id.length >= 1 && id.length <= 64;
}
export function isValidProviderType(type) {
    return ['openai', 'anthropic', 'google', 'local', 'custom'].includes(type);
}
export function isValidUrl(url) {
    try {
        new URL(url);
        return true;
    }
    catch {
        return false;
    }
}
//# sourceMappingURL=validation.js.map