/**
 * Vault Package
 *
 * Secure encryption and storage for sensitive data with key derivation,
 * encryption/decryption, and vault management.
 */

export {
  VaultConfig,
  EncryptedData,
  SerializedEncryptedData,
  DEFAULT_VAULT_CONFIG,
  AES_GCM_IV_LENGTH,
  AES_GCM_TAG_LENGTH,
  CURRENT_VERSION,
  VaultError,
  InvalidPasswordError,
  CorruptedDataError,
  DecryptionError,
} from './types.js';

export {
  deriveKey,
  generateSalt,
  generateIV,
  encrypt,
  decrypt,
  secureClear,
  secureCompare,
  serializeEncryptedData,
  deserializeEncryptedData,
} from './crypto.js';

export {
  Vault,
  createVault,
  encryptWithPassword,
  decryptWithPassword,
} from './vault.js';

export {
  KeyCache,
  KeyCacheEntry,
  KeyCacheConfig,
  DEFAULT_CACHE_CONFIG,
  createKeyCache,
} from './key-cache.js';

export {
  PasswordManager,
  PasswordManagerConfig,
  DEFAULT_PASSWORD_MANAGER_CONFIG,
  createPasswordManager,
} from './password-manager.js';

export {
  VaultStorage,
  createVaultStorage,
} from './storage.js';

export {
  BackupManager,
  BackupInfo,
} from './backup.js';

export {
  VaultFile,
  VaultEntry,
  VaultMetadata,
  EntryData,
  EntryMetadata,
  EntryType,
  IntegrityReport,
  Migration,
  VAULT_FORMAT_VERSION,
  MAX_BACKUPS,
  StorageError,
  FileNotFoundError,
  CorruptedVaultError,
  EntryNotFoundError,
  DuplicateEntryError,
  MigrationError,
} from './storage-types.js';