/**
 * Vault Module
 *
 * Provides high-level vault functionality for encryption and decryption
 * with secure key management.
 */

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

/**
 * Main Vault class for encryption operations
 */
export class Vault {
  private _config: VaultConfig;
  private _key: Buffer | null = null;
  private _salt: Buffer | null = null;
  private _initialized = false;

  /**
   * Creates a new Vault instance
   * @param config - Optional vault configuration
   */
  constructor(config: Partial<VaultConfig> = {}) {
    this._config = { ...DEFAULT_VAULT_CONFIG, ...config };
  }

  /**
   * Initializes the vault with a password
   * @param password - The master password
   * @throws VaultError if already initialized
   */
  async initialize(password: string): Promise<void> {
    if (this._initialized) {
      throw new VaultError('Vault is already initialized');
    }

    this._salt = generateSalt(this._config.saltLength);
    this._key = await deriveKey(password, this._salt, this._config);
    this._initialized = true;
  }

  /**
   * Derives a key from password and salt
   * @param password - The password
   * @param salt - The salt
   * @returns The derived key
   */
  async deriveKey(password: string, salt: Buffer): Promise<Buffer> {
    return deriveKey(password, salt, this._config);
  }

  /**
   * Encrypts plaintext data
   * @param plaintext - Data to encrypt
   * @returns Encrypted data object
   * @throws VaultError if not initialized
   */
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

  /**
   * Decrypts encrypted data
   * @param encrypted - Encrypted data object
   * @param password - Optional password (uses session key if not provided)
   * @returns Decrypted plaintext
   * @throws DecryptionError if decryption fails
   * @throws VaultError if not initialized and no password provided
   */
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

  /**
   * Securely clears data from memory
   * @param data - Buffer to clear
   */
  secureClear(data: Buffer): void {
    secureClear(data);
  }

  /**
   * Clears the encryption key from memory
   */
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

  /**
   * Checks if the vault is initialized
   * @returns True if initialized
   */
  get isInitialized(): boolean {
    return this._initialized;
  }

  /**
   * Gets the salt (if initialized)
   * @returns Salt buffer or null
   */
  get salt(): Buffer | null {
    return this._salt ? Buffer.from(this._salt) : null;
  }

  /**
   * Gets the vault configuration
   * @returns Copy of the configuration
   */
  getConfig(): VaultConfig {
    return { ...this._config };
  }

  /**
   * Ensures the vault is initialized
   * @throws VaultError if not initialized
   */
  private _ensureInitialized(): void {
    if (!this._initialized || !this._key) {
      throw new VaultError('Vault is not initialized. Call initialize() first.');
    }
  }
}

/**
 * Creates and initializes a new vault
 * @param password - Master password
 * @param config - Optional configuration
 * @returns Initialized vault instance
 */
export async function createVault(password: string, config?: Partial<VaultConfig>): Promise<Vault> {
  const vault = new Vault(config);
  await vault.initialize(password);
  return vault;
}

/**
 * Encrypts data with a password (convenience function)
 * @param plaintext - Data to encrypt
 * @param password - Password for encryption
 * @param config - Optional vault configuration
 * @returns Encrypted data
 */
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

/**
 * Decrypts data with a password (convenience function)
 * @param encrypted - Encrypted data
 * @param password - Password for decryption
 * @param config - Optional vault configuration
 * @returns Decrypted plaintext
 */
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
