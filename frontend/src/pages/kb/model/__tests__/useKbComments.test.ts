import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useKbComments } from '../useKbComments';

const hoisted = vi.hoisted(() => ({
  listCommentThreads: vi.fn(),
  createCommentThread: vi.fn(),
  addComment: vi.fn(),
  resolveCommentThread: vi.fn(),
  reopenCommentThread: vi.fn(),
}));

vi.mock('@/entities/kb', () => ({
  KbRepository: {
    listCommentThreads: hoisted.listCommentThreads,
    createCommentThread: hoisted.createCommentThread,
    addComment: hoisted.addComment,
    resolveCommentThread: hoisted.resolveCommentThread,
    reopenCommentThread: hoisted.reopenCommentThread,
  },
}));

const SLUG = 'w-3f2a9c';
const PAGE = 'p1';

const author = (name = '田中 太郎', userId = 1) => ({ userId, name });

const comment = (id: string, text: string) => ({
  id,
  author: author(),
  body: [{ type: 'text', text }],
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
});

const thread = (id: string, overrides: Partial<ReturnType<typeof baseThread>> = {}) => ({
  ...baseThread(id),
  ...overrides,
});

function baseThread(id: string) {
  return {
    id,
    createdBy: author(),
    resolvedAt: null as string | null,
    resolvedBy: null as null | ReturnType<typeof author>,
    createdAt: '2026-09-01T00:00:00Z',
    comments: [comment(`${id}-c1`, '質問です')],
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.listCommentThreads.mockResolvedValue([thread('t1')]);
  hoisted.createCommentThread.mockResolvedValue(thread('t2'));
  hoisted.addComment.mockResolvedValue(comment('t1-c2', '返信です'));
  hoisted.resolveCommentThread.mockResolvedValue(
    thread('t1', { resolvedAt: '2026-09-02T00:00:00Z', resolvedBy: author('鈴木 花子', 2) }),
  );
  hoisted.reopenCommentThread.mockResolvedValue(thread('t1'));
});

