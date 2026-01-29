import { z } from 'zod';
export declare const ProviderType: z.ZodEnum<["openai", "anthropic", "google", "local", "custom"]>;
export declare const ModelConfigSchema: z.ZodObject<{
    id: z.ZodString;
    name: z.ZodString;
    maxTokens: z.ZodNumber;
    supportsStreaming: z.ZodDefault<z.ZodBoolean>;
    supportsVision: z.ZodDefault<z.ZodBoolean>;
    supportsTools: z.ZodDefault<z.ZodBoolean>;
    costPer1KInput: z.ZodOptional<z.ZodNumber>;
    costPer1KOutput: z.ZodOptional<z.ZodNumber>;
}, "strip", z.ZodTypeAny, {
    id: string;
    name: string;
    maxTokens: number;
    supportsStreaming: boolean;
    supportsVision: boolean;
    supportsTools: boolean;
    costPer1KInput?: number | undefined;
    costPer1KOutput?: number | undefined;
}, {
    id: string;
    name: string;
    maxTokens: number;
    supportsStreaming?: boolean | undefined;
    supportsVision?: boolean | undefined;
    supportsTools?: boolean | undefined;
    costPer1KInput?: number | undefined;
    costPer1KOutput?: number | undefined;
}>;
export declare const RateLimitConfigSchema: z.ZodObject<{
    requestsPerMinute: z.ZodOptional<z.ZodNumber>;
    tokensPerMinute: z.ZodOptional<z.ZodNumber>;
    requestsPerDay: z.ZodOptional<z.ZodNumber>;
}, "strip", z.ZodTypeAny, {
    requestsPerMinute?: number | undefined;
    tokensPerMinute?: number | undefined;
    requestsPerDay?: number | undefined;
}, {
    requestsPerMinute?: number | undefined;
    tokensPerMinute?: number | undefined;
    requestsPerDay?: number | undefined;
}>;
export declare const ProviderConfigSchema: z.ZodObject<{
    id: z.ZodString;
    name: z.ZodString;
    type: z.ZodEnum<["openai", "anthropic", "google", "local", "custom"]>;
    enabled: z.ZodDefault<z.ZodBoolean>;
    defaultModel: z.ZodOptional<z.ZodString>;
    apiKey: z.ZodOptional<z.ZodString>;
    baseUrl: z.ZodOptional<z.ZodString>;
    organization: z.ZodOptional<z.ZodString>;
    models: z.ZodArray<z.ZodObject<{
        id: z.ZodString;
        name: z.ZodString;
        maxTokens: z.ZodNumber;
        supportsStreaming: z.ZodDefault<z.ZodBoolean>;
        supportsVision: z.ZodDefault<z.ZodBoolean>;
        supportsTools: z.ZodDefault<z.ZodBoolean>;
        costPer1KInput: z.ZodOptional<z.ZodNumber>;
        costPer1KOutput: z.ZodOptional<z.ZodNumber>;
    }, "strip", z.ZodTypeAny, {
        id: string;
        name: string;
        maxTokens: number;
        supportsStreaming: boolean;
        supportsVision: boolean;
        supportsTools: boolean;
        costPer1KInput?: number | undefined;
        costPer1KOutput?: number | undefined;
    }, {
        id: string;
        name: string;
        maxTokens: number;
        supportsStreaming?: boolean | undefined;
        supportsVision?: boolean | undefined;
        supportsTools?: boolean | undefined;
        costPer1KInput?: number | undefined;
        costPer1KOutput?: number | undefined;
    }>, "many">;
    rateLimits: z.ZodOptional<z.ZodObject<{
        requestsPerMinute: z.ZodOptional<z.ZodNumber>;
        tokensPerMinute: z.ZodOptional<z.ZodNumber>;
        requestsPerDay: z.ZodOptional<z.ZodNumber>;
    }, "strip", z.ZodTypeAny, {
        requestsPerMinute?: number | undefined;
        tokensPerMinute?: number | undefined;
        requestsPerDay?: number | undefined;
    }, {
        requestsPerMinute?: number | undefined;
        tokensPerMinute?: number | undefined;
        requestsPerDay?: number | undefined;
    }>>;
    timeout: z.ZodDefault<z.ZodNumber>;
    retries: z.ZodDefault<z.ZodNumber>;
}, "strip", z.ZodTypeAny, {
    id: string;
    name: string;
    type: "openai" | "anthropic" | "google" | "local" | "custom";
    enabled: boolean;
    models: {
        id: string;
        name: string;
        maxTokens: number;
        supportsStreaming: boolean;
        supportsVision: boolean;
        supportsTools: boolean;
        costPer1KInput?: number | undefined;
        costPer1KOutput?: number | undefined;
    }[];
    timeout: number;
    retries: number;
    defaultModel?: string | undefined;
    apiKey?: string | undefined;
    baseUrl?: string | undefined;
    organization?: string | undefined;
    rateLimits?: {
        requestsPerMinute?: number | undefined;
        tokensPerMinute?: number | undefined;
        requestsPerDay?: number | undefined;
    } | undefined;
}, {
    id: string;
    name: string;
    type: "openai" | "anthropic" | "google" | "local" | "custom";
    models: {
        id: string;
        name: string;
        maxTokens: number;
        supportsStreaming?: boolean | undefined;
        supportsVision?: boolean | undefined;
        supportsTools?: boolean | undefined;
        costPer1KInput?: number | undefined;
        costPer1KOutput?: number | undefined;
    }[];
    enabled?: boolean | undefined;
    defaultModel?: string | undefined;
    apiKey?: string | undefined;
    baseUrl?: string | undefined;
    organization?: string | undefined;
    rateLimits?: {
        requestsPerMinute?: number | undefined;
        tokensPerMinute?: number | undefined;
        requestsPerDay?: number | undefined;
    } | undefined;
    timeout?: number | undefined;
    retries?: number | undefined;
}>;
export declare const ProvidersConfigSchema: z.ZodObject<{
    version: z.ZodDefault<z.ZodNumber>;
    defaultProvider: z.ZodOptional<z.ZodString>;
    providers: z.ZodArray<z.ZodObject<{
        id: z.ZodString;
        name: z.ZodString;
        type: z.ZodEnum<["openai", "anthropic", "google", "local", "custom"]>;
        enabled: z.ZodDefault<z.ZodBoolean>;
        defaultModel: z.ZodOptional<z.ZodString>;
        apiKey: z.ZodOptional<z.ZodString>;
        baseUrl: z.ZodOptional<z.ZodString>;
        organization: z.ZodOptional<z.ZodString>;
        models: z.ZodArray<z.ZodObject<{
            id: z.ZodString;
            name: z.ZodString;
            maxTokens: z.ZodNumber;
            supportsStreaming: z.ZodDefault<z.ZodBoolean>;
            supportsVision: z.ZodDefault<z.ZodBoolean>;
            supportsTools: z.ZodDefault<z.ZodBoolean>;
            costPer1KInput: z.ZodOptional<z.ZodNumber>;
            costPer1KOutput: z.ZodOptional<z.ZodNumber>;
        }, "strip", z.ZodTypeAny, {
            id: string;
            name: string;
            maxTokens: number;
            supportsStreaming: boolean;
            supportsVision: boolean;
            supportsTools: boolean;
            costPer1KInput?: number | undefined;
            costPer1KOutput?: number | undefined;
        }, {
            id: string;
            name: string;
            maxTokens: number;
            supportsStreaming?: boolean | undefined;
            supportsVision?: boolean | undefined;
            supportsTools?: boolean | undefined;
            costPer1KInput?: number | undefined;
            costPer1KOutput?: number | undefined;
        }>, "many">;
        rateLimits: z.ZodOptional<z.ZodObject<{
            requestsPerMinute: z.ZodOptional<z.ZodNumber>;
            tokensPerMinute: z.ZodOptional<z.ZodNumber>;
            requestsPerDay: z.ZodOptional<z.ZodNumber>;
        }, "strip", z.ZodTypeAny, {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        }, {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        }>>;
        timeout: z.ZodDefault<z.ZodNumber>;
        retries: z.ZodDefault<z.ZodNumber>;
    }, "strip", z.ZodTypeAny, {
        id: string;
        name: string;
        type: "openai" | "anthropic" | "google" | "local" | "custom";
        enabled: boolean;
        models: {
            id: string;
            name: string;
            maxTokens: number;
            supportsStreaming: boolean;
            supportsVision: boolean;
            supportsTools: boolean;
            costPer1KInput?: number | undefined;
            costPer1KOutput?: number | undefined;
        }[];
        timeout: number;
        retries: number;
        defaultModel?: string | undefined;
        apiKey?: string | undefined;
        baseUrl?: string | undefined;
        organization?: string | undefined;
        rateLimits?: {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        } | undefined;
    }, {
        id: string;
        name: string;
        type: "openai" | "anthropic" | "google" | "local" | "custom";
        models: {
            id: string;
            name: string;
            maxTokens: number;
            supportsStreaming?: boolean | undefined;
            supportsVision?: boolean | undefined;
            supportsTools?: boolean | undefined;
            costPer1KInput?: number | undefined;
            costPer1KOutput?: number | undefined;
        }[];
        enabled?: boolean | undefined;
        defaultModel?: string | undefined;
        apiKey?: string | undefined;
        baseUrl?: string | undefined;
        organization?: string | undefined;
        rateLimits?: {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        } | undefined;
        timeout?: number | undefined;
        retries?: number | undefined;
    }>, "many">;
}, "strip", z.ZodTypeAny, {
    version: number;
    providers: {
        id: string;
        name: string;
        type: "openai" | "anthropic" | "google" | "local" | "custom";
        enabled: boolean;
        models: {
            id: string;
            name: string;
            maxTokens: number;
            supportsStreaming: boolean;
            supportsVision: boolean;
            supportsTools: boolean;
            costPer1KInput?: number | undefined;
            costPer1KOutput?: number | undefined;
        }[];
        timeout: number;
        retries: number;
        defaultModel?: string | undefined;
        apiKey?: string | undefined;
        baseUrl?: string | undefined;
        organization?: string | undefined;
        rateLimits?: {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        } | undefined;
    }[];
    defaultProvider?: string | undefined;
}, {
    providers: {
        id: string;
        name: string;
        type: "openai" | "anthropic" | "google" | "local" | "custom";
        models: {
            id: string;
            name: string;
            maxTokens: number;
            supportsStreaming?: boolean | undefined;
            supportsVision?: boolean | undefined;
            supportsTools?: boolean | undefined;
            costPer1KInput?: number | undefined;
            costPer1KOutput?: number | undefined;
        }[];
        enabled?: boolean | undefined;
        defaultModel?: string | undefined;
        apiKey?: string | undefined;
        baseUrl?: string | undefined;
        organization?: string | undefined;
        rateLimits?: {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        } | undefined;
        timeout?: number | undefined;
        retries?: number | undefined;
    }[];
    version?: number | undefined;
    defaultProvider?: string | undefined;
}>;
export declare const ValidationResultSchema: z.ZodObject<{
    valid: z.ZodBoolean;
    errors: z.ZodArray<z.ZodString, "many">;
}, "strip", z.ZodTypeAny, {
    valid: boolean;
    errors: string[];
}, {
    valid: boolean;
    errors: string[];
}>;
export declare const ConnectionTestResultSchema: z.ZodObject<{
    success: z.ZodBoolean;
    latency: z.ZodOptional<z.ZodNumber>;
    error: z.ZodOptional<z.ZodString>;
    modelsAvailable: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
}, "strip", z.ZodTypeAny, {
    success: boolean;
    latency?: number | undefined;
    error?: string | undefined;
    modelsAvailable?: string[] | undefined;
}, {
    success: boolean;
    latency?: number | undefined;
    error?: string | undefined;
    modelsAvailable?: string[] | undefined;
}>;
export type ProviderType = z.infer<typeof ProviderType>;
export type ModelConfig = z.infer<typeof ModelConfigSchema>;
export type RateLimitConfig = z.infer<typeof RateLimitConfigSchema>;
export type ProviderConfig = z.infer<typeof ProviderConfigSchema>;
export type ProvidersConfig = z.infer<typeof ProvidersConfigSchema>;
export type ValidationResult = z.infer<typeof ValidationResultSchema>;
export type ConnectionTestResult = z.infer<typeof ConnectionTestResultSchema>;
export declare class ProviderError extends Error {
    constructor(message: string);
}
export declare class ProviderNotFoundError extends ProviderError {
    constructor(id: string);
}
export declare class ProviderValidationError extends ProviderError {
    errors: string[];
    constructor(errors: string[]);
}
export declare class ProviderAlreadyExistsError extends ProviderError {
    constructor(id: string);
}
export declare const CURRENT_VERSION = 1;
//# sourceMappingURL=types.d.ts.map