import { describe, it, expect, beforeEach } from 'vitest';
import {
  deriveKey,
  generateSalt,
  generateIV,
  encrypt,
  decrypt,
  secureClear,
  secureCompare,
  serializeEncryptedData,
  deserializeEncryptedData,
} from '../../src/crypto.js';
import { DEFAULT_VAULT_CONFIG, DecryptionError, VaultError, InvalidPasswordError, CorruptedDataError } from '../../src/types.js';

describe('deriveKey', () => {
  it('should derive 32-byte key from password', async () => {
    const password = 'test-password';
    const salt = generateSalt();
    
    const key = await deriveKey(password, salt);
    
    expect(key).toBeInstanceOf(Buffer);
    expect(key.length).toBe(32);
  });

  it('should derive same key with same password and salt', async () => {
    const password = 'test-password';
    const salt = generateSalt();
    
    const key1 = await deriveKey(password, salt);
    const key2 = await deriveKey(password, salt);
    
    expect(key1.toString('hex')).toBe(key2.toString('hex'));
  });

  it('should derive different keys with different passwords', async () => {
    const salt = generateSalt();
    
    const key1 = await deriveKey('password1', salt);
    const key2 = await deriveKey('password2', salt);
    
    expect(key1.toString('hex')).not.toBe(key2.toString('hex'));
  });

  it('should derive different keys with different salts', async () => {
    const password = 'test-password';
    const salt1 = generateSalt();
    const salt2 = generateSalt();
    
    const key1 = await deriveKey(password, salt1);
    const key2 = await deriveKey(password, salt2);
    
    expect(key1.toString('hex')).not.toBe(key2.toString('hex'));
  });

  it('should use custom config parameters', async () => {
    const password = 'test-password';
    const salt = generateSalt();
    const config = {
      ...DEFAULT_VAULT_CONFIG,
      memoryCost: 32768,
      timeCost: 2,
    };
    
    const key = await deriveKey(password, salt, config);
    
    expect(key.length).toBe(32);
  });

  it('should throw on unsupported algorithm', async () => {
    const password = 'test-password';
    const salt = generateSalt();
    const config = {
      ...DEFAULT_VAULT_CONFIG,
      algorithm: 'pbkdf2' as const,
    };
    
    await expect(deriveKey(password, salt, config)).rejects.toThrow('Unsupported algorithm');
  });
});

describe('generateSalt', () => {
  it('should generate salt of default length', () => {
    const salt = generateSalt();
    
    expect(salt).toBeInstanceOf(Buffer);
    expect(salt.length).toBe(32);
  });

  it('should generate salt of custom length', () => {
    const salt = generateSalt(16);
    
    expect(salt.length).toBe(16);
  });

  it('should generate unique salts', () => {
    const salt1 = generateSalt();
    const salt2 = generateSalt();
    
    expect(salt1.toString('hex')).not.toBe(salt2.toString('hex'));
  });
});

describe('generateIV', () => {
  it('should generate IV of default length (12 bytes)', () => {
    const iv = generateIV();
    
    expect(iv).toBeInstanceOf(Buffer);
    expect(iv.length).toBe(12);
  });

  it('should generate unique IVs', () => {
    const iv1 = generateIV();
    const iv2 = generateIV();
    
    expect(iv1.toString('hex')).not.toBe(iv2.toString('hex'));
  });
});

describe('encrypt', () => {
  it('should encrypt data and return ciphertext and tag', () => {
    const plaintext = Buffer.from('Hello, World!');
    const key = Buffer.alloc(32, 0x42);
    const iv = generateIV();
    
    const result = encrypt(plaintext, key, iv);
    
    expect(result.ciphertext).toBeInstanceOf(Buffer);
    expect(result.ciphertext.length).toBeGreaterThan(0);
    expect(result.tag).toBeInstanceOf(Buffer);
    expect(result.tag.length).toBe(16);
  });

  it('should throw on invalid key length', () => {
    const plaintext = Buffer.from('test');
    const key = Buffer.alloc(16, 0x42);
    const iv = generateIV();
    
    expect(() => encrypt(plaintext, key, iv)).toThrow('Invalid key length');
  });

  it('should throw on invalid IV length', () => {
    const plaintext = Buffer.from('test');
    const key = Buffer.alloc(32, 0x42);
    const iv = Buffer.alloc(16, 0x42);
    
    expect(() => encrypt(plaintext, key, iv)).toThrow('Invalid IV length');
  });
});

