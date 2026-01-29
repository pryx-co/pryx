export const DEFAULT_VAULT_CONFIG = {
    algorithm: 'argon2id',
    memoryCost: 65536,
    timeCost: 3,
    parallelism: 4,
    saltLength: 32,
    keyLength: 32,
};
export const AES_GCM_IV_LENGTH = 12;
export const AES_GCM_TAG_LENGTH = 16;
export const CURRENT_VERSION = 1;
export class VaultError extends Error {
    constructor(message) {
        super(message);
        this.name = 'VaultError';
    }
}
export class InvalidPasswordError extends VaultError {
    constructor() {
        super('Invalid password provided');
        this.name = 'InvalidPasswordError';
    }
}
export class CorruptedDataError extends VaultError {
    constructor(message = 'Data appears to be corrupted') {
        super(message);
        this.name = 'CorruptedDataError';
    }
}
export class DecryptionError extends VaultError {
    constructor(message = 'Decryption failed') {
        super(message);
        this.name = 'DecryptionError';
    }
}
//# sourceMappingURL=types.js.map