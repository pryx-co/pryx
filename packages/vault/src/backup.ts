/**
 * Vault Backup Manager Module
 *
 * Manages backup creation, restoration, and cleanup for vault files.
 */

import { copyFile, mkdir, readdir, stat, unlink } from 'fs/promises';
import { dirname, basename, join } from 'path';
import { MAX_BACKUPS } from './storage-types.js';

/**
 * Information about a backup file
 */
export interface BackupInfo {
  path: string;
  createdAt: Date;
  size: number;
}

/**
 * Manages backups for vault files
 */
export class BackupManager {
  private backupDir: string;

  /**
   * Creates a new BackupManager instance
   * @param backupDir - Directory to store backups
   */
  constructor(backupDir: string) {
    this.backupDir = backupDir;
  }

  /**
   * Creates a backup of a file
   * @param filePath - Path to the file to backup
   * @returns Path to the created backup
   */
  async createBackup(filePath: string): Promise<string> {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const originalName = basename(filePath);
    const backupPath = join(this.backupDir, `${originalName}.${timestamp}.backup`);

    await mkdir(this.backupDir, { recursive: true });
    await copyFile(filePath, backupPath);

    await this.cleanupOldBackups(originalName);

    return backupPath;
  }

  /**
   * Lists all backups for a file
   * @param originalName - Original file name
   * @returns Array of backup information sorted by date (newest first)
   */
  async listBackups(originalName: string): Promise<BackupInfo[]> {
    try {
      const files = await readdir(this.backupDir);
      const backups: BackupInfo[] = [];

      for (const file of files) {
        if (file.startsWith(originalName) && file.endsWith('.backup')) {
          const path = join(this.backupDir, file);
          const stats = await stat(path);
          backups.push({
            path,
            createdAt: stats.mtime,
            size: stats.size,
          });
        }
      }

      return backups.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
    } catch {
      return [];
    }
  }

  /**
   * Restores a backup to a target path
   * @param backupPath - Path to the backup file
   * @param targetPath - Path to restore to
   */
  async restoreBackup(backupPath: string, targetPath: string): Promise<void> {
    await mkdir(dirname(targetPath), { recursive: true });
    await copyFile(backupPath, targetPath);
  }

  /**
   * Cleans up old backups keeping only the most recent ones
   * @param originalName - Original file name
   */
  async cleanupOldBackups(originalName: string): Promise<void> {
    const backups = await this.listBackups(originalName);

    if (backups.length > MAX_BACKUPS) {
      const toDelete = backups.slice(MAX_BACKUPS);
      for (const backup of toDelete) {
        try {
          await unlink(backup.path);
        } catch {
          // Ignore errors when deleting old backups
        }
      }
    }
  }

  /**
   * Gets the most recent backup for a file
   * @param originalName - Original file name
   * @returns Latest backup info or null if no backups exist
   */
  async getLatestBackup(originalName: string): Promise<BackupInfo | null> {
    const backups = await this.listBackups(originalName);
    return backups[0] || null;
  }
}
