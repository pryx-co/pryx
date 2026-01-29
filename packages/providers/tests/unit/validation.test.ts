import { describe, it, expect } from 'vitest';
import {
  validateProviderConfig,
  assertValidProviderConfig,
  isValidProviderId,
  isValidProviderType,
  isValidUrl,
} from '../../src/validation.js';
import { ProviderValidationError } from '../../src/types.js';

describe('validateProviderConfig', () => {
  const validConfig = {
    id: 'openai',
    name: 'OpenAI',
    type: 'openai' as const,
    enabled: true,
    apiKey: 'sk-test',
    models: [{
      id: 'gpt-4',
      name: 'GPT-4',
      maxTokens: 8192,
      supportsStreaming: true,
      supportsVision: true,
      supportsTools: true,
    }],
  };

  it('should validate correct config', () => {
    const result = validateProviderConfig(validConfig);
    
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('should reject config with invalid id format', () => {
    const config = { ...validConfig, id: 'Invalid ID!' };
    const result = validateProviderConfig(config);
    
    expect(result.valid).toBe(false);
    expect(result.errors.length).toBeGreaterThan(0);
  });

  it('should reject config with empty name', () => {
    const config = { ...validConfig, name: '' };
    const result = validateProviderConfig(config);
    
    expect(result.valid).toBe(false);
  });

  it('should reject config with invalid type', () => {
    const config = { ...validConfig, type: 'invalid' };
    const result = validateProviderConfig(config);
    
    expect(result.valid).toBe(false);
  });

  it('should reject config with empty models array', () => {
    const config = { ...validConfig, models: [] };
    const result = validateProviderConfig(config);
    
    expect(result.valid).toBe(false);
  });

  it('should reject custom provider without baseUrl', () => {
    const config = {
      ...validConfig,
      type: 'custom' as const,
      baseUrl: undefined,
    };
    const result = validateProviderConfig(config);
    
    expect(result.valid).toBe(false);
    expect(result.errors.some((e) => e.includes('baseUrl'))).toBe(true);
  });

  it('should reject local provider with apiKey', () => {
    const config = {
      ...validConfig,
      type: 'local' as const,
      apiKey: 'should-not-have',
    };
    const result = validateProviderConfig(config);
    
    expect(result.valid).toBe(false);
    expect(result.errors.some((e) => e.includes('API key'))).toBe(true);
  });



  it('should reject config with invalid defaultModel', () => {
    const config = { ...validConfig, defaultModel: 'nonexistent-model' };
    const result = validateProviderConfig(config);
    
    expect(result.valid).toBe(false);
    expect(result.errors.some((e) => e.includes('Default model'))).toBe(true);
  });

  it('should accept config with valid defaultModel', () => {
    const config = { ...validConfig, defaultModel: 'gpt-4' };
    const result = validateProviderConfig(config);
    
    expect(result.valid).toBe(true);
  });

  it('should reject config with invalid baseUrl', () => {
    const config = { ...validConfig, baseUrl: 'not-a-url' };
    const result = validateProviderConfig(config);
    
    expect(result.valid).toBe(false);
  });

  it('should accept config with valid baseUrl', () => {
    const config = { ...validConfig, baseUrl: 'https://api.example.com' };
    const result = validateProviderConfig(config);
    
    expect(result.valid).toBe(true);
  });
});

describe('assertValidProviderConfig', () => {
  const validConfig = {
    id: 'test',
    name: 'Test',
    type: 'openai' as const,
    enabled: true,
    apiKey: 'key',
    models: [{
      id: 'model',
      name: 'Model',
      maxTokens: 1000,
      supportsStreaming: true,
      supportsVision: false,
      supportsTools: false,
    }],
  };

  it('should return config when valid', () => {
    const result = assertValidProviderConfig(validConfig);
    
    expect(result.id).toBe('test');
  });

  it('should throw when invalid', () => {
    const config = { ...validConfig, id: '' };
    
    expect(() => assertValidProviderConfig(config)).toThrow(ProviderValidationError);
  });
});

describe('isValidProviderId', () => {
  it('should return true for valid ids', () => {
    expect(isValidProviderId('openai')).toBe(true);
    expect(isValidProviderId('anthropic-v2')).toBe(true);
    expect(isValidProviderId('custom_provider')).toBe(true);
    expect(isValidProviderId('a')).toBe(true);
  });

  it('should return false for invalid ids', () => {
    expect(isValidProviderId('')).toBe(false);
    expect(isValidProviderId('Invalid ID')).toBe(false);
    expect(isValidProviderId('test@provider')).toBe(false);
    expect(isValidProviderId('a'.repeat(65))).toBe(false);
  });
});

describe('isValidProviderType', () => {
  it('should return true for valid types', () => {
    expect(isValidProviderType('openai')).toBe(true);
    expect(isValidProviderType('anthropic')).toBe(true);
    expect(isValidProviderType('google')).toBe(true);
    expect(isValidProviderType('local')).toBe(true);
    expect(isValidProviderType('custom')).toBe(true);
  });

  it('should return false for invalid types', () => {
    expect(isValidProviderType('invalid')).toBe(false);
    expect(isValidProviderType('azure')).toBe(false);
    expect(isValidProviderType('')).toBe(false);
  });
});

describe('isValidUrl', () => {
  it('should return true for valid URLs', () => {
    expect(isValidUrl('https://api.openai.com')).toBe(true);
    expect(isValidUrl('http://localhost:3000')).toBe(true);
    expect(isValidUrl('https://example.com/path')).toBe(true);
  });

  it('should return false for invalid URLs', () => {
    expect(isValidUrl('not-a-url')).toBe(false);
    expect(isValidUrl('')).toBe(false);
  });
});
