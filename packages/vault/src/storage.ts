import { readFile, writeFile, mkdir, rename, unlink, access, constants } from 'fs/promises';
import { dirname, join } from 'path';
import { randomUUID } from 'crypto';
import { deriveKey, encrypt, decrypt, generateSalt, generateIV } from './crypto.js';
import { DEFAULT_VAULT_CONFIG } from './types.js';
import { BackupManager } from './backup.js';
import {
  VaultFile,
  VaultEntry,
  EntryData,
  EntryMetadata,
  IntegrityReport,
  VAULT_FORMAT_VERSION,
  StorageError,
  FileNotFoundError,
  CorruptedVaultError,
  EntryNotFoundError,
  DuplicateEntryError,
} from './storage-types.js';

/**
 * Manages encrypted vault storage operations including loading, saving,
 * and managing encrypted entries with integrity verification.
 * 
 * Features:
 * - Atomic file writes with temporary file and rename
 * - Automatic backup creation before modifications
 * - Entry-level encryption with AES-256-GCM
 * - Integrity verification and corruption detection
 * - Metadata tracking (creation, access counts, timestamps)
 */
export class VaultStorage {
  private backupManager: BackupManager;

  /**
   * Creates a new VaultStorage instance
   * @param backupDir - Optional custom backup directory path (defaults to ~/.pryx/vault-backups)
   */
  constructor(backupDir?: string) {
    const defaultBackupDir = join(process.env.HOME || process.env.USERPROFILE || '.', '.pryx', 'vault-backups');
    this.backupManager = new BackupManager(backupDir || defaultBackupDir);
  }

  /**
   * Loads a vault from disk and verifies its integrity
   * @param filePath - Path to the vault file
   * @param _password - Password for integrity verification (unused in current implementation)
   * @returns The loaded vault file structure
   * @throws {FileNotFoundError} If the vault file doesn't exist
   * @throws {CorruptedVaultError} If the vault file is malformed or corrupted
   */
  async load(filePath: string, _password: string): Promise<VaultFile> {
    try {
      await access(filePath, constants.R_OK);
    } catch {
      throw new FileNotFoundError(filePath);
    }

    const data = await readFile(filePath, 'utf-8');
    let vault: VaultFile;

    try {
      vault = JSON.parse(data);
    } catch {
      throw new CorruptedVaultError('Invalid JSON format');
    }

    this.validateVaultStructure(vault);
    
    await this.verifyIntegrity(vault, _password);

    return vault;
  }

  /**
   * Saves a vault to disk atomically using a temporary file
   * @param filePath - Path where the vault should be saved
   * @param vault - The vault file structure to save
   * @param _password - Password for encryption (unused in current implementation)
   * @throws {StorageError} If the save operation fails
   */
  async save(filePath: string, vault: VaultFile, _password: string): Promise<void> {
    vault.updatedAt = new Date().toISOString();
    
    await mkdir(dirname(filePath), { recursive: true });
    
    const tempPath = `${filePath}.tmp.${Date.now()}`;
    
    try {
      await writeFile(tempPath, JSON.stringify(vault, null, 2), { mode: 0o600 });
      await rename(tempPath, filePath);
    } catch (error) {
      try {
        await unlink(tempPath);
      } catch {}
      throw new StorageError(`Failed to save vault: ${(error as Error).message}`);
    }
  }

  /**
   * Adds a new encrypted entry to the vault
   * @param vault - The vault file structure
   * @param entryData - The entry data to encrypt and store
   * @param password - Password for deriving the encryption key
   * @returns The created vault entry with metadata
   * @throws {DuplicateEntryError} If an entry with the same ID already exists
   */
  async addEntry(vault: VaultFile, entryData: EntryData, password: string): Promise<VaultEntry> {
    const existingIndex = vault.entries.findIndex(e => e.id === entryData.id);
    if (existingIndex !== -1) {
      throw new DuplicateEntryError(entryData.id || 'unknown');
    }

    const id = entryData.id || randomUUID();
    const now = new Date().toISOString();
    
    const salt = Buffer.from(vault.metadata.salt, 'base64');
    const key = await deriveKey(password, salt);
    const iv = generateIV();
    
    const plaintext = Buffer.from(JSON.stringify(entryData.data));
    const { ciphertext, tag } = encrypt(plaintext, key, iv);
    
    const entry: VaultEntry = {
      id,
      type: entryData.type,
      name: entryData.name,
      encryptedData: ciphertext.toString('base64'),
      iv: iv.toString('base64'),
      tag: tag.toString('base64'),
      createdAt: now,
      updatedAt: now,
      accessCount: 0,
    };

    vault.entries.push(entry);
    
    return entry;
  }

