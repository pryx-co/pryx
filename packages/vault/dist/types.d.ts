export interface VaultConfig {
    algorithm: 'argon2id' | 'pbkdf2';
    memoryCost: number;
    timeCost: number;
    parallelism: number;
    saltLength: number;
    keyLength: number;
}
export interface EncryptedData {
    ciphertext: Buffer;
    iv: Buffer;
    salt: Buffer;
    tag: Buffer;
    version: number;
}
export interface SerializedEncryptedData {
    ciphertext: string;
    iv: string;
    salt: string;
    tag: string;
    version: number;
}
export declare const DEFAULT_VAULT_CONFIG: VaultConfig;
export declare const AES_GCM_IV_LENGTH = 12;
export declare const AES_GCM_TAG_LENGTH = 16;
export declare const CURRENT_VERSION = 1;
export declare class VaultError extends Error {
    constructor(message: string);
}
export declare class InvalidPasswordError extends VaultError {
    constructor();
}
export declare class CorruptedDataError extends VaultError {
    constructor(message?: string);
}
export declare class DecryptionError extends VaultError {
    constructor(message?: string);
}
//# sourceMappingURL=types.d.ts.map