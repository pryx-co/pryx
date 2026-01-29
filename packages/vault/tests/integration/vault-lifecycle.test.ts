import { describe, it, expect, beforeEach } from 'vitest';
import { Vault } from '../../src/vault.js';
import { DecryptionError } from '../../src/types.js';

describe('Vault Lifecycle Integration', () => {
  let vault: Vault;

  beforeEach(() => {
    vault = new Vault();
  });

  it('should complete full lifecycle: init → encrypt → decrypt → clear', async () => {
    const password = 'master-password';
    const plaintext = Buffer.from('sensitive credential data');

    await vault.initialize(password);
    expect(vault.isInitialized).toBe(true);

    const encrypted = await vault.encrypt(plaintext);
    expect(encrypted.ciphertext.length).toBeGreaterThan(0);

    const decrypted = await vault.decrypt(encrypted);
    expect(decrypted.toString()).toBe(plaintext.toString());

    vault.clearKey();
    expect(vault.isInitialized).toBe(false);
  });

  it('should handle multiple encryption operations', async () => {
    await vault.initialize('password');

    const data1 = Buffer.from('first secret');
    const data2 = Buffer.from('second secret');
    const data3 = Buffer.from('third secret');

    const encrypted1 = await vault.encrypt(data1);
    const encrypted2 = await vault.encrypt(data2);
    const encrypted3 = await vault.encrypt(data3);

    const decrypted1 = await vault.decrypt(encrypted1);
    const decrypted2 = await vault.decrypt(encrypted2);
    const decrypted3 = await vault.decrypt(encrypted3);

    expect(decrypted1.toString()).toBe('first secret');
    expect(decrypted2.toString()).toBe('second secret');
    expect(decrypted3.toString()).toBe('third secret');
  });

  it('should use unique IV for each encryption', async () => {
    await vault.initialize('password');

    const data = Buffer.from('same data');
    const encrypted1 = await vault.encrypt(data);
    const encrypted2 = await vault.encrypt(data);

    expect(encrypted1.iv.toString('hex')).not.toBe(encrypted2.iv.toString('hex'));
    expect(encrypted1.ciphertext.toString('hex')).not.toBe(encrypted2.ciphertext.toString('hex'));
  });

  it('should handle concurrent encryption operations', async () => {
    await vault.initialize('password');

    const promises = Array.from({ length: 10 }, async (_, i) => {
      const data = Buffer.from(`data-${i}`);
      const encrypted = await vault.encrypt(data);
      const decrypted = await vault.decrypt(encrypted);
      return decrypted.toString();
    });

    const results = await Promise.all(promises);

    for (let i = 0; i < 10; i++) {
      expect(results[i]).toBe(`data-${i}`);
    }
  });

  it('should handle large data encryption', async () => {
    await vault.initialize('password');

    const largeData = Buffer.alloc(1024 * 1024, 0x42);
    const encrypted = await vault.encrypt(largeData);
    const decrypted = await vault.decrypt(encrypted);

    expect(decrypted.toString('hex')).toBe(largeData.toString('hex'));
  });

  it('should handle empty data', async () => {
    await vault.initialize('password');

    const emptyData = Buffer.alloc(0);
    const encrypted = await vault.encrypt(emptyData);
    const decrypted = await vault.decrypt(encrypted);

    expect(decrypted.length).toBe(0);
  });

  it('should handle unicode data', async () => {
    await vault.initialize('password');

    const unicodeData = Buffer.from('Hello 世界 🌍 مرحبا', 'utf8');
    const encrypted = await vault.encrypt(unicodeData);
    const decrypted = await vault.decrypt(encrypted);

    expect(decrypted.toString('utf8')).toBe('Hello 世界 🌍 مرحبا');
  });

  it('should fail decryption with wrong password', async () => {
    const correctPassword = 'correct-password';
    const wrongPassword = 'wrong-password';
    const plaintext = Buffer.from('secret');

    await vault.initialize(correctPassword);
    const encrypted = await vault.encrypt(plaintext);

    const wrongVault = new Vault();
    await wrongVault.initialize(wrongPassword);

    await expect(wrongVault.decrypt(encrypted)).rejects.toThrow(DecryptionError);
  });

  it('should fail decryption with tampered data', async () => {
    await vault.initialize('password');
    const plaintext = Buffer.from('secret');
    const encrypted = await vault.encrypt(plaintext);

    encrypted.ciphertext[0] ^= 0xFF;

    await expect(vault.decrypt(encrypted)).rejects.toThrow(DecryptionError);
  });

  it('should recover from error state', async () => {
    await vault.initialize('password');
    const plaintext = Buffer.from('test');
    const encrypted = await vault.encrypt(plaintext);

    encrypted.ciphertext[0] ^= 0xFF;
    await expect(vault.decrypt(encrypted)).rejects.toThrow();

    encrypted.ciphertext[0] ^= 0xFF;
    const decrypted = await vault.decrypt(encrypted);
    expect(decrypted.toString()).toBe('test');
  });
});
