import { ProviderConfig, ValidationResult } from './types.js';
export declare function validateProviderConfig(config: unknown): ValidationResult;
export declare function assertValidProviderConfig(config: unknown): ProviderConfig;
export declare function isValidProviderId(id: string): boolean;
export declare function isValidProviderType(type: string): type is ProviderConfig['type'];
export declare function isValidUrl(url: string): boolean;
//# sourceMappingURL=validation.d.ts.map