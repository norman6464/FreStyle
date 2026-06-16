import { describe, it, expect } from 'vitest';
import { monacoLanguageOf, pad } from '../exerciseFormat';

describe('monacoLanguageOf', () => {
  it.each([
    ['php', 'php'],
    ['go', 'go'],
    ['sql', 'sql'],
    ['python', 'python'],
    ['javascript', 'javascript'],
    ['typescript', 'typescript'],
    ['bash', 'shell'],
    ['sh', 'shell'],
  ])('%s → Monaco %s', (lang, expected) => {
    expect(monacoLanguageOf(lang)).toBe(expected);
  });

  it('未対応の言語は plaintext に fallback する', () => {
    expect(monacoLanguageOf('docker')).toBe('plaintext');
    expect(monacoLanguageOf('')).toBe('plaintext');
  });
});

describe('pad', () => {
  it('2 桁 0 埋め', () => {
    expect(pad(3)).toBe('03');
    expect(pad(12)).toBe('12');
    expect(pad(123)).toBe('123');
  });
});
