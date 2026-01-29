import { describe, it, expect, beforeEach } from 'vitest';
import { Vault, createVault, encryptWithPassword, decryptWithPassword } from '../../src/vault.js';
import { VaultError, DecryptionError } from '../../src/types.js';

describe('Vault', () => {
  let vault: Vault;

  beforeEach(() => {
    vault = new Vault();
  });

  describe('constructor', () => {
    it('should create vault with default config', () => {
      expect(vault).toBeInstanceOf(Vault);
      expect(vault.isInitialized).toBe(false);
    });

    it('should create vault with custom config', () => {
      const customVault = new Vault({ memoryCost: 32768 });
      expect(customVault).toBeInstanceOf(Vault);
    });
  });

  describe('initialize', () => {
    it('should initialize with password', async () => {
      await vault.initialize('test-password');
      
      expect(vault.isInitialized).toBe(true);
      expect(vault.salt).toBeInstanceOf(Buffer);
      expect(vault.salt!.length).toBe(32);
    });

    it('should throw when already initialized', async () => {
      await vault.initialize('test-password');
      
      await expect(vault.initialize('another-password')).rejects.toThrow(VaultError);
    });
  });

  describe('deriveKey', () => {
    it('should derive key from password and salt', async () => {
      const salt = Buffer.alloc(32, 0x42);
      const key = await vault.deriveKey('test-password', salt);
      
      expect(key).toBeInstanceOf(Buffer);
      expect(key.length).toBe(32);
    });
  });

  describe('encrypt', () => {
    it('should encrypt plaintext', async () => {
      await vault.initialize('test-password');
      const plaintext = Buffer.from('Hello, World!');
      
      const encrypted = await vault.encrypt(plaintext);
      
      expect(encrypted.ciphertext).toBeInstanceOf(Buffer);
      expect(encrypted.iv).toBeInstanceOf(Buffer);
      expect(encrypted.salt).toBeInstanceOf(Buffer);
      expect(encrypted.tag).toBeInstanceOf(Buffer);
      expect(encrypted.version).toBe(1);
    });

    it('should throw when not initialized', async () => {
      const plaintext = Buffer.from('test');
      
      await expect(vault.encrypt(plaintext)).rejects.toThrow(VaultError);
    });
  });

  describe('decrypt', () => {
    it('should decrypt encrypted data', async () => {
      await vault.initialize('test-password');
      const plaintext = Buffer.from('Hello, World!');
      const encrypted = await vault.encrypt(plaintext);
      
      const decrypted = await vault.decrypt(encrypted);
      
      expect(decrypted.toString()).toBe(plaintext.toString());
    });

    it('should throw when not initialized', async () => {
      const encrypted = {
        ciphertext: Buffer.from('test'),
        iv: Buffer.alloc(12),
        salt: Buffer.alloc(32),
        tag: Buffer.alloc(16),
        version: 1,
      };
      
      await expect(vault.decrypt(encrypted)).rejects.toThrow(VaultError);
    });

    it('should throw on unsupported version', async () => {
      await vault.initialize('test-password');
      const encrypted = {
        ciphertext: Buffer.from('test'),
        iv: Buffer.alloc(12),
        salt: Buffer.alloc(32),
        tag: Buffer.alloc(16),
        version: 999,
      };
      
      await expect(vault.decrypt(encrypted)).rejects.toThrow(DecryptionError);
    });
  });

  describe('clearKey', () => {
    it('should clear key and reset state', async () => {
      await vault.initialize('test-password');
      expect(vault.isInitialized).toBe(true);
      
      vault.clearKey();
      
      expect(vault.isInitialized).toBe(false);
    });
  });

  describe('secureClear', () => {
    it('should clear data', async () => {
      const data = Buffer.from('sensitive');
      
      vault.secureClear(data);
      
      expect(data.toString()).toBe('\x00'.repeat(data.length));
    });
  });

  describe('getConfig', () => {
    it('should return config copy', () => {
      const config = vault.getConfig();
      
      expect(config.memoryCost).toBe(65536);
      expect(config.timeCost).toBe(3);
      expect(config.parallelism).toBe(4);
    });
  });
});

describe('createVault', () => {
  it('should create and initialize vault', async () => {
    const vault = await createVault('test-password');
    
    expect(vault.isInitialized).toBe(true);
  });
});

describe('encryptWithPassword', () => {
  it('should encrypt with password', async () => {
    const plaintext = Buffer.from('secret');
    
    const encrypted = await encryptWithPassword(plaintext, 'test-password');
    
    expect(encrypted.ciphertext).toBeInstanceOf(Buffer);
  });
});

describe('decryptWithPassword', () => {
  it('should decrypt with correct password', async () => {
    const plaintext = Buffer.from('secret');
    const encrypted = await encryptWithPassword(plaintext, 'test-password');
    
    const decrypted = await decryptWithPassword(encrypted, 'test-password');
    
    expect(decrypted.toString()).toBe(plaintext.toString());
  });

  it('should throw with wrong password', async () => {
    const plaintext = Buffer.from('secret');
    const encrypted = await encryptWithPassword(plaintext, 'correct-password');
    
    await expect(decryptWithPassword(encrypted, 'wrong-password')).rejects.toThrow(DecryptionError);
  });
});
