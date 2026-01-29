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
