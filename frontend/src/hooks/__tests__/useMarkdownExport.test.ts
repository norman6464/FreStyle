import { renderHook } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useMarkdownExport } from '../useMarkdownExport';

describe('useMarkdownExport', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('copyAsMarkdownがタイトル付きMarkdownを返す', () => {
    const { result } = renderHook(() => useMarkdownExport());
    const doc = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: '本文テスト' }] },
      ],
    };
    const md = result.current.copyAsMarkdown('テストノート', JSON.stringify(doc));
    expect(md).toBe('# テストノート\n\n本文テスト');
  });

  it('copyAsMarkdownがタイトルなしの場合は本文のみ返す', () => {
    const { result } = renderHook(() => useMarkdownExport());
    const doc = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: '本文のみ' }] },
      ],
    };
    const md = result.current.copyAsMarkdown('', JSON.stringify(doc));
    expect(md).toBe('本文のみ');
  });

  it('exportAsMarkdownがBlobを作成しダウンロードする', () => {
    const { result } = renderHook(() => useMarkdownExport());

    const mockClick = vi.fn();
    const originalCreateElement = document.createElement.bind(document);
    const createElementSpy = vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
      if (tag === 'a') {
        return { click: mockClick, href: '', download: '' } as unknown as HTMLAnchorElement;
      }
      return originalCreateElement(tag);
    });
    const appendSpy = vi.spyOn(document.body, 'appendChild').mockImplementation((node) => node);
    const removeSpy = vi.spyOn(document.body, 'removeChild').mockImplementation((node) => node);

    const createURL = vi.fn(() => 'blob:test');
    const revokeURL = vi.fn();
    globalThis.URL.createObjectURL = createURL;
    globalThis.URL.revokeObjectURL = revokeURL;

    const doc = { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: 'テスト' }] }] };
    result.current.exportAsMarkdown('テスト', JSON.stringify(doc));

    expect(createURL).toHaveBeenCalled();
    expect(mockClick).toHaveBeenCalled();
    expect(revokeURL).toHaveBeenCalledWith('blob:test');

    createElementSpy.mockRestore();
    appendSpy.mockRestore();
    removeSpy.mockRestore();
  });

  it('copyAsMarkdownが複数段落を正しく変換する', () => {
    const { result } = renderHook(() => useMarkdownExport());
    const doc = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: '段落1' }] },
        { type: 'paragraph', content: [{ type: 'text', text: '段落2' }] },
      ],
    };
    const md = result.current.copyAsMarkdown('タイトル', JSON.stringify(doc));
    expect(md).toBe('# タイトル\n\n段落1\n\n段落2');
  });

  it('exportAsMarkdownがタイトルなしの場合「無題.md」をファイル名にする', () => {
    const { result } = renderHook(() => useMarkdownExport());

    let downloadFileName = '';
    const originalCreateElement = document.createElement.bind(document);
    const createElementSpy = vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
      if (tag === 'a') {
        return {
          click: vi.fn(),
          href: '',
          set download(val: string) { downloadFileName = val; },
          get download() { return downloadFileName; },
        } as unknown as HTMLAnchorElement;
      }
      return originalCreateElement(tag);
    });
    const appendSpy = vi.spyOn(document.body, 'appendChild').mockImplementation((node) => node);
    const removeSpy = vi.spyOn(document.body, 'removeChild').mockImplementation((node) => node);
    globalThis.URL.createObjectURL = vi.fn(() => 'blob:test');
    globalThis.URL.revokeObjectURL = vi.fn();

    const doc = { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: 'テスト' }] }] };
    result.current.exportAsMarkdown('', JSON.stringify(doc));

    expect(downloadFileName).toBe('無題.md');

    createElementSpy.mockRestore();
    appendSpy.mockRestore();
    removeSpy.mockRestore();
  });
});
