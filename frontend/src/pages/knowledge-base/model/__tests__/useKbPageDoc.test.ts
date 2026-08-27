import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useKbPageDoc } from '../useKbPageDoc';

const hoisted = vi.hoisted(() => ({ fetchPage: vi.fn() }));

vi.mock('@/entities/knowledge-base', () => ({
  KnowledgeBaseRepository: { fetchPage: hoisted.fetchPage },
}));

const pageDoc = (title: string) => ({
  page: {
    id: 'p1',
    spaceId: 's1',
    position: 'a0',
    title,
    createdByUserId: 1,
    createdAt: '2026-08-01T00:00:00Z',
    updatedAt: '2026-08-01T00:00:00Z',
  },
  doc: { type: 'doc', content: [] },
});

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.fetchPage.mockResolvedValue(pageDoc('設計メモ'));
});

describe('useKbPageDoc', () => {
  it('ページ ID が無ければ何も取りに行かない', () => {
    renderHook(() => useKbPageDoc('acme', undefined));

    expect(hoisted.fetchPage).not.toHaveBeenCalled();
  });

  it('本文を取ってくる', async () => {
    const { result } = renderHook(() => useKbPageDoc('acme', 'p1'));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.data?.page.title).toBe('設計メモ');
    expect(result.current.error).toBeNull();
  });

  it('失敗しても「見る権限がありません」とは言わない', async () => {
    // backend は「無い」と「見えない」を撃ち分けていない（撃ち分けると ID の総当たりで
    // 実在が分かる）。フロントで名指しすると、そこだけが隠していることを喋る。
    hoisted.fetchPage.mockRejectedValue(new Error('404'));

    const { result } = renderHook(() => useKbPageDoc('acme', 'p1'));

    await waitFor(() => expect(result.current.error).not.toBeNull());
    expect(result.current.error).not.toMatch(/権限/);
    expect(result.current.data).toBeNull();
  });

  it('速く行き来しても、古い応答が新しいページを上書きしない', async () => {
    let resolveOld: (value: unknown) => void = () => {};
    hoisted.fetchPage.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveOld = resolve;
        }),
    );

    const { result, rerender } = renderHook(({ id }) => useKbPageDoc('acme', id), {
      initialProps: { id: 'old' },
    });

    hoisted.fetchPage.mockResolvedValue(pageDoc('新しいページ'));
    rerender({ id: 'new' });
    await waitFor(() => expect(result.current.data?.page.title).toBe('新しいページ'));

    // 先に投げた要求がいま返ってくる。後から届いても採用してはいけない。
    //
    // act で包んで**解決を実際に流し切ってから**確かめる。ここを waitFor にすると、
    // 最初の判定が古い応答の反映より先に走って必ず通り、検査として意味を成さない。
    await act(async () => {
      resolveOld(pageDoc('古いページ'));
    });
    expect(result.current.data?.page.title).toBe('新しいページ');
  });
});
