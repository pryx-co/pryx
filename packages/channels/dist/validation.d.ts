import { ChannelConfig, ValidationResult, ChannelType } from './types.js';
export declare function validateChannelConfig(config: unknown): ValidationResult;
export declare function assertValidChannelConfig(config: unknown): ChannelConfig;
export declare function isValidChannelId(id: string): boolean;
export declare function isValidChannelType(type: string): type is ChannelType;
export declare function matchesFilterPatterns(message: string, patterns: string[]): boolean;
export declare function isUserAllowed(userId: string, allowedList: string[], blockedList: string[]): boolean;
//# sourceMappingURL=validation.d.ts.map