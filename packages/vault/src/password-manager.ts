import { Vault, createVault } from './vault.js';
import { KeyCache, createKeyCache } from './key-cache.js';
import { VaultConfig, VaultError } from './types.js';

/**
 * Configuration options for the PasswordManager
 */
export interface PasswordManagerConfig {
  /** Optional vault configuration overrides */
  vaultConfig?: Partial<VaultConfig>;
  /** Optional key cache configuration */
  cacheConfig?: Parameters<typeof createKeyCache>[0];
  /** Auto-lock timeout in milliseconds (default: 5 minutes) */
  autoLockMs?: number;
}

/**
 * Default configuration for password manager
 * - Auto-lock: 5 minutes of inactivity
 */
export const DEFAULT_PASSWORD_MANAGER_CONFIG = {
  autoLockMs: 5 * 60 * 1000,
};

/**
 * High-level password manager that provides a secure interface for encryption
 * and decryption operations with automatic key caching and auto-lock functionality.
 *
 * Features:
 * - Secure vault management with password-based unlocking
 * - Automatic key caching for performance
 * - Auto-lock after period of inactivity
 * - Password change capability
 *
 * @example
 * ```typescript
 * const pm = new PasswordManager({ autoLockMs: 60000 });
 * await pm.unlock('my-password');
 * const encrypted = await pm.encrypt(Buffer.from('secret data'));
 * pm.lock();
 * ```
 */
export class PasswordManager {
  private _vault: Vault | null = null;
  private _keyCache: KeyCache;
  private _config: PasswordManagerConfig;
  private _autoLockTimer: NodeJS.Timeout | null = null;
  private _isLocked = true;

  /**
   * Creates a new PasswordManager instance
   * @param config - Optional configuration for vault, cache, and auto-lock settings
   */
  constructor(config: PasswordManagerConfig = {}) {
    this._config = {
      ...DEFAULT_PASSWORD_MANAGER_CONFIG,
      ...config,
    };
    this._keyCache = createKeyCache(config.cacheConfig);
  }

  /**
   * Unlocks the password manager with the provided password
   * @param password - The password to unlock the vault
   * @throws {VaultError} If the password is invalid
   */
  async unlock(password: string): Promise<void> {
    if (!this._isLocked) {
      return;
    }

    try {
      this._vault = await createVault(password, this._config.vaultConfig);
      this._isLocked = false;
      this._resetAutoLockTimer();
    } catch (error) {
      throw new VaultError('Failed to unlock vault: invalid password');
    }
  }

  /**
   * Locks the password manager, clearing keys and cache
   */
  lock(): void {
    if (this._vault) {
      this._vault.clearKey();
      this._vault = null;
    }
    this._keyCache.invalidateAll();
    this._isLocked = true;
    this._clearAutoLockTimer();
  }

  /**
   * Encrypts plaintext data using the unlocked vault
   * @param plaintext - The data to encrypt
   * @returns The encrypted data
   * @throws {VaultError} If the manager is locked
   */
  async encrypt(plaintext: Buffer): Promise<import('./types.js').EncryptedData> {
    this._ensureUnlocked();
    this._resetAutoLockTimer();
    return this._vault!.encrypt(plaintext);
  }

  /**
   * Decrypts encrypted data using the unlocked vault
   * @param encrypted - The encrypted data to decrypt
   * @returns The decrypted plaintext
   * @throws {VaultError} If the manager is locked
   */
  async decrypt(encrypted: import('./types.js').EncryptedData): Promise<Buffer> {
    this._ensureUnlocked();
    this._resetAutoLockTimer();
    return this._vault!.decrypt(encrypted);
  }

  /**
   * Changes the vault password (must be unlocked first)
   * @param _oldPassword - The current password (for verification in future implementations)
   * @param newPassword - The new password to set
   * @throws {VaultError} If the manager is locked or password change fails
   */
  async changePassword(_oldPassword: string, newPassword: string): Promise<void> {
    this._ensureUnlocked();

    try {
      const newVault = await createVault(newPassword, this._config.vaultConfig);

      this._vault = newVault;
      this._keyCache.invalidateAll();
      this._resetAutoLockTimer();
    } catch (error) {
      throw new VaultError('Failed to change password');
    }
  }

  /**
   * Checks if the password manager is unlocked
   * @returns True if unlocked and ready for operations
   */
  isUnlocked(): boolean {
    return !this._isLocked && this._vault !== null;
  }

  /**
   * Checks if the password manager is locked
   * @returns True if locked
   */
  isLocked(): boolean {
    return this._isLocked;
  }

  /**
   * Gets the remaining time until auto-lock
   * @returns Time in milliseconds until auto-lock, or null if already locked/no timer
   */
  getRemainingLockTime(): number | null {
    if (this._isLocked || !this._autoLockTimer) {
      return null;
    }
    return this._config.autoLockMs ?? DEFAULT_PASSWORD_MANAGER_CONFIG.autoLockMs;
  }

  /**
   * Destroys the password manager, locking it and cleaning up resources
   */
  destroy(): void {
    this.lock();
    this._keyCache.destroy();
  }

  private _ensureUnlocked(): void {
    if (this._isLocked || !this._vault) {
      throw new VaultError('Vault is locked. Call unlock() first.');
    }
  }

  private _resetAutoLockTimer(): void {
    this._clearAutoLockTimer();
    const timeout = this._config.autoLockMs ?? DEFAULT_PASSWORD_MANAGER_CONFIG.autoLockMs;
    if (timeout > 0) {
      this._autoLockTimer = setTimeout(() => {
        this.lock();
      }, timeout);
    }
  }

  private _clearAutoLockTimer(): void {
    if (this._autoLockTimer) {
      clearTimeout(this._autoLockTimer);
      this._autoLockTimer = null;
    }
  }
}

/**
 * Creates a new PasswordManager instance
 * @param config - Optional configuration for the password manager
 * @returns A new PasswordManager instance
 */
export function createPasswordManager(config?: PasswordManagerConfig): PasswordManager {
  return new PasswordManager(config);
}
