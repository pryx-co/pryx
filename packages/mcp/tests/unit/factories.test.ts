import { describe, it, expect } from 'vitest';
import { createStorage, MCPStorage } from '../../src/storage.js';
import { createRegistry, MCPRegistry } from '../../src/registry.js';

describe('Factory Functions', () => {
  describe('createStorage', () => {
    it('should create a new MCPStorage instance', () => {
      const storage = createStorage();
      expect(storage).toBeInstanceOf(MCPStorage);
    });
  });

  describe('createRegistry', () => {
    it('should create a new MCPRegistry instance', () => {
      const registry = createRegistry();
      expect(registry).toBeInstanceOf(MCPRegistry);
      expect(registry.size).toBe(0);
    });
  });
});
