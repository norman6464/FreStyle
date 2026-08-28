import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useNotePageDoc } from '../useNotePageDoc';

const hoisted = vi.hoisted(() => ({
  resolvePage: vi.fn(),
  replaceContent: vi.fn(),
}));

vi.mock('@/entities/note', () => ({
  NoteRepository: {
    resolvePage: hoisted.resolvePage,
    replaceContent: hoisted.replaceContent,
  },
}));

const resolved = (title: string, canEdit = true) => ({
  workspaceSlug: 'w-3f2a9c',
  page: {
    id: 'p1',
    spaceId: 's1',
    title,
    createdByUserId: 1,
    createdAt: '2026-08-01T00:00:00Z',
    updatedAt: '2026-08-01T00:00:00Z',
  },
  doc: { type: 'doc', content: [] },
  canEdit,
});

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.resolvePage.mockResolvedValue(resolved('設計メモ'));
  hoisted.replaceContent.mockResolvedValue({ doc: { type: 'doc', content: [] }, builtAt: '2026-08-27T00:00:00Z' });
});

describe('useNotePageDoc', () => {
  it('ページ ID が無ければ何も取りに行かない', () => {
    renderHook(() => useNotePageDoc(undefined));

    expect(hoisted.resolvePage).not.toHaveBeenCalled();
  });

  it('ID だけでページと所属ワークスペースを解決する', async () => {
    const { result } = renderHook(() => useNotePageDoc('p1'));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.data?.page.title).toBe('設計メモ');
    expect(result.current.data?.workspaceSlug).toBe('w-3f2a9c');
    expect(result.current.data?.canEdit).toBe(true);
    expect(result.current.error).toBeNull();
  });

  it('失敗しても「見る権限がありません」とは言わない', async () => {
    // backend は「無い」と「見えない」を撃ち分けていない（撃ち分けると ID の総当たりで
    // 実在が分かる）。フロントで名指しすると、そこだけが隠していることを喋る。
    hoisted.resolvePage.mockRejectedValue(new Error('404'));

    const { result } = renderHook(() => useNotePageDoc('p1'));

    await waitFor(() => expect(result.current.error).not.toBeNull());
    expect(result.current.error).not.toMatch(/権限/);
    expect(result.current.data).toBeNull();
  });

  it('速く行き来しても、古い応答が新しいページを上書きしない', async () => {
    let resolveOld: (value: unknown) => void = () => {};
    hoisted.resolvePage.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveOld = resolve;
        }),
    );

    const { result, rerender } = renderHook(({ id }) => useNotePageDoc(id), {
      initialProps: { id: 'old' },
    });

    hoisted.resolvePage.mockResolvedValue(resolved('新しいページ'));
    rerender({ id: 'new' });
    await waitFor(() => expect(result.current.data?.page.title).toBe('新しいページ'));

    // 先に投げた要求がいま返ってくる。後から届いても採用してはいけない。
    //
    // act で包んで**解決を実際に流し切ってから**確かめる。ここを waitFor にすると、
    // 最初の判定が古い応答の反映より先に走って必ず通り、検査として意味を成さない。
    await act(async () => {
      resolveOld(resolved('古いページ'));
    });
    expect(result.current.data?.page.title).toBe('新しいページ');
  });

  it('書くとデバウンス後に保存し、状態が unsaved → saving → saved と動く', async () => {
    vi.useFakeTimers();
    try {
      const { result } = renderHook(() => useNotePageDoc('p1'));

      // 解決を流し切る（fake timer 下なので waitFor は使わない）。
      await act(async () => {});
      expect(result.current.data).not.toBeNull();

      act(() => {
        result.current.onDocChange({ type: 'doc', content: [{ type: 'paragraph' }] });
      });
      expect(result.current.saveStatus).toBe('unsaved');
      expect(hoisted.replaceContent).not.toHaveBeenCalled();

      // デバウンスの間合いを越えると 1 回だけ保存が飛ぶ。
      await act(async () => {
        vi.advanceTimersByTime(1000);
      });
      expect(hoisted.replaceContent).toHaveBeenCalledTimes(1);
      expect(hoisted.replaceContent).toHaveBeenCalledWith(
        'w-3f2a9c',
        'p1',
        { type: 'doc', content: [{ type: 'paragraph' }] },
      );
      expect(result.current.saveStatus).toBe('saved');
    } finally {
      vi.useRealTimers();
    }
  });

  it('保存が失敗したら unsaved に戻す（保存できた顔をしない）', async () => {
    vi.useFakeTimers();
    try {
      hoisted.replaceContent.mockRejectedValue(new Error('500'));
      const { result } = renderHook(() => useNotePageDoc('p1'));
      await act(async () => {});

      act(() => {
        result.current.onDocChange({ type: 'doc', content: [] });
      });
      await act(async () => {
        vi.advanceTimersByTime(1000);
      });

      expect(result.current.saveStatus).toBe('unsaved');
    } finally {
      vi.useRealTimers();
    }
  });
});
