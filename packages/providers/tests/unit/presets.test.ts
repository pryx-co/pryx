import { describe, it, expect } from 'vitest';
import {
  OPENAI_MODELS,
  ANTHROPIC_MODELS,
  GOOGLE_MODELS,
  LOCAL_MODELS,
  OPENAI_PRESET,
  ANTHROPIC_PRESET,
  GOOGLE_PRESET,
  OLLAMA_PRESET,
  LMSTUDIO_PRESET,
  BUILTIN_PRESETS,
  getPreset,
  getAllPresets,
  getPresetIds,
} from '../../src/presets.js';

describe('OPENAI_MODELS', () => {
  it('should contain GPT-4 models', () => {
    expect(OPENAI_MODELS.some((m) => m.id === 'gpt-4o')).toBe(true);
    expect(OPENAI_MODELS.some((m) => m.id === 'gpt-4o-mini')).toBe(true);
  });

  it('should have correct properties', () => {
    const gpt4o = OPENAI_MODELS.find((m) => m.id === 'gpt-4o');
    expect(gpt4o?.maxTokens).toBe(128000);
    expect(gpt4o?.supportsStreaming).toBe(true);
    expect(gpt4o?.supportsVision).toBe(true);
    expect(gpt4o?.supportsTools).toBe(true);
  });
});

describe('ANTHROPIC_MODELS', () => {
  it('should contain Claude 3 models', () => {
    expect(ANTHROPIC_MODELS.some((m) => m.id === 'claude-3-opus-20240229')).toBe(true);
    expect(ANTHROPIC_MODELS.some((m) => m.id === 'claude-3-sonnet-20240229')).toBe(true);
  });
});

describe('GOOGLE_MODELS', () => {
  it('should contain Gemini models', () => {
    expect(GOOGLE_MODELS.some((m) => m.id === 'gemini-1.5-pro')).toBe(true);
  });
});

describe('OPENAI_PRESET', () => {
  it('should have correct configuration', () => {
    expect(OPENAI_PRESET.id).toBe('openai');
    expect(OPENAI_PRESET.type).toBe('openai');
    expect(OPENAI_PRESET.enabled).toBe(true);
    expect(OPENAI_PRESET.defaultModel).toBe('gpt-4o');
  });
});

describe('ANTHROPIC_PRESET', () => {
  it('should have correct configuration', () => {
    expect(ANTHROPIC_PRESET.id).toBe('anthropic');
    expect(ANTHROPIC_PRESET.type).toBe('anthropic');
    expect(ANTHROPIC_PRESET.defaultModel).toBe('claude-3-sonnet-20240229');
  });
});

describe('OLLAMA_PRESET', () => {
  it('should have correct configuration', () => {
    expect(OLLAMA_PRESET.id).toBe('ollama');
    expect(OLLAMA_PRESET.type).toBe('local');
    expect(OLLAMA_PRESET.enabled).toBe(false);
    expect(OLLAMA_PRESET.baseUrl).toBe('http://localhost:11434');
  });
});

describe('BUILTIN_PRESETS', () => {
  it('should contain all presets', () => {
    expect(BUILTIN_PRESETS.length).toBeGreaterThanOrEqual(5);
    expect(BUILTIN_PRESETS.some((p) => p.id === 'openai')).toBe(true);
    expect(BUILTIN_PRESETS.some((p) => p.id === 'anthropic')).toBe(true);
  });
});

describe('getPreset', () => {
  it('should return preset by id', () => {
    const preset = getPreset('openai');
    expect(preset).toBeDefined();
    expect(preset?.id).toBe('openai');
  });

  it('should return undefined for unknown preset', () => {
    const preset = getPreset('unknown');
    expect(preset).toBeUndefined();
  });
});

describe('getAllPresets', () => {
  it('should return all presets', () => {
    const presets = getAllPresets();
    expect(presets.length).toBe(BUILTIN_PRESETS.length);
  });
});

describe('getPresetIds', () => {
  it('should return all preset ids', () => {
    const ids = getPresetIds();
    expect(ids).toContain('openai');
    expect(ids).toContain('anthropic');
  });
});