describe('useKbComments', () => {
  it('閉じている間は取りに行かない', () => {
    renderHook(() => useKbComments(SLUG, PAGE, false));

    expect(hoisted.listCommentThreads).not.toHaveBeenCalled();
  });

  it('ページが決まっていなければ、開いていても取りに行かない', () => {
    renderHook(() => useKbComments(undefined, undefined, true));

    expect(hoisted.listCommentThreads).not.toHaveBeenCalled();
  });

  it('開くと取得する', async () => {
    const { result } = renderHook(() => useKbComments(SLUG, PAGE, true));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(hoisted.listCommentThreads).toHaveBeenCalledWith(SLUG, PAGE);
    expect(result.current.threads).toEqual([thread('t1')]);
  });

  it('読み込みに失敗したら理由を出し、古いスレッドを残さない', async () => {
    hoisted.listCommentThreads.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useKbComments(SLUG, PAGE, true));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toMatch(/コメントを読み込めませんでした/);
    expect(result.current.threads).toHaveLength(0);
  });

  describe('createThread', () => {
    it('成功したら応答のスレッドを末尾に足す', async () => {
      const { result } = renderHook(() => useKbComments(SLUG, PAGE, true));
      await waitFor(() => expect(result.current.loading).toBe(false));

      const body = [{ type: 'text', text: '新しいスレッド' }];
      await act(async () => {
        await result.current.createThread(body);
      });

      expect(hoisted.createCommentThread).toHaveBeenCalledWith(SLUG, PAGE, body);
      expect(result.current.threads.map((t) => t.id)).toEqual(['t1', 't2']);
    });

    it('失敗は投げる。saving は元に戻り、一覧は変わらない', async () => {
      hoisted.createCommentThread.mockRejectedValue(new Error('forbidden'));
      const { result } = renderHook(() => useKbComments(SLUG, PAGE, true));
      await waitFor(() => expect(result.current.loading).toBe(false));

      await expect(
        act(async () => {
          await result.current.createThread([{ type: 'text', text: 'x' }]);
        }),
      ).rejects.toThrow();

      expect(result.current.saving).toBe(false);
      expect(result.current.threads).toHaveLength(1);
    });
  });

  describe('reply', () => {
    it('成功したら該当スレッドの comments に追加する（他のスレッドは変えない）', async () => {
      hoisted.listCommentThreads.mockResolvedValue([thread('t1'), thread('t2')]);
      const { result } = renderHook(() => useKbComments(SLUG, PAGE, true));
      await waitFor(() => expect(result.current.threads).toHaveLength(2));

      const body = [{ type: 'text', text: '返信です' }];
      await act(async () => {
        await result.current.reply('t1', body);
      });

      expect(hoisted.addComment).toHaveBeenCalledWith(SLUG, PAGE, 't1', body);
      const t1 = result.current.threads.find((t) => t.id === 't1');
      const t2 = result.current.threads.find((t) => t.id === 't2');
      expect(t1?.comments.map((c) => c.id)).toEqual(['t1-c1', 't1-c2']);
      expect(t2?.comments.map((c) => c.id)).toEqual(['t2-c1']);
    });

    it('失敗は投げる', async () => {
      hoisted.addComment.mockRejectedValue(new Error('forbidden'));
      const { result } = renderHook(() => useKbComments(SLUG, PAGE, true));
      await waitFor(() => expect(result.current.loading).toBe(false));

      await expect(
        act(async () => {
          await result.current.reply('t1', [{ type: 'text', text: 'x' }]);
        }),
      ).rejects.toThrow();
      expect(result.current.saving).toBe(false);
    });
  });

  describe('resolve / reopen', () => {
    it('resolve は応答の解決状態（resolvedAt / resolvedBy）だけを該当スレッドへ差し込む', async () => {
      const { result } = renderHook(() => useKbComments(SLUG, PAGE, true));
      await waitFor(() => expect(result.current.loading).toBe(false));

      await act(async () => {
        await result.current.resolve('t1');
      });

      expect(hoisted.resolveCommentThread).toHaveBeenCalledWith(SLUG, PAGE, 't1');
      expect(result.current.threads[0].resolvedAt).toBe('2026-09-02T00:00:00Z');
      expect(result.current.threads[0].resolvedBy).toEqual(author('鈴木 花子', 2));
    });

    it('reopen は応答の解決状態だけを該当スレッドへ差し込む', async () => {
      hoisted.listCommentThreads.mockResolvedValue([
        thread('t1', { resolvedAt: '2026-09-02T00:00:00Z', resolvedBy: author('鈴木 花子', 2) }),
      ]);
      const { result } = renderHook(() => useKbComments(SLUG, PAGE, true));
      await waitFor(() => expect(result.current.loading).toBe(false));

      await act(async () => {
        await result.current.reopen('t1');
      });

      expect(hoisted.reopenCommentThread).toHaveBeenCalledWith(SLUG, PAGE, 't1');
      expect(result.current.threads[0].resolvedAt).toBeNull();
    });

    it('resolve/reopen の応答が空の comments を返しても、手元の発言一覧は消えない（サーバーは発言を引き直さない設計）', async () => {
      // 実際の backend は resolve/reopen の応答で comments を常に空配列で返す
      // （発言を引き直さない設計）。スレッドを丸ごと差し替えると、解決するたびに
      // 表示中の発言が消えてしまう — 解決状態の欄だけを差し込むことで防ぐ。
      hoisted.resolveCommentThread.mockResolvedValue(
        thread('t1', {
          resolvedAt: '2026-09-02T00:00:00Z',
          resolvedBy: author('鈴木 花子', 2),
          comments: [],
        }),
      );
      const { result } = renderHook(() => useKbComments(SLUG, PAGE, true));
      await waitFor(() => expect(result.current.threads[0].comments).toHaveLength(1));

      await act(async () => {
        await result.current.resolve('t1');
      });

      expect(result.current.threads[0].comments).toEqual([comment('t1-c1', '質問です')]);
      expect(result.current.threads[0].resolvedAt).toBe('2026-09-02T00:00:00Z');
    });

    it('失敗は投げる', async () => {
      hoisted.resolveCommentThread.mockRejectedValue(new Error('forbidden'));
      const { result } = renderHook(() => useKbComments(SLUG, PAGE, true));
      await waitFor(() => expect(result.current.loading).toBe(false));

      await expect(
        act(async () => {
          await result.current.resolve('t1');
        }),
      ).rejects.toThrow();
      expect(result.current.saving).toBe(false);
    });
  });
});

