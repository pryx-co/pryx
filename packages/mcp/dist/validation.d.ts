import { MCPServerConfig, ValidationResult, TransportType } from './types.js';
export declare function validateMCPServerConfig(config: unknown): ValidationResult;
export declare function assertValidMCPServerConfig(config: unknown): MCPServerConfig;
export declare function isValidServerId(id: string): boolean;
export declare function isValidTransportType(type: string): type is TransportType;
export declare function isValidUrl(url: string): boolean;
export declare function isValidWebSocketUrl(url: string): boolean;
export declare function calculateBackoff(attempt: number, baseMs?: number): number;
//# sourceMappingURL=validation.d.ts.map