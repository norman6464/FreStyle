import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useCommentEdits } from '../useCommentEdits';

const hoisted = vi.hoisted(() => ({ fetchTicketCommentEdits: vi.fn() }));

vi.mock('@/entities/ticket', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/ticket')>();
  return {
    ...actual,
    TicketRepository: { fetchTicketCommentEdits: hoisted.fetchTicketCommentEdits },
  };
});

beforeEach(() => {
  vi.clearAllMocks();
});

describe('useCommentEdits', () => {
  it('load を呼ぶまで何もしない', () => {
    renderHook(() => useCommentEdits('acme', 't-1', 'c-1'));
    expect(hoisted.fetchTicketCommentEdits).not.toHaveBeenCalled();
  });

  it('load すると編集履歴を引く', async () => {
    hoisted.fetchTicketCommentEdits.mockResolvedValue([
      { id: 'e-1', editor: { userId: 1, name: '田中 太郎' }, previousBody: [], editedAt: '' },
    ]);
    const { result } = renderHook(() => useCommentEdits('acme', 't-1', 'c-1'));

    await act(async () => {
      await result.current.load();
    });

    expect(result.current.edits).toHaveLength(1);
    expect(hoisted.fetchTicketCommentEdits).toHaveBeenCalledWith('acme', 't-1', 'c-1');
  });

  it('失敗したら文言を出す', async () => {
    hoisted.fetchTicketCommentEdits.mockRejectedValue(new Error('x'));
    const { result } = renderHook(() => useCommentEdits('acme', 't-1', 'c-1'));

    await act(async () => {
      await result.current.load();
    });

    expect(result.current.error).toBe('編集履歴を読み込めませんでした。');
  });

  it('連打しても最後の呼び出しの結果だけを反映する', async () => {
    let resolveFirst!: (v: unknown[]) => void;
    hoisted.fetchTicketCommentEdits.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveFirst = resolve;
        }),
    );
    hoisted.fetchTicketCommentEdits.mockResolvedValueOnce([
      { id: 'e-2', editor: { userId: 1, name: 'x' }, previousBody: [], editedAt: '' },
    ]);

    const { result } = renderHook(() => useCommentEdits('acme', 't-1', 'c-1'));

    let firstLoad: Promise<void>;
    act(() => {
      firstLoad = result.current.load();
    });
    await act(async () => {
      await result.current.load();
    });

    resolveFirst([{ id: 'e-1-old', editor: { userId: 1, name: 'x' }, previousBody: [], editedAt: '' }]);
    await act(async () => {
      await firstLoad;
    });

    expect(result.current.edits.map((e) => e.id)).toEqual(['e-2']);
  });
});