describe('useKbComments の宛先', () => {
  it('宛先が無くなったら状態を畳む（次に開いたとき前のページのスレッドを出さない）', async () => {
    const { result, rerender } = renderHook(
      ({ open }: { open: boolean }) => useKbComments(SLUG, PAGE, open),
      { initialProps: { open: true } },
    );
    await waitFor(() => expect(result.current.threads).toHaveLength(1));

    rerender({ open: false });

    expect(result.current.threads).toHaveLength(0);
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('閉じたあとに着地した読み込み応答は捨てる', async () => {
    let settle: (value: unknown) => void = () => {};
    hoisted.listCommentThreads.mockImplementationOnce(
      () => new Promise((resolve) => { settle = resolve; }),
    );

    const { result, rerender } = renderHook(
      ({ open }: { open: boolean }) => useKbComments(SLUG, PAGE, open),
      { initialProps: { open: true } },
    );
    rerender({ open: false });

    await act(async () => {
      settle([thread('t1')]);
    });

    expect(result.current.threads).toHaveLength(0);
    expect(result.current.loading).toBe(false);
  });

  it('速く開き直したとき、古い応答で新しい結果を上書きしない（別ページへの移動）', async () => {
    let settleFirst: (value: unknown) => void = () => {};
    hoisted.listCommentThreads.mockImplementationOnce(
      () => new Promise((resolve) => { settleFirst = resolve; }),
    );

    const { result, rerender } = renderHook(
      ({ page }: { page: string }) => useKbComments(SLUG, page, true),
      { initialProps: { page: 'p-old' } },
    );

    hoisted.listCommentThreads.mockResolvedValue([thread('t-new')]);
    rerender({ page: 'p-new' });
    await waitFor(() => expect(result.current.threads).toHaveLength(1));
    expect(result.current.threads[0].id).toBe('t-new');

    // 遅れて着地した旧ページの応答は捨てる。
    await act(async () => {
      settleFirst([thread('t-old')]);
    });
    expect(result.current.threads.map((t) => t.id)).toEqual(['t-new']);
  });

  it('書き込み中に宛先が変わったら、応答が返っても状態に反映しない', async () => {
    let finishResolve: (value: unknown) => void = () => {};
    hoisted.resolveCommentThread.mockImplementationOnce(
      () => new Promise((resolve) => { finishResolve = resolve; }),
    );

    const { result, rerender } = renderHook(
      ({ open }: { open: boolean }) => useKbComments(SLUG, PAGE, open),
      { initialProps: { open: true } },
    );
    await waitFor(() => expect(result.current.loading).toBe(false));

    const resolving = result.current.resolve('t1');
    await waitFor(() => expect(result.current.saving).toBe(true));

    // 応答が返る前にパネルを閉じる。
    rerender({ open: false });

    await act(async () => {
      finishResolve(
        thread('t1', { resolvedAt: '2026-09-02T00:00:00Z', resolvedBy: author('鈴木 花子', 2) }),
      );
      await resolving;
    });

    // 畳んだあとの状態には触れていない。
    expect(result.current.threads).toHaveLength(0);
    expect(result.current.saving).toBe(false);
  });

  it('同じページへの古い読み込みが後から着地しても捨てる（連番）', async () => {
    let settleFirst: (value: unknown) => void = () => {};
    hoisted.listCommentThreads.mockImplementationOnce(
      () => new Promise((resolve) => { settleFirst = resolve; }),
    );

    const { result, rerender } = renderHook(
      ({ open }: { open: boolean }) => useKbComments(SLUG, PAGE, open),
      { initialProps: { open: true } },
    );

    // 1 本目が飛んでいる間に、いったん閉じてまた開く（2 本目を起こす）。
    rerender({ open: false });
    hoisted.listCommentThreads.mockResolvedValue([thread('t-second')]);
    rerender({ open: true });
    await waitFor(() => expect(result.current.threads).toHaveLength(1));
    expect(result.current.threads[0].id).toBe('t-second');

    // 遅れて着地した 1 本目は捨てる。
    await act(async () => {
      settleFirst([thread('t-first')]);
    });
    expect(result.current.threads.map((t) => t.id)).toEqual(['t-second']);
  });
});
