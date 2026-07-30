import { describe, it, expect } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useDocumentMeta } from '../useDocumentMeta';

describe('useDocumentMeta', () => {
  it('title / description / canonical / robots を head に設定する', () => {
    renderHook(() =>
      useDocumentMeta({
        title: 'テストタイトル',
        description: 'テストの説明文',
        canonical: 'https://example.com/x',
        robots: 'noindex, nofollow',
      }),
    );

    expect(document.title).toBe('テストタイトル');
    expect(
      document.head.querySelector('meta[name="description"]')?.getAttribute('content'),
    ).toBe('テストの説明文');
    expect(
      document.head.querySelector('link[rel="canonical"]')?.getAttribute('href'),
    ).toBe('https://example.com/x');
    expect(
      document.head.querySelector('meta[name="robots"]')?.getAttribute('content'),
    ).toBe('noindex, nofollow');
  });

  it('canonical 省略時は現在の origin + pathname を使う', () => {
    renderHook(() => useDocumentMeta({ title: 'x' }));
    const href = document.head.querySelector('link[rel="canonical"]')?.getAttribute('href');
    expect(href).toBe(`${window.location.origin}${window.location.pathname}`);
  });

  it('robots 未指定時は既存の robots メタを削除する（noindex の残留防止）', () => {
    // 認証画面が noindex を設定した後、公開LPへ SPA 遷移するシナリオ
    renderHook(() => useDocumentMeta({ robots: 'noindex, nofollow' }));
    expect(document.head.querySelector('meta[name="robots"]')).not.toBeNull();

    renderHook(() => useDocumentMeta({ title: '公開ページ' }));
    expect(document.head.querySelector('meta[name="robots"]')).toBeNull();
  });
});