describe('decrypt', () => {
  it('should decrypt encrypted data correctly', () => {
    const plaintext = Buffer.from('Hello, World!');
    const key = Buffer.alloc(32, 0x42);
    const iv = generateIV();
    
    const encrypted = encrypt(plaintext, key, iv);
    const decrypted = decrypt(encrypted.ciphertext, key, iv, encrypted.tag);
    
    expect(decrypted.toString()).toBe(plaintext.toString());
  });

  it('should throw on wrong key', () => {
    const plaintext = Buffer.from('secret');
    const key1 = Buffer.alloc(32, 0x42);
    const key2 = Buffer.alloc(32, 0x24);
    const iv = generateIV();
    
    const encrypted = encrypt(plaintext, key1, iv);
    
    expect(() => decrypt(encrypted.ciphertext, key2, iv, encrypted.tag)).toThrow(DecryptionError);
  });

  it('should throw on tampered ciphertext', () => {
    const plaintext = Buffer.from('secret');
    const key = Buffer.alloc(32, 0x42);
    const iv = generateIV();
    
    const encrypted = encrypt(plaintext, key, iv);
    encrypted.ciphertext[0] ^= 0xFF;
    
    expect(() => decrypt(encrypted.ciphertext, key, iv, encrypted.tag)).toThrow(DecryptionError);
  });

  it('should throw on tampered tag', () => {
    const plaintext = Buffer.from('secret');
    const key = Buffer.alloc(32, 0x42);
    const iv = generateIV();
    
    const encrypted = encrypt(plaintext, key, iv);
    encrypted.tag[0] ^= 0xFF;
    
    expect(() => decrypt(encrypted.ciphertext, key, iv, encrypted.tag)).toThrow(DecryptionError);
  });

  it('should throw on invalid key length', () => {
    const ciphertext = Buffer.from('test');
    const key = Buffer.alloc(16, 0x42);
    const iv = generateIV();
    const tag = Buffer.alloc(16, 0x42);
    
    expect(() => decrypt(ciphertext, key, iv, tag)).toThrow('Invalid key length');
  });
});

describe('secureClear', () => {
  it('should clear buffer contents', () => {
    const buffer = Buffer.from('sensitive data');
    
    secureClear(buffer);
    
    expect(buffer.toString()).toBe('\x00'.repeat(buffer.length));
  });
});

describe('secureCompare', () => {
  it('should return true for identical buffers', () => {
    const buf1 = Buffer.from('test');
    const buf2 = Buffer.from('test');
    
    expect(secureCompare(buf1, buf2)).toBe(true);
  });

  it('should return false for different buffers', () => {
    const buf1 = Buffer.from('test1');
    const buf2 = Buffer.from('test2');
    
    expect(secureCompare(buf1, buf2)).toBe(false);
  });

  it('should return false for different lengths', () => {
    const buf1 = Buffer.from('test');
    const buf2 = Buffer.from('testing');
    
    expect(secureCompare(buf1, buf2)).toBe(false);
  });
});

describe('serializeEncryptedData', () => {
  it('should serialize to base64 strings', () => {
    const data = {
      ciphertext: Buffer.from('ciphertext'),
      iv: Buffer.from('iv'),
      salt: Buffer.from('salt'),
      tag: Buffer.from('tag'),
      version: 1,
    };
    
    const serialized = serializeEncryptedData(data);
    
    expect(typeof serialized.ciphertext).toBe('string');
    expect(typeof serialized.iv).toBe('string');
    expect(typeof serialized.salt).toBe('string');
    expect(typeof serialized.tag).toBe('string');
    expect(serialized.version).toBe(1);
  });
});

describe('deserializeEncryptedData', () => {
  it('should deserialize from base64 strings', () => {
    const original = {
      ciphertext: Buffer.from('ciphertext'),
      iv: Buffer.from('iv'),
      salt: Buffer.from('salt'),
      tag: Buffer.from('tag'),
      version: 1,
    };
    
    const serialized = serializeEncryptedData(original);
    const deserialized = deserializeEncryptedData(serialized);
    
    expect(deserialized.ciphertext.toString()).toBe(original.ciphertext.toString());
    expect(deserialized.iv.toString()).toBe(original.iv.toString());
    expect(deserialized.salt.toString()).toBe(original.salt.toString());
    expect(deserialized.tag.toString()).toBe(original.tag.toString());
    expect(deserialized.version).toBe(original.version);
  });
});

describe('Error classes', () => {
  it('should create VaultError', () => {
    const error = new VaultError('test message');
    expect(error.message).toBe('test message');
    expect(error.name).toBe('VaultError');
  });

  it('should create InvalidPasswordError', () => {
    const error = new InvalidPasswordError();
    expect(error.message).toBe('Invalid password provided');
    expect(error.name).toBe('InvalidPasswordError');
  });

  it('should create CorruptedDataError with default message', () => {
    const error = new CorruptedDataError();
    expect(error.message).toBe('Data appears to be corrupted');
    expect(error.name).toBe('CorruptedDataError');
  });

  it('should create CorruptedDataError with custom message', () => {
    const error = new CorruptedDataError('custom error');
    expect(error.message).toBe('custom error');
    expect(error.name).toBe('CorruptedDataError');
  });

  it('should create DecryptionError with default message', () => {
    const error = new DecryptionError();
    expect(error.message).toBe('Decryption failed');
    expect(error.name).toBe('DecryptionError');
  });

  it('should create DecryptionError with custom message', () => {
    const error = new DecryptionError('custom decryption error');
    expect(error.message).toBe('custom decryption error');
    expect(error.name).toBe('DecryptionError');
  });
});