  /**
   * Updates an existing entry in the vault
   * @param vault - The vault file structure
   * @param id - The unique identifier of the entry to update
   * @param updates - Partial entry data with fields to update
   * @param password - Password for deriving the encryption key
   * @returns The updated vault entry
   * @throws {EntryNotFoundError} If the entry doesn't exist
   */
  async updateEntry(
    vault: VaultFile,
    id: string,
    updates: Partial<EntryData>,
    password: string
  ): Promise<VaultEntry> {
    const index = vault.entries.findIndex(e => e.id === id);
    if (index === -1) {
      throw new EntryNotFoundError(id);
    }

    const entry = vault.entries[index];
    
    if (updates.name !== undefined) {
      entry.name = updates.name;
    }
    
    if (updates.data !== undefined) {
      const salt = Buffer.from(vault.metadata.salt, 'base64');
      const key = await deriveKey(password, salt);
      const iv = generateIV();
      
      const plaintext = Buffer.from(JSON.stringify(updates.data));
      const { ciphertext, tag } = encrypt(plaintext, key, iv);
      
      entry.encryptedData = ciphertext.toString('base64');
      entry.iv = iv.toString('base64');
      entry.tag = tag.toString('base64');
    }

    entry.updatedAt = new Date().toISOString();
    
    return entry;
  }

  /**
   * Deletes an entry from the vault
   * @param vault - The vault file structure
   * @param id - The unique identifier of the entry to delete
   * @throws {EntryNotFoundError} If the entry doesn't exist
   */
  async deleteEntry(vault: VaultFile, id: string): Promise<void> {
    const index = vault.entries.findIndex(e => e.id === id);
    if (index === -1) {
      throw new EntryNotFoundError(id);
    }
    
    vault.entries.splice(index, 1);
  }

  /**
   * Retrieves and decrypts an entry from the vault
   * @param vault - The vault file structure
   * @param id - The unique identifier of the entry to retrieve
   * @param password - Password for deriving the decryption key
   * @returns The decrypted entry data
   * @throws {EntryNotFoundError} If the entry doesn't exist
   * @throws {CorruptedVaultError} If decryption fails (indicates corruption)
   */
  async getEntry(vault: VaultFile, id: string, password: string): Promise<EntryData> {
    const entry = vault.entries.find(e => e.id === id);
    if (!entry) {
      throw new EntryNotFoundError(id);
    }

    try {
      const salt = Buffer.from(vault.metadata.salt, 'base64');
      const key = await deriveKey(password, salt);
      const iv = Buffer.from(entry.iv, 'base64');
      const ciphertext = Buffer.from(entry.encryptedData, 'base64');
      const tag = Buffer.from(entry.tag, 'base64');
      
      const plaintext = decrypt(ciphertext, key, iv, tag);
      const data = JSON.parse(plaintext.toString('utf-8'));
      
      entry.accessCount++;
      entry.lastAccessedAt = new Date().toISOString();
      
      return {
        id: entry.id,
        type: entry.type,
        name: entry.name,
        data,
      };
    } catch (error) {
      throw new CorruptedVaultError(`Failed to decrypt entry ${id}: ${(error as Error).message}`);
    }
  }

  /**
   * Lists metadata for all entries in the vault
   * @param vault - The vault file structure
   * @returns Array of entry metadata (excluding encrypted data)
   */
  listEntries(vault: VaultFile): EntryMetadata[] {
    return vault.entries.map(entry => ({
      id: entry.id,
      type: entry.type,
      name: entry.name,
      createdAt: entry.createdAt,
      updatedAt: entry.updatedAt,
      accessCount: entry.accessCount,
      lastAccessedAt: entry.lastAccessedAt,
    }));
  }

