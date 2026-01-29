import { VaultConfig, EncryptedData } from './types.js';
export declare class Vault {
    private _config;
    private _key;
    private _salt;
    private _initialized;
    constructor(config?: Partial<VaultConfig>);
    initialize(password: string): Promise<void>;
    deriveKey(password: string, salt: Buffer): Promise<Buffer>;
    encrypt(plaintext: Buffer): Promise<EncryptedData>;
    decrypt(encrypted: EncryptedData, password?: string): Promise<Buffer>;
    secureClear(data: Buffer): void;
    clearKey(): void;
    get isInitialized(): boolean;
    get salt(): Buffer | null;
    getConfig(): VaultConfig;
    private _ensureInitialized;
}
export declare function createVault(password: string, config?: Partial<VaultConfig>): Promise<Vault>;
export declare function encryptWithPassword(plaintext: Buffer, password: string, config?: Partial<VaultConfig>): Promise<EncryptedData>;
export declare function decryptWithPassword(encrypted: EncryptedData, password: string, config?: Partial<VaultConfig>): Promise<Buffer>;
//# sourceMappingURL=vault.d.ts.map