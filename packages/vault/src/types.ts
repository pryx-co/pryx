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

export const DEFAULT_VAULT_CONFIG: VaultConfig = {
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
  constructor(message: string) {
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
  constructor(message: string = 'Data appears to be corrupted') {
    super(message);
    this.name = 'CorruptedDataError';
  }
}

export class DecryptionError extends VaultError {
  constructor(message: string = 'Decryption failed') {
    super(message);
    this.name = 'DecryptionError';
  }
}
