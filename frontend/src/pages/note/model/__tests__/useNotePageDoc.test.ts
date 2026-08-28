import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useNotePageDoc } from '../useNotePageDoc';

const hoisted = vi.hoisted(() => ({
  resolvePage: vi.fn(),
  replaceContent: vi.fn(),
  renamePage: vi.fn(),
  emit: vi.fn(),
}));

vi.mock('@/entities/note', () => ({
  NoteRepository: {
    resolvePage: hoisted.resolvePage,
    replaceContent: hoisted.replaceContent,
    renamePage: hoisted.renamePage,
  },
  emitNoteTreeEvent: hoisted.emit,
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

  it('前の保存が終わるまで次の PUT を送らない（丸ごと置換なので順序が命）', async () => {
    // 並行に送ると、後から書いた本文の PUT が先に完了し、古い本文の PUT が
    // 後から着地して上書きし得る。送信は必ず 1 本ずつ・編集順で。
    vi.useFakeTimers();
    try {
      let resolveFirst: (value: unknown) => void = () => {};
      hoisted.replaceContent.mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveFirst = resolve;
          }),
      );

      const { result } = renderHook(() => useNotePageDoc('p1'));
      await act(async () => {});

      // 1 回目の編集 → デバウンス発火で PUT(A) が飛ぶ（保留のまま）。
      act(() => {
        result.current.onDocChange({ type: 'doc', content: [{ type: 'paragraph' }] });
      });
      await act(async () => {
        vi.advanceTimersByTime(1000);
      });
      expect(hoisted.replaceContent).toHaveBeenCalledTimes(1);

      // PUT(A) が飛んでいる間に 2 回目の編集 → デバウンスが切れても送らない。
      act(() => {
        result.current.onDocChange({ type: 'doc', content: [] });
      });
      await act(async () => {
        vi.advanceTimersByTime(1000);
      });
      expect(hoisted.replaceContent).toHaveBeenCalledTimes(1);

      // PUT(A) が完了したら、残っていた本文が続けて送られる。
      await act(async () => {
        resolveFirst({ doc: { type: 'doc', content: [] }, builtAt: '2026-08-28T00:00:00Z' });
      });
      expect(hoisted.replaceContent).toHaveBeenCalledTimes(2);
      expect(hoisted.replaceContent).toHaveBeenLastCalledWith('w-3f2a9c', 'p1', {
        type: 'doc',
        content: [],
      });
      expect(result.current.saveStatus).toBe('saved');
    } finally {
      vi.useRealTimers();
    }
  });

  it('renameTitle は改名し、画面の題名を確定後の値へ差し替え、木にも知らせる', async () => {
    const renamed = {
      id: 'p1',
      spaceId: 's1',
      title: '設計メモ v2',
      createdByUserId: 1,
      createdAt: '2026-08-01T00:00:00Z',
      updatedAt: '2026-08-28T00:00:00Z',
    };
    hoisted.renamePage.mockResolvedValue(renamed);
    const { result } = renderHook(() => useNotePageDoc('p1'));
    await waitFor(() => expect(result.current.data).not.toBeNull());

    await act(async () => {
      await result.current.renameTitle('設計メモ v2');
    });

    expect(hoisted.renamePage).toHaveBeenCalledWith('w-3f2a9c', 'p1', '設計メモ v2');
    expect(result.current.data?.page.title).toBe('設計メモ v2');
    expect(hoisted.emit).toHaveBeenCalledWith({ type: 'page-renamed', page: renamed });
  });

  it('renameTitle の失敗は投げ、画面の題名は変えない', async () => {
    hoisted.renamePage.mockRejectedValue(new Error('403'));
    const { result } = renderHook(() => useNotePageDoc('p1'));
    await waitFor(() => expect(result.current.data).not.toBeNull());

    await expect(result.current.renameTitle('だめな改名')).rejects.toThrow();
    expect(result.current.data?.page.title).toBe('設計メモ');
    expect(hoisted.emit).not.toHaveBeenCalled();
  });

  it('ページを移っても、書きかけの保存は**書いた時点のページ**へ送る（移った先を潰さない）', async () => {
    // 旧: 宛先を送信時に読む → 移った先の resolve が宛先を差し替え、旧ページの全文が
    // 新ページへ PUT され、丸ごと置換なので新ページの本文が消えていた。
    vi.useFakeTimers();
    try {
      let resolveFirstPut: (value: unknown) => void = () => {};
      hoisted.replaceContent.mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveFirstPut = resolve;
          }),
      );

      const { result, rerender } = renderHook(({ id }) => useNotePageDoc(id), {
        initialProps: { id: 'p1' },
      });
      await act(async () => {});

      // p1 で書く → PUT(A) が飛ぶ（保留）。さらに書いて残りを作る。
      act(() => {
        result.current.onDocChange({ type: 'doc', content: [{ type: 'paragraph' }] });
      });
      await act(async () => {
        vi.advanceTimersByTime(1000);
      });
      act(() => {
        result.current.onDocChange({ type: 'doc', content: [] });
      });

      // 子ページ p2 へ移る（resolve は p2 を返し、宛先 ref は p2 に差し替わる）。
      hoisted.resolvePage.mockResolvedValue({
        ...resolved('子ページ'),
        page: { ...resolved('子ページ').page, id: 'p2' },
      });
      rerender({ id: 'p2' });
      await act(async () => {});

      // PUT(A) が完了 → 残りが流れる。宛先は**書いた時点の p1** でなければならない。
      await act(async () => {
        resolveFirstPut({ doc: { type: 'doc', content: [] }, builtAt: '2026-08-28T00:00:00Z' });
      });
      expect(hoisted.replaceContent).toHaveBeenCalledTimes(2);
      expect(hoisted.replaceContent).toHaveBeenLastCalledWith('w-3f2a9c', 'p1', {
        type: 'doc',
        content: [],
      });
    } finally {
      vi.useRealTimers();
    }
  });

  it('改名の応答が別ページへ移った後に返っても、移った先の表示を上書きしない', async () => {
    let resolveRename: (value: unknown) => void = () => {};
    hoisted.renamePage.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveRename = resolve;
        }),
    );
    const { result, rerender } = renderHook(({ id }) => useNotePageDoc(id), {
      initialProps: { id: 'p1' },
    });
    await waitFor(() => expect(result.current.data).not.toBeNull());

    // p1 の改名を送ったまま p2 へ移る。
    const renamePromise = result.current.renameTitle('旧ページの新題名');
    hoisted.resolvePage.mockResolvedValue({
      ...resolved('別ページ'),
      page: { ...resolved('別ページ').page, id: 'p2' },
    });
    rerender({ id: 'p2' });
    await waitFor(() => expect(result.current.data?.page.id).toBe('p2'));

    // 遅れて p1 の改名応答が着地しても、p2 の表示はそのまま。
    await act(async () => {
      resolveRename({
        id: 'p1',
        spaceId: 's1',
        title: '旧ページの新題名',
        createdByUserId: 1,
        createdAt: '2026-08-01T00:00:00Z',
        updatedAt: '2026-08-28T00:00:00Z',
      });
      await renamePromise;
    });
    expect(result.current.data?.page.id).toBe('p2');
    expect(result.current.data?.page.title).toBe('別ページ');
    // 改名自体はサーバーで成立しているので、木への知らせは出す。
    expect(hoisted.emit).toHaveBeenCalledWith(
      expect.objectContaining({ type: 'page-renamed' }),
    );
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
