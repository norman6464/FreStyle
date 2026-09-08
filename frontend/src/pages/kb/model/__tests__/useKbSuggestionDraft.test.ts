import { renderHook, act, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useKbSuggestionDraft } from '../useKbSuggestionDraft';

const hoisted = vi.hoisted(() => ({ createSuggestion: vi.fn() }));

vi.mock('@/entities/kb', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/kb')>();
  return { ...actual, KbRepository: { createSuggestion: hoisted.createSuggestion } };
});

const doc = { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: '本文' }] }] };

beforeEach(() => {
  vi.clearAllMocks();
});

describe('useKbSuggestionDraft', () => {
  it('初期状態は閉じている', () => {
    const { result } = renderHook(() => useKbSuggestionDraft('w-1', 'p-1'));
    expect(result.current.open).toBe(false);
    expect(result.current.draft).toBeNull();
  });

  it('start でドラフトモードに入り、渡した doc が下書きの初期値になる', () => {
    const { result } = renderHook(() => useKbSuggestionDraft('w-1', 'p-1'));
    act(() => result.current.start(doc));
    expect(result.current.open).toBe(true);
    expect(result.current.draft).toEqual(doc);
  });

  it('changeDraft で下書きが更新される（APIへは送らない）', () => {
    const { result } = renderHook(() => useKbSuggestionDraft('w-1', 'p-1'));
    act(() => result.current.start(doc));
    const changed = { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: '書き換え後' }] }] };
    act(() => result.current.changeDraft(changed));
    expect(result.current.draft).toEqual(changed);
    expect(hoisted.createSuggestion).not.toHaveBeenCalled();
  });

  it('cancel でドラフトを破棄して閉じる', () => {
    const { result } = renderHook(() => useKbSuggestionDraft('w-1', 'p-1'));
    act(() => result.current.start(doc));
    act(() => result.current.cancel());
    expect(result.current.open).toBe(false);
    expect(result.current.draft).toBeNull();
    expect(hoisted.createSuggestion).not.toHaveBeenCalled();
  });

  it('submit は1回だけ createSuggestion を呼び、成功したら閉じて true を返す', async () => {
    hoisted.createSuggestion.mockResolvedValue({ id: 's-1', doc, status: 'open', author: { userId: 1, name: '' }, createdAt: '2026-09-01T00:00:00Z' });
    const { result } = renderHook(() => useKbSuggestionDraft('w-1', 'p-1'));
    act(() => result.current.start(doc));

    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.submit();
    });

    expect(ok).toBe(true);
    expect(hoisted.createSuggestion).toHaveBeenCalledTimes(1);
    expect(hoisted.createSuggestion).toHaveBeenCalledWith('w-1', 'p-1', doc);
    expect(result.current.open).toBe(false);
  });

  it('submit が失敗したらドラフトモードのまま、入力を保持し、エラーを持つ', async () => {
    hoisted.createSuggestion.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useKbSuggestionDraft('w-1', 'p-1'));
    act(() => result.current.start(doc));

    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.submit();
    });

    expect(ok).toBe(false);
    expect(result.current.open).toBe(true);
    expect(result.current.draft).toEqual(doc);
    expect(result.current.error).toBe('提案を送信できませんでした。');
  });

  it('ページを移ったら書きかけの下書きを持ち越さない', async () => {
    const { result, rerender } = renderHook(
      ({ pageId }: { pageId: string }) => useKbSuggestionDraft('w-1', pageId),
      { initialProps: { pageId: 'p-1' } },
    );
    act(() => result.current.start(doc));
    expect(result.current.open).toBe(true);

    rerender({ pageId: 'p-2' });
    await waitFor(() => expect(result.current.open).toBe(false));
    expect(result.current.draft).toBeNull();
  });
});
