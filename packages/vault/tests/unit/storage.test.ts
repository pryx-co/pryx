import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { VaultStorage, createVaultStorage } from '../../src/storage';
import { FileNotFoundError, CorruptedVaultError, EntryNotFoundError, DuplicateEntryError, VaultFile } from '../../src/storage-types';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

describe('VaultStorage', () => {
  let tempDir: string;
  let vaultPath: string;
  let backupDir: string;
  let storage: VaultStorage;
  const password = 'test-password-123';

  beforeEach(() => {
    tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'vault-storage-test-'));
    vaultPath = path.join(tempDir, 'vault.dat');
    backupDir = path.join(tempDir, 'backups');
    storage = createVaultStorage(backupDir);
  });

  afterEach(() => {
    fs.rmSync(tempDir, { recursive: true, force: true });
  });

  describe('createEmptyVault', () => {
    it('should create empty vault with correct structure', () => {
      const vault = storage.createEmptyVault();

      expect(vault.version).toBe(1);
      expect(vault.entries).toEqual([]);
      expect(vault.metadata).toBeDefined();
      expect(vault.metadata.algorithm).toBe('argon2id+aes-256-gcm');
      expect(vault.metadata.salt).toBeDefined();
      expect(vault.createdAt).toBeDefined();
      expect(vault.updatedAt).toBeDefined();
    });
  });

  describe('save and load', () => {
    it('should save and load vault successfully', async () => {
      const vault = storage.createEmptyVault();
      await storage.save(vaultPath, vault, password);

      const loaded = await storage.load(vaultPath, password);

      expect(loaded.version).toBe(vault.version);
      expect(loaded.metadata.salt).toBe(vault.metadata.salt);
      expect(loaded.entries).toEqual([]);
    });

    it('should throw FileNotFoundError for non-existent file', async () => {
      await expect(storage.load('/nonexistent/vault.dat', password)).rejects.toThrow(FileNotFoundError);
    });

    it('should throw CorruptedVaultError for invalid JSON', async () => {
      fs.writeFileSync(vaultPath, 'invalid json {', { mode: 0o600 });
      await expect(storage.load(vaultPath, password)).rejects.toThrow(CorruptedVaultError);
    });

    it('should create parent directories when saving', async () => {
      const nestedPath = path.join(tempDir, 'nested', 'deep', 'vault.dat');
      const vault = storage.createEmptyVault();

      await storage.save(nestedPath, vault, password);

      expect(fs.existsSync(nestedPath)).toBe(true);
    });

    it('should set file permissions to 0o600', async () => {
      const vault = storage.createEmptyVault();
      await storage.save(vaultPath, vault, password);

      const stats = fs.statSync(vaultPath);
      const mode = stats.mode & 0o777;
      expect(mode).toBe(0o600);
    });
  });

  describe('addEntry', () => {
    it('should add entry to vault', async () => {
      const vault = storage.createEmptyVault();
      const entryData = {
        type: 'credential' as const,
        name: 'Test Entry',
        data: { username: 'test', password: 'secret' },
      };

      const entry = await storage.addEntry(vault, entryData, password);

      expect(entry.id).toBeDefined();
      expect(entry.name).toBe('Test Entry');
      expect(entry.type).toBe('credential');
      expect(entry.encryptedData).toBeDefined();
      expect(entry.iv).toBeDefined();
      expect(entry.tag).toBeDefined();
      expect(vault.entries).toHaveLength(1);
    });

    it('should use provided id if given', async () => {
      const vault = storage.createEmptyVault();
      const entryData = {
        id: 'custom-id',
        type: 'credential' as const,
        name: 'Test Entry',
        data: { username: 'test' },
      };

      const entry = await storage.addEntry(vault, entryData, password);

      expect(entry.id).toBe('custom-id');
    });

    it('should throw DuplicateEntryError for duplicate id', async () => {
      const vault = storage.createEmptyVault();
      const entryData = {
        id: 'duplicate-id',
        type: 'credential' as const,
        name: 'Test Entry',
        data: { username: 'test' },
      };

      await storage.addEntry(vault, entryData, password);
      await expect(storage.addEntry(vault, entryData, password)).rejects.toThrow(DuplicateEntryError);
    });
  });

  describe('getEntry', () => {
    it('should retrieve and decrypt entry', async () => {
      const vault = storage.createEmptyVault();
      const entryData = {
        type: 'credential' as const,
        name: 'Test Entry',
        data: { username: 'test', password: 'secret' },
      };

      const entry = await storage.addEntry(vault, entryData, password);
      const retrieved = await storage.getEntry(vault, entry.id, password);

      expect(retrieved.id).toBe(entry.id);
      expect(retrieved.type).toBe('credential');
      expect(retrieved.name).toBe('Test Entry');
      expect(retrieved.data).toEqual({ username: 'test', password: 'secret' });
    });

    it('should update access count and last accessed', async () => {
      const vault = storage.createEmptyVault();
      const entryData = {
        type: 'credential' as const,
        name: 'Test Entry',
        data: { username: 'test' },
      };

      const entry = await storage.addEntry(vault, entryData, password);
      expect(entry.accessCount).toBe(0);
      expect(entry.lastAccessedAt).toBeUndefined();

      await storage.getEntry(vault, entry.id, password);

      expect(vault.entries[0].accessCount).toBe(1);
      expect(vault.entries[0].lastAccessedAt).toBeDefined();
    });

    it('should throw EntryNotFoundError for non-existent entry', async () => {
      const vault = storage.createEmptyVault();
      await expect(storage.getEntry(vault, 'nonexistent', password)).rejects.toThrow(EntryNotFoundError);
    });
  });

  describe('updateEntry', () => {
    it('should update entry name', async () => {
      const vault = storage.createEmptyVault();
      const entryData = {
        type: 'credential' as const,
        name: 'Original Name',
        data: { username: 'test' },
      };

      const entry = await storage.addEntry(vault, entryData, password);
      const updated = await storage.updateEntry(vault, entry.id, { name: 'Updated Name' }, password);

      expect(updated.name).toBe('Updated Name');
      expect(vault.entries[0].name).toBe('Updated Name');
    });

    it('should update entry data', async () => {
      const vault = storage.createEmptyVault();
      const entryData = {
        type: 'credential' as const,
        name: 'Test Entry',
        data: { username: 'original' },
      };

      const entry = await storage.addEntry(vault, entryData, password);
      await storage.updateEntry(vault, entry.id, { data: { username: 'updated' } }, password);

      const retrieved = await storage.getEntry(vault, entry.id, password);
      expect(retrieved.data).toEqual({ username: 'updated' });
    });

    it('should throw EntryNotFoundError for non-existent entry', async () => {
      const vault = storage.createEmptyVault();
      await expect(storage.updateEntry(vault, 'nonexistent', { name: 'Test' }, password)).rejects.toThrow(EntryNotFoundError);
    });
  });

  describe('deleteEntry', () => {
    it('should delete entry from vault', async () => {
      const vault = storage.createEmptyVault();
      const entryData = {
        type: 'credential' as const,
        name: 'Test Entry',
        data: { username: 'test' },
      };

      const entry = await storage.addEntry(vault, entryData, password);
      expect(vault.entries).toHaveLength(1);

      await storage.deleteEntry(vault, entry.id);

      expect(vault.entries).toHaveLength(0);
    });

    it('should throw EntryNotFoundError for non-existent entry', async () => {
      const vault = storage.createEmptyVault();
      await expect(storage.deleteEntry(vault, 'nonexistent')).rejects.toThrow(EntryNotFoundError);
    });
  });

  describe('listEntries', () => {
    it('should return metadata for all entries', () => {
      const vault = storage.createEmptyVault();
      vault.entries = [
        {
          id: 'entry-1',
          type: 'credential',
          name: 'Entry 1',
          encryptedData: 'data',
          iv: 'iv',
          tag: 'tag',
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-01T00:00:00Z',
          accessCount: 5,
          lastAccessedAt: '2024-01-02T00:00:00Z',
        },
      ];

      const entries = storage.listEntries(vault);

      expect(entries).toHaveLength(1);
      expect(entries[0].id).toBe('entry-1');
      expect(entries[0].name).toBe('Entry 1');
      expect(entries[0].type).toBe('credential');
      expect(entries[0].accessCount).toBe(5);
      expect('encryptedData' in entries[0]).toBe(false);
    });
  });

  describe('verifyIntegrity', () => {
    it('should return valid for correct vault', async () => {
      const vault = storage.createEmptyVault();
      const report = await storage.verifyIntegrity(vault);

      expect(report.valid).toBe(true);
      expect(report.errors).toHaveLength(0);
      expect(report.entryCount).toBe(0);
    });

    it('should detect invalid version', async () => {
      const vault = storage.createEmptyVault();
      vault.version = 999;

      const report = await storage.verifyIntegrity(vault);

      expect(report.valid).toBe(false);
      expect(report.errors).toContain('Unsupported vault version: 999. Expected: 1');
    });

    it('should detect missing metadata', async () => {
      const vault = storage.createEmptyVault();
      (vault as Partial<VaultFile>).metadata = undefined;

      const report = await storage.verifyIntegrity(vault);

      expect(report.valid).toBe(false);
      expect(report.errors).toContain('Missing or invalid vault metadata');
    });
  });
});

describe('createVaultStorage', () => {
  it('should create VaultStorage instance', () => {
    const storage = createVaultStorage();
    expect(storage).toBeInstanceOf(VaultStorage);
  });

  it('should use custom backup directory', async () => {
    const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'vault-test-'));
    const backupDir = path.join(tempDir, 'custom-backups');
    
    const storage = createVaultStorage(backupDir);
    const vault = storage.createEmptyVault();
    const vaultPath = path.join(tempDir, 'vault.dat');
    
    await storage.save(vaultPath, vault, 'password');
    await storage.createBackup(vaultPath);
    
    expect(fs.existsSync(backupDir)).toBe(true);
    
    fs.rmSync(tempDir, { recursive: true, force: true });
  });
});
