import { DEFAULT_VAULT_CONFIG, CURRENT_VERSION, VaultError, DecryptionError, } from './types.js';
import { deriveKey, generateSalt, generateIV, encrypt, decrypt, secureClear, secureCompare, } from './crypto.js';
export class Vault {
    _config;
    _key = null;
    _salt = null;
    _initialized = false;
    constructor(config = {}) {
        this._config = { ...DEFAULT_VAULT_CONFIG, ...config };
    }
    async initialize(password) {
        if (this._initialized) {
            throw new VaultError('Vault is already initialized');
        }
        this._salt = generateSalt(this._config.saltLength);
        this._key = await deriveKey(password, this._salt, this._config);
        this._initialized = true;
    }
    async deriveKey(password, salt) {
        return deriveKey(password, salt, this._config);
    }
    async encrypt(plaintext) {
        this._ensureInitialized();
        const iv = generateIV();
        const { ciphertext, tag } = encrypt(plaintext, this._key, iv);
        return {
            ciphertext,
            iv,
            salt: Buffer.from(this._salt),
            tag,
            version: CURRENT_VERSION,
        };
    }
    async decrypt(encrypted, password) {
        if (encrypted.version !== CURRENT_VERSION) {
            throw new DecryptionError(`Unsupported version: ${encrypted.version}`);
        }
        let key;
        if (password) {
            key = await this.deriveKey(password, encrypted.salt);
        }
        else if (this._initialized && this._key && this._salt) {
            if (!secureCompare(this._salt, encrypted.salt)) {
                throw new DecryptionError('Salt mismatch - provide password to decrypt');
            }
            key = Buffer.from(this._key);
        }
        else {
            throw new VaultError('Vault not initialized and no password provided');
        }
        try {
            return decrypt(encrypted.ciphertext, key, encrypted.iv, encrypted.tag);
        }
        finally {
            secureClear(key);
        }
    }
    secureClear(data) {
        secureClear(data);
    }
    clearKey() {
        if (this._key) {
            secureClear(this._key);
            this._key = null;
        }
        if (this._salt) {
            secureClear(this._salt);
            this._salt = null;
        }
        this._initialized = false;
    }
    get isInitialized() {
        return this._initialized;
    }
    get salt() {
        return this._salt ? Buffer.from(this._salt) : null;
    }
    getConfig() {
        return { ...this._config };
    }
    _ensureInitialized() {
        if (!this._initialized || !this._key) {
            throw new VaultError('Vault is not initialized. Call initialize() first.');
        }
    }
}
export async function createVault(password, config) {
    const vault = new Vault(config);
    await vault.initialize(password);
    return vault;
}
export async function encryptWithPassword(plaintext, password, config) {
    const vault = await createVault(password, config);
    try {
        return await vault.encrypt(plaintext);
    }
    finally {
        vault.clearKey();
    }
}
export async function decryptWithPassword(encrypted, password, config) {
    const vault = new Vault(config);
    const key = await vault.deriveKey(password, encrypted.salt);
    try {
        return decrypt(encrypted.ciphertext, key, encrypted.iv, encrypted.tag);
    }
    finally {
        secureClear(key);
    }
}
//# sourceMappingURL=vault.js.map