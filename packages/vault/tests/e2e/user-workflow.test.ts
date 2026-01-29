import { describe, it, expect } from 'vitest';
import { Vault, encryptWithPassword, decryptWithPassword } from '../../src/vault.js';
import { DecryptionError } from '../../src/types.js';

describe('User Workflow E2E', () => {
  it('should complete full user workflow: create vault → store credential → retrieve credential', async () => {
    const masterPassword = 'MyStr0ng!Mast3r#P@ssw0rd';
    const credential = JSON.stringify({
      service: 'openai',
      apiKey: 'sk-1234567890abcdef',
      organization: 'org-test123',
    });

    const vault = new Vault();
    await vault.initialize(masterPassword);

    const encrypted = await vault.encrypt(Buffer.from(credential));

    const decrypted = await vault.decrypt(encrypted);
    const retrieved = JSON.parse(decrypted.toString());

    expect(retrieved.service).toBe('openai');
    expect(retrieved.apiKey).toBe('sk-1234567890abcdef');
    expect(retrieved.organization).toBe('org-test123');

    vault.clearKey();
  });

  it('should handle wrong password scenario', async () => {
    const correctPassword = 'CorrectP@ss123';
    const wrongPassword = 'WrongP@ss456';
    const secret = 'my-secret-data';

    const vault = new Vault();
    await vault.initialize(correctPassword);
    const encrypted = await vault.encrypt(Buffer.from(secret));
    vault.clearKey();

    const wrongVault = new Vault();
    await wrongVault.initialize(wrongPassword);

    await expect(wrongVault.decrypt(encrypted)).rejects.toThrow(DecryptionError);
  });

  it('should handle multiple credentials', async () => {
    const password = 'master-password';
    const credentials = [
      { service: 'openai', key: 'sk-openai-123' },
      { service: 'anthropic', key: 'sk-anthropic-456' },
      { service: 'google', key: 'sk-google-789' },
    ];

    const vault = new Vault();
    await vault.initialize(password);

    const encryptedCredentials = await Promise.all(
      credentials.map(async (cred) => ({
        service: cred.service,
        encrypted: await vault.encrypt(Buffer.from(JSON.stringify(cred))),
      }))
    );

    vault.clearKey();

    for (const { service, encrypted } of encryptedCredentials) {
      const decrypted = await decryptWithPassword(encrypted, password);
      const parsed = JSON.parse(decrypted.toString());
      expect(parsed.service).toBe(service);
    }
  });

  it('should encrypt and decrypt using convenience functions', async () => {
    const password = 'simple-password';
    const data = Buffer.from('test data');

    const encrypted = await encryptWithPassword(data, password);

    const decrypted = await decryptWithPassword(encrypted, password);

    expect(decrypted.toString()).toBe('test data');
  });

  it('should fail with wrong password using convenience functions', async () => {
    const correctPassword = 'correct';
    const wrongPassword = 'wrong';
    const data = Buffer.from('secret');

    const encrypted = await encryptWithPassword(data, correctPassword);

    await expect(decryptWithPassword(encrypted, wrongPassword)).rejects.toThrow(DecryptionError);
  });

  it('should handle binary data', async () => {
    const password = 'binary-test';
    const binaryData = Buffer.from([0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD]);

    const encrypted = await encryptWithPassword(binaryData, password);
    const decrypted = await decryptWithPassword(encrypted, password);

    expect(Buffer.compare(decrypted, binaryData)).toBe(0);
  });

  it('should handle special characters in password', async () => {
    const specialPassword = 'p@$$w0rd!#$%^*()_+-=[]{}|;:,.?<>';
    const data = Buffer.from('test');

    const encrypted = await encryptWithPassword(data, specialPassword);
    const decrypted = await decryptWithPassword(encrypted, specialPassword);

    expect(decrypted.toString()).toBe('test');
  });

  it('should handle very long password', async () => {
    const longPassword = 'a'.repeat(1000);
    const data = Buffer.from('test');

    const encrypted = await encryptWithPassword(data, longPassword);
    const decrypted = await decryptWithPassword(encrypted, longPassword);

    expect(decrypted.toString()).toBe('test');
  });

  it('should handle empty password', async () => {
    const emptyPassword = '';
    const data = Buffer.from('test');

    const encrypted = await encryptWithPassword(data, emptyPassword);
    const decrypted = await decryptWithPassword(encrypted, emptyPassword);

    expect(decrypted.toString()).toBe('test');
  });
});

describe('Performance Benchmarks', () => {
  it('should complete key derivation in reasonable time', async () => {
    const password = 'benchmark-password';
    const iterations = 5;

    const start = performance.now();

    for (let i = 0; i < iterations; i++) {
      const vault = new Vault();
      await vault.initialize(password);
      vault.clearKey();
    }

    const duration = performance.now() - start;
    const avgTime = duration / iterations;

    expect(avgTime).toBeLessThan(1000);
  });

  it('should encrypt data efficiently', async () => {
    const vault = new Vault();
    await vault.initialize('password');

    const data = Buffer.from('x'.repeat(10000));
    const iterations = 100;

    const start = performance.now();

    for (let i = 0; i < iterations; i++) {
      await vault.encrypt(data);
    }

    const duration = performance.now() - start;
    const avgTime = duration / iterations;

    expect(avgTime).toBeLessThan(10);
  });

  it('should decrypt data efficiently', async () => {
    const vault = new Vault();
    await vault.initialize('password');

    const data = Buffer.from('x'.repeat(10000));
    const encrypted = await vault.encrypt(data);
    const iterations = 100;

    const start = performance.now();

    for (let i = 0; i < iterations; i++) {
      await vault.decrypt(encrypted);
    }

    const duration = performance.now() - start;
    const avgTime = duration / iterations;

    expect(avgTime).toBeLessThan(5);
  });
});
