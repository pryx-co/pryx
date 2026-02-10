/**
 * Vault Storage Types Module
 *
 * Defines types and interfaces for vault storage, entries, and errors.
 */

/**
 * Current vault format version
 */
export const VAULT_FORMAT_VERSION = 1;

/**
 * Maximum number of backups to keep
 */
export const MAX_BACKUPS = 5;

/**
 * Type of vault entry
 */
export type EntryType = 'credential' | 'api-key' | 'token' | 'note';

/**
 * Vault metadata structure
 */
export interface VaultMetadata {
  salt: string;
  algorithm: string;
  iterations: number;
  memoryCost: number;
}

/**
 * Vault entry structure
 */
export interface VaultEntry {
  id: string;
  type: EntryType;
  name: string;
  encryptedData: string;
  iv: string;
  tag: string;
  createdAt: string;
  updatedAt: string;
  accessCount: number;
  lastAccessedAt?: string;
}

/**
 * Vault file structure
 */
export interface VaultFile {
  version: number;
  createdAt: string;
  updatedAt: string;
  metadata: VaultMetadata;
  entries: VaultEntry[];
}

/**
 * Entry data structure
 */
export interface EntryData {
  id?: string;
  type: EntryType;
  name: string;
  data: Record<string, unknown>;
}

/**
 * Entry metadata structure
 */
export interface EntryMetadata {
  id: string;
  type: EntryType;
  name: string;
  createdAt: string;
  updatedAt: string;
  accessCount: number;
  lastAccessedAt?: string;
}

/**
 * Integrity report structure
 */
export interface IntegrityReport {
  valid: boolean;
  errors: string[];
  entryCount: number;
  corruptedEntries: string[];
}

/**
 * Migration structure
 */
export interface Migration {
  fromVersion: number;
  toVersion: number;
  migrate: (data: unknown) => unknown;
}

/**
 * Base storage error class
 */
export class StorageError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'StorageError';
  }
}

/**
 * Error thrown when a file is not found
 */
export class FileNotFoundError extends StorageError {
  constructor(filePath: string) {
    super(`Vault file not found: ${filePath}`);
    this.name = 'FileNotFoundError';
  }
}

/**
 * Error thrown when the vault appears corrupted
 */
export class CorruptedVaultError extends StorageError {
  constructor(message: string = 'Vault file appears to be corrupted') {
    super(message);
    this.name = 'CorruptedVaultError';
  }
}

/**
 * Error thrown when an entry is not found
 */
export class EntryNotFoundError extends StorageError {
  constructor(entryId: string) {
    super(`Entry not found: ${entryId}`);
    this.name = 'EntryNotFoundError';
  }
}

/**
 * Error thrown when an entry already exists
 */
export class DuplicateEntryError extends StorageError {
  constructor(entryId: string) {
    super(`Entry already exists: ${entryId}`);
    this.name = 'DuplicateEntryError';
  }
}

/**
 * Error thrown when migration fails
 */
export class MigrationError extends StorageError {
  constructor(fromVersion: number, toVersion: number) {
    super(`Migration failed from version ${fromVersion} to ${toVersion}`);
    this.name = 'MigrationError';
  }
}
