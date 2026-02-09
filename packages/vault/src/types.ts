/**
 * Vault configuration interface
 */
export interface VaultConfig {
  algorithm: 'argon2id' | 'pbkdf2';
  memoryCost: number;
  timeCost: number;
  parallelism: number;
  saltLength: number;
  keyLength: number;
}

/**
 * Encrypted data structure
 */
export interface EncryptedData {
  ciphertext: Buffer;
  iv: Buffer;
  salt: Buffer;
  tag: Buffer;
  version: number;
}

/**
 * Serialized encrypted data structure
 */
export interface SerializedEncryptedData {
  ciphertext: string;
  iv: string;
  salt: string;
  tag: string;
  version: number;
}

/**
 * Default vault configuration
 */
export const DEFAULT_VAULT_CONFIG: VaultConfig = {
  algorithm: 'argon2id',
  memoryCost: 65536,
  timeCost: 3,
  parallelism: 4,
  saltLength: 32,
  keyLength: 32,
};

/**
 * AES-GCM IV length in bytes
 */
export const AES_GCM_IV_LENGTH = 12;

/**
 * AES-GCM tag length in bytes
 */
export const AES_GCM_TAG_LENGTH = 16;

/**
 * Current encryption version
 */
export const CURRENT_VERSION = 1;

/**
 * Base vault error class
 */
export class VaultError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'VaultError';
  }
}

/**
 * Error thrown when an invalid password is provided
 */
export class InvalidPasswordError extends VaultError {
  constructor() {
    super('Invalid password provided');
    this.name = 'InvalidPasswordError';
  }
}

/**
 * Error thrown when data appears to be corrupted
 */
export class CorruptedDataError extends VaultError {
  constructor(message: string = 'Data appears to be corrupted') {
    super(message);
    this.name = 'CorruptedDataError';
  }
}

/**
 * Error thrown when decryption fails
 */
export class DecryptionError extends VaultError {
  constructor(message: string = 'Decryption failed') {
    super(message);
    this.name = 'DecryptionError';
  }
}