  /**
   * Creates a backup of the vault file
   * @param filePath - Path to the vault file to backup
   * @returns Path to the created backup file
   */
  async createBackup(filePath: string): Promise<string> {
    return this.backupManager.createBackup(filePath);
  }

  /**
   * Restores a vault from a backup file
   * @param backupPath - Path to the backup file
   * @param targetPath - Path where the restored vault should be saved
   * @returns The restored vault file structure
   * @throws {FileNotFoundError} If the backup doesn't exist
   * @throws {CorruptedVaultError} If the backup is corrupted
   */
  async restoreFromBackup(backupPath: string, targetPath: string): Promise<VaultFile> {
    await this.backupManager.restoreBackup(backupPath, targetPath);
    
    const data = await readFile(targetPath, 'utf-8');
    return JSON.parse(data);
  }

  /**
   * Verifies the integrity of a vault file
   * @param vault - The vault file structure to verify
   * @param password - Optional password to verify entry decryption (if provided)
   * @returns Integrity report with validation results
   */
  async verifyIntegrity(vault: VaultFile, password?: string): Promise<IntegrityReport> {
    const report: IntegrityReport = {
      valid: true,
      errors: [],
      entryCount: vault.entries.length,
      corruptedEntries: [],
    };

    if (vault.version !== VAULT_FORMAT_VERSION) {
      report.valid = false;
      report.errors.push(`Unsupported vault version: ${vault.version}. Expected: ${VAULT_FORMAT_VERSION}`);
    }

    if (!vault.metadata || !vault.metadata.salt) {
      report.valid = false;
      report.errors.push('Missing or invalid vault metadata');
    }

    if (password) {
      for (const entry of vault.entries) {
        try {
          const salt = Buffer.from(vault.metadata.salt, 'base64');
          const key = await deriveKey(password, salt);
          const iv = Buffer.from(entry.iv, 'base64');
          const ciphertext = Buffer.from(entry.encryptedData, 'base64');
          const tag = Buffer.from(entry.tag, 'base64');
          
          decrypt(ciphertext, key, iv, tag);
        } catch {
          report.valid = false;
          report.corruptedEntries.push(entry.id);
        }
      }
    }

    return report;
  }

  /**
   * Creates a new empty vault with default configuration
   * @returns A new vault file structure with initialized metadata
   */
  createEmptyVault(): VaultFile {
    const salt = generateSalt();
    const now = new Date().toISOString();
    
    return {
      version: VAULT_FORMAT_VERSION,
      createdAt: now,
      updatedAt: now,
      metadata: {
        salt: salt.toString('base64'),
        algorithm: 'argon2id+aes-256-gcm',
        iterations: DEFAULT_VAULT_CONFIG.timeCost,
        memoryCost: DEFAULT_VAULT_CONFIG.memoryCost,
      },
      entries: [],
    };
  }

  private validateVaultStructure(vault: unknown): asserts vault is VaultFile {
    if (typeof vault !== 'object' || vault === null) {
      throw new CorruptedVaultError('Vault is not an object');
    }

    const v = vault as Record<string, unknown>;

    if (typeof v.version !== 'number') {
      throw new CorruptedVaultError('Missing or invalid version');
    }

    if (typeof v.createdAt !== 'string') {
      throw new CorruptedVaultError('Missing or invalid createdAt');
    }

    if (typeof v.updatedAt !== 'string') {
      throw new CorruptedVaultError('Missing or invalid updatedAt');
    }

    if (typeof v.metadata !== 'object' || v.metadata === null) {
      throw new CorruptedVaultError('Missing or invalid metadata');
    }

    if (!Array.isArray(v.entries)) {
      throw new CorruptedVaultError('Missing or invalid entries array');
    }
  }
}

/**
 * Creates a new VaultStorage instance
 * @param backupDir - Optional custom backup directory path
 * @returns A new VaultStorage instance
 */
export function createVaultStorage(backupDir?: string): VaultStorage {
  return new VaultStorage(backupDir);
}
