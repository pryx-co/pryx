import { deriveKey, secureClear } from './crypto.js';
import { VaultConfig, DEFAULT_VAULT_CONFIG } from './types.js';

/**
 * Represents a cached encryption key entry
 */
export interface KeyCacheEntry {
  /** The cached encryption key */
  key: Buffer;
  /** The salt used for key derivation */
  salt: Buffer;
  /** Timestamp when the entry was created */
  createdAt: number;
  /** Timestamp of last access */
  lastAccessedAt: number;
  /** Number of times this key has been accessed */
  accessCount: number;
}

/**
 * Configuration options for the key cache
 */
export interface KeyCacheConfig {
  /** Maximum age of cached keys in milliseconds */
  maxAgeMs: number;
  /** Maximum idle time before eviction in milliseconds */
  maxIdleMs: number;
  /** Maximum number of entries to cache */
  maxEntries: number;
}

/**
 * Default configuration for key cache
 * - Max age: 30 minutes
 * - Max idle: 5 minutes
 * - Max entries: 10
 */
export const DEFAULT_CACHE_CONFIG: KeyCacheConfig = {
  maxAgeMs: 30 * 60 * 1000, // 30 minutes max age
  maxIdleMs: 5 * 60 * 1000,  // 5 minutes idle timeout
  maxEntries: 10,
};

/**
 * LRU cache for derived encryption keys to avoid expensive Argon2id operations.
 *
 * Features:
 * - Automatic key derivation caching
 * - LRU eviction when max entries reached
 * - Time-based expiration (max age and idle time)
 * - Automatic cleanup of expired entries
 * - Secure memory clearing on eviction
 *
 * @example
 * ```typescript
 * const cache = new KeyCache({ maxEntries: 5 });
 * const key = await cache.getKey(password, salt);
 * // Key is cached for subsequent accesses
 * ```
 */
export class KeyCache {
  private _cache: Map<string, KeyCacheEntry> = new Map();
  private _config: KeyCacheConfig;
  private _cleanupInterval: NodeJS.Timeout | null = null;

  /**
   * Creates a new KeyCache instance
   * @param config - Optional cache configuration overrides
   */
  constructor(config: Partial<KeyCacheConfig> = {}) {
    this._config = { ...DEFAULT_CACHE_CONFIG, ...config };
    this._startCleanupInterval();
  }

  /**
   * Gets a cached key or derives a new one if not cached or expired
   * @param password - The password to derive the key from
   * @param salt - The salt for key derivation
   * @param vaultConfig - Optional vault configuration for key derivation
   * @returns The derived or cached encryption key
   */
  async getKey(
    password: string,
    salt: Buffer,
    vaultConfig: VaultConfig = DEFAULT_VAULT_CONFIG
  ): Promise<Buffer> {
    const cacheKey = this._generateCacheKey(password, salt);
    const entry = this._cache.get(cacheKey);

    if (entry && this._isValid(entry)) {
      entry.lastAccessedAt = Date.now();
      entry.accessCount++;
      return Buffer.from(entry.key);
    }

    const key = await deriveKey(password, salt, vaultConfig);
    this._store(cacheKey, key, salt);
    return key;
  }

  /**
   * Invalidates a specific cached key
   * @param password - The password of the key to invalidate
   * @param salt - The salt of the key to invalidate
   */
  invalidate(password: string, salt: Buffer): void {
    const cacheKey = this._generateCacheKey(password, salt);
    const entry = this._cache.get(cacheKey);
    if (entry) {
      secureClear(entry.key);
      this._cache.delete(cacheKey);
    }
  }

  /**
   * Invalidates all cached keys and clears the cache
   */
  invalidateAll(): void {
    for (const entry of this._cache.values()) {
      secureClear(entry.key);
    }
    this._cache.clear();
  }

