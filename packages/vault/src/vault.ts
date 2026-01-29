import {
  VaultConfig,
  EncryptedData,
  DEFAULT_VAULT_CONFIG,
  CURRENT_VERSION,
  VaultError,
  DecryptionError,
} from './types.js';
import {
  deriveKey,
  generateSalt,
  generateIV,
  encrypt,
  decrypt,
  secureClear,
  secureCompare,
} from './crypto.js';

export class Vault {
  private _config: VaultConfig;
  private _key: Buffer | null = null;
  private _salt: Buffer | null = null;
  private _initialized = false;

  constructor(config: Partial<VaultConfig> = {}) {
    this._config = { ...DEFAULT_VAULT_CONFIG, ...config };
  }

  async initialize(password: string): Promise<void> {
    if (this._initialized) {
      throw new VaultError('Vault is already initialized');
    }

    this._salt = generateSalt(this._config.saltLength);
    this._key = await deriveKey(password, this._salt, this._config);
    this._initialized = true;
  }

  async deriveKey(password: string, salt: Buffer): Promise<Buffer> {
    return deriveKey(password, salt, this._config);
  }

  async encrypt(plaintext: Buffer): Promise<EncryptedData> {
    this._ensureInitialized();
    
    const iv = generateIV();
    const { ciphertext, tag } = encrypt(plaintext, this._key!, iv);
    
    return {
      ciphertext,
      iv,
      salt: Buffer.from(this._salt!),
      tag,
      version: CURRENT_VERSION,
    };
  }

  async decrypt(encrypted: EncryptedData, password?: string): Promise<Buffer> {
    if (encrypted.version !== CURRENT_VERSION) {
      throw new DecryptionError(`Unsupported version: ${encrypted.version}`);
    }
    
    let key: Buffer;
    
    if (password) {
      key = await this.deriveKey(password, encrypted.salt);
    } else if (this._initialized && this._key && this._salt) {
      if (!secureCompare(this._salt, encrypted.salt)) {
        throw new DecryptionError('Salt mismatch - provide password to decrypt');
      }
      key = Buffer.from(this._key);
    } else {
      throw new VaultError('Vault not initialized and no password provided');
    }
    
    try {
      return decrypt(encrypted.ciphertext, key, encrypted.iv, encrypted.tag);
    } finally {
      secureClear(key);
    }
  }

  secureClear(data: Buffer): void {
    secureClear(data);
  }

  clearKey(): void {
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

  get isInitialized(): boolean {
    return this._initialized;
  }

  get salt(): Buffer | null {
    return this._salt ? Buffer.from(this._salt) : null;
  }

  getConfig(): VaultConfig {
    return { ...this._config };
  }

  private _ensureInitialized(): void {
    if (!this._initialized || !this._key) {
      throw new VaultError('Vault is not initialized. Call initialize() first.');
    }
  }
}

export async function createVault(password: string, config?: Partial<VaultConfig>): Promise<Vault> {
  const vault = new Vault(config);
  await vault.initialize(password);
  return vault;
}

export async function encryptWithPassword(
  plaintext: Buffer,
  password: string,
  config?: Partial<VaultConfig>
): Promise<EncryptedData> {
  const vault = await createVault(password, config);
  try {
    return await vault.encrypt(plaintext);
  } finally {
    vault.clearKey();
  }
}

export async function decryptWithPassword(
  encrypted: EncryptedData,
  password: string,
  config?: Partial<VaultConfig>
): Promise<Buffer> {
  const vault = new Vault(config);
  const key = await vault.deriveKey(password, encrypted.salt);
  try {
    return decrypt(encrypted.ciphertext, key, encrypted.iv, encrypted.tag);
  } finally {
    secureClear(key);
  }
}
