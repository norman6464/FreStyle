import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTableOfContents } from '../useTableOfContents';

describe('useTableOfContents', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('空コンテンツの場合は空の見出しリストを返す', () => {
    const { result } = renderHook(() => useTableOfContents(''));
    expect(result.current.headings).toEqual([]);
  });

  it('見出しがないコンテンツでは空リストを返す', () => {
    const content = JSON.stringify({
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: 'テキスト' }] },
      ],
    });
    const { result } = renderHook(() => useTableOfContents(content));
    expect(result.current.headings).toEqual([]);
  });

  it('H1見出しを抽出する', () => {
    const content = JSON.stringify({
      type: 'doc',
      content: [
        { type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: '大見出し' }] },
      ],
    });
    const { result } = renderHook(() => useTableOfContents(content));
    expect(result.current.headings).toEqual([
      { level: 1, text: '大見出し', id: 'heading-0' },
    ]);
  });

  it('複数レベルの見出しを抽出する', () => {
    const content = JSON.stringify({
      type: 'doc',
      content: [
        { type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: 'タイトル' }] },
        { type: 'paragraph', content: [{ type: 'text', text: 'テキスト' }] },
        { type: 'heading', attrs: { level: 2 }, content: [{ type: 'text', text: 'セクション' }] },
        { type: 'heading', attrs: { level: 3 }, content: [{ type: 'text', text: 'サブセクション' }] },
      ],
    });
    const { result } = renderHook(() => useTableOfContents(content));
    expect(result.current.headings).toEqual([
      { level: 1, text: 'タイトル', id: 'heading-0' },
      { level: 2, text: 'セクション', id: 'heading-1' },
      { level: 3, text: 'サブセクション', id: 'heading-2' },
    ]);
  });

  it('テキストが空の見出しは除外する', () => {
    const content = JSON.stringify({
      type: 'doc',
      content: [
        { type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: '大見出し' }] },
        { type: 'heading', attrs: { level: 2 } },
        { type: 'heading', attrs: { level: 3 }, content: [] },
      ],
    });
    const { result } = renderHook(() => useTableOfContents(content));
    expect(result.current.headings).toEqual([
      { level: 1, text: '大見出し', id: 'heading-0' },
    ]);
  });

  it('レガシーMarkdownの場合は空リストを返す', () => {
    const { result } = renderHook(() => useTableOfContents('普通のテキスト'));
    expect(result.current.headings).toEqual([]);
  });

  it('不正なJSONの場合は空リストを返す', () => {
    const { result } = renderHook(() => useTableOfContents('invalid{{{'));
    expect(result.current.headings).toEqual([]);
  });

  it('isOpenの初期値はfalse', () => {
    const { result } = renderHook(() => useTableOfContents(''));
    expect(result.current.isOpen).toBe(false);
  });

  it('toggleで開閉できる', () => {
    const { result } = renderHook(() => useTableOfContents(''));

    act(() => {
      result.current.toggle();
    });
    expect(result.current.isOpen).toBe(true);

    act(() => {
      result.current.toggle();
    });
    expect(result.current.isOpen).toBe(false);
  });
});