  /**
   * Gets statistics about the current cache state
   * @returns Cache statistics including size, access counts, and entry timestamps
   */
  getStats(): {
    size: number;
    totalAccessCount: number;
    oldestEntry: number | null;
    newestEntry: number | null;
  } {
    let totalAccessCount = 0;
    let oldestEntry: number | null = null;
    let newestEntry: number | null = null;

    for (const entry of this._cache.values()) {
      totalAccessCount += entry.accessCount;
      if (!oldestEntry || entry.createdAt < oldestEntry) {
        oldestEntry = entry.createdAt;
      }
      if (!newestEntry || entry.createdAt > newestEntry) {
        newestEntry = entry.createdAt;
      }
    }

    return {
      size: this._cache.size,
      totalAccessCount,
      oldestEntry,
      newestEntry,
    };
  }

  /**
   * Destroys the cache, clearing all keys and stopping cleanup interval
   */
  destroy(): void {
    if (this._cleanupInterval) {
      clearInterval(this._cleanupInterval);
      this._cleanupInterval = null;
    }
    this.invalidateAll();
  }

  private _generateCacheKey(password: string, salt: Buffer): string {
    // Use synchronous crypto for cache key generation
    const { createHash } = require('crypto');
    return createHash('sha256').update(password).update(salt).digest('hex');
  }

  private _isValid(entry: KeyCacheEntry): boolean {
    const now = Date.now();
    const age = now - entry.createdAt;
    const idle = now - entry.lastAccessedAt;

    return age < this._config.maxAgeMs && idle < this._config.maxIdleMs;
  }

  private _store(cacheKey: string, key: Buffer, salt: Buffer): void {
    if (this._cache.size >= this._config.maxEntries) {
      this._evictLRU();
    }

    this._cache.set(cacheKey, {
      key: Buffer.from(key),
      salt: Buffer.from(salt),
      createdAt: Date.now(),
      lastAccessedAt: Date.now(),
      accessCount: 1,
    });
  }

  private _evictLRU(): void {
    let oldestKey: string | null = null;
    let oldestAccess = Infinity;

    for (const [key, entry] of this._cache.entries()) {
      if (entry.lastAccessedAt < oldestAccess) {
        oldestAccess = entry.lastAccessedAt;
        oldestKey = key;
      }
    }

    if (oldestKey) {
      const entry = this._cache.get(oldestKey);
      if (entry) {
        secureClear(entry.key);
      }
      this._cache.delete(oldestKey);
    }
  }

  private _startCleanupInterval(): void {
    this._cleanupInterval = setInterval(() => {
      this._cleanup();
    }, 60000); // Run cleanup every minute
  }

  private _cleanup(): void {
    const now = Date.now();
    for (const [key, entry] of this._cache.entries()) {
      const age = now - entry.createdAt;
      const idle = now - entry.lastAccessedAt;

      if (age >= this._config.maxAgeMs || idle >= this._config.maxIdleMs) {
        secureClear(entry.key);
        this._cache.delete(key);
      }
    }
  }
}

/**
 * Creates a new KeyCache instance with the specified configuration.
 *
 * This factory function provides a convenient way to instantiate a KeyCache
 * without directly using the `new` keyword. It accepts a partial configuration
 * object to customize caching behavior, such as maximum entry age, idle timeout,
 * and cache size limits.
 *
 * @param config - Optional configuration object to override default cache settings.
 *   - `maxAgeMs`: Maximum age of cached keys in milliseconds (default: 30 minutes)
 *   - `maxIdleMs`: Maximum idle time before eviction in milliseconds (default: 5 minutes)
 *   - `maxEntries`: Maximum number of entries to cache (default: 10)
 * @returns A new {@link KeyCache} instance configured with the provided options.
 *
 * @example
 * ```typescript
 * const cache = createKeyCache({ maxEntries: 20, maxAgeMs: 60000 });
 * const key = await cache.getKey(password, salt);
 * ```
 */
export function createKeyCache(config?: Partial<KeyCacheConfig>): KeyCache {
  return new KeyCache(config);
}
