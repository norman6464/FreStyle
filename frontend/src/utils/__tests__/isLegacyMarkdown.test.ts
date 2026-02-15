import { describe, it, expect } from 'vitest';
import { isLegacyMarkdown } from '../isLegacyMarkdown';

describe('isLegacyMarkdown', () => {
  it('空文字列はfalseを返す', () => {
    expect(isLegacyMarkdown('')).toBe(false);
  });

  it('nullish値はfalseを返す', () => {
    expect(isLegacyMarkdown(null as unknown as string)).toBe(false);
    expect(isLegacyMarkdown(undefined as unknown as string)).toBe(false);
  });

  it('TipTap JSONはfalseを返す', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'text', text: 'Hello' }] }],
    });
    expect(isLegacyMarkdown(json)).toBe(false);
  });

  it('空のTipTap docはfalseを返す', () => {
    const json = JSON.stringify({ type: 'doc', content: [] });
    expect(isLegacyMarkdown(json)).toBe(false);
  });

  it('プレーンテキストはtrueを返す', () => {
    expect(isLegacyMarkdown('普通のテキスト')).toBe(true);
  });

  it('マークダウン見出しはtrueを返す', () => {
    expect(isLegacyMarkdown('# 見出し')).toBe(true);
  });

  it('マークダウンリストはtrueを返す', () => {
    expect(isLegacyMarkdown('- 項目1\n- 項目2')).toBe(true);
  });

  it('不正なJSONはtrueを返す', () => {
    expect(isLegacyMarkdown('{invalid json}')).toBe(true);
  });

  it('typeがdocでないJSONはtrueを返す', () => {
    const json = JSON.stringify({ type: 'other', content: [] });
    expect(isLegacyMarkdown(json)).toBe(true);
  });
});
