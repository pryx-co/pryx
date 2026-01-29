import { randomBytes, createCipheriv, createDecipheriv, timingSafeEqual } from 'crypto';
import {
  VaultConfig,
  EncryptedData,
  SerializedEncryptedData,
  DEFAULT_VAULT_CONFIG,
  AES_GCM_IV_LENGTH,
  AES_GCM_TAG_LENGTH,
  DecryptionError,
} from './types.js';

const AES_ALGORITHM = 'aes-256-gcm';

export async function deriveKey(password: string, salt: Buffer, config: VaultConfig = DEFAULT_VAULT_CONFIG): Promise<Buffer> {
  if (config.algorithm === 'argon2id') {
    const argon2 = await import('argon2');
    const hashResult = await argon2.hash(password, {
      type: argon2.argon2id,
      salt,
      memoryCost: config.memoryCost,
      timeCost: config.timeCost,
      parallelism: config.parallelism,
      hashLength: config.keyLength,
      raw: true,
    });
    return hashResult;
  }
  
  throw new Error(`Unsupported algorithm: ${config.algorithm}`);
}

export function generateSalt(length: number = DEFAULT_VAULT_CONFIG.saltLength): Buffer {
  return randomBytes(length);
}

export function generateIV(length: number = AES_GCM_IV_LENGTH): Buffer {
  return randomBytes(length);
}

export function encrypt(plaintext: Buffer, key: Buffer, iv: Buffer): { ciphertext: Buffer; tag: Buffer } {
  if (key.length !== 32) {
    throw new Error(`Invalid key length: ${key.length}. Expected 32 bytes.`);
  }
  
  if (iv.length !== AES_GCM_IV_LENGTH) {
    throw new Error(`Invalid IV length: ${iv.length}. Expected ${AES_GCM_IV_LENGTH} bytes.`);
  }
  
  const cipher = createCipheriv(AES_ALGORITHM, key, iv);
  const ciphertext = Buffer.concat([cipher.update(plaintext), cipher.final()]);
  const tag = cipher.getAuthTag();
  
  return { ciphertext, tag };
}

export function decrypt(ciphertext: Buffer, key: Buffer, iv: Buffer, tag: Buffer): Buffer {
  if (key.length !== 32) {
    throw new DecryptionError(`Invalid key length: ${key.length}`);
  }
  
  if (iv.length !== AES_GCM_IV_LENGTH) {
    throw new DecryptionError(`Invalid IV length: ${iv.length}`);
  }
  
  if (tag.length !== AES_GCM_TAG_LENGTH) {
    throw new DecryptionError(`Invalid tag length: ${tag.length}`);
  }
  
  try {
    const decipher = createDecipheriv(AES_ALGORITHM, key, iv);
    decipher.setAuthTag(tag);
    const plaintext = Buffer.concat([decipher.update(ciphertext), decipher.final()]);
    return plaintext;
  } catch (error) {
    throw new DecryptionError('Authentication failed - data may be corrupted or key is incorrect');
  }
}

export function secureClear(buffer: Buffer): void {
  buffer.fill(0);
}

export function secureCompare(a: Buffer, b: Buffer): boolean {
  if (a.length !== b.length) {
    return false;
  }
  return timingSafeEqual(a, b);
}

export function serializeEncryptedData(data: EncryptedData): SerializedEncryptedData {
  return {
    ciphertext: data.ciphertext.toString('base64'),
    iv: data.iv.toString('base64'),
    salt: data.salt.toString('base64'),
    tag: data.tag.toString('base64'),
    version: data.version,
  };
}

export function deserializeEncryptedData(data: SerializedEncryptedData): EncryptedData {
  return {
    ciphertext: Buffer.from(data.ciphertext, 'base64'),
    iv: Buffer.from(data.iv, 'base64'),
    salt: Buffer.from(data.salt, 'base64'),
    tag: Buffer.from(data.tag, 'base64'),
    version: data.version,
  };
}
