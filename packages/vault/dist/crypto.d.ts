import { VaultConfig, EncryptedData, SerializedEncryptedData } from './types.js';
export declare function deriveKey(password: string, salt: Buffer, config?: VaultConfig): Promise<Buffer>;
export declare function generateSalt(length?: number): Buffer;
export declare function generateIV(length?: number): Buffer;
export declare function encrypt(plaintext: Buffer, key: Buffer, iv: Buffer): {
    ciphertext: Buffer;
    tag: Buffer;
};
export declare function decrypt(ciphertext: Buffer, key: Buffer, iv: Buffer, tag: Buffer): Buffer;
export declare function secureClear(buffer: Buffer): void;
export declare function secureCompare(a: Buffer, b: Buffer): boolean;
export declare function serializeEncryptedData(data: EncryptedData): SerializedEncryptedData;
export declare function deserializeEncryptedData(data: SerializedEncryptedData): EncryptedData;
//# sourceMappingURL=crypto.d.ts.map