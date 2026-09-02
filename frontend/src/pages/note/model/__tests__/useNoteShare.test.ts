import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useNoteShare } from '../useNoteShare';

const hoisted = vi.hoisted(() => ({
  listPageGrants: vi.fn(),
  listGrantablePrincipals: vi.fn(),
  grantPageRole: vi.fn(),
  revokePageRole: vi.fn(),
}));

vi.mock('@/entities/note', () => ({
  NoteRepository: {
    listPageGrants: hoisted.listPageGrants,
    listGrantablePrincipals: hoisted.listGrantablePrincipals,
    grantPageRole: hoisted.grantPageRole,
    revokePageRole: hoisted.revokePageRole,
  },
}));

const SLUG = 'w-3f2a9c';
const PAGE = 'p1';

const grant = (principalId: string, role = 'editor') => ({
  pageId: PAGE,
  principalId,
  role,
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
});

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.listPageGrants.mockResolvedValue([grant('pr-tanaka')]);
  hoisted.listGrantablePrincipals.mockResolvedValue([
    { id: 'pr-tanaka', kind: 'user', name: '田中 太郎' },
    { id: 'pr-dev', kind: 'group', name: '開発チーム' },
  ]);
  hoisted.grantPageRole.mockResolvedValue(grant('pr-dev'));
  hoisted.revokePageRole.mockResolvedValue(undefined);
});

describe('useNoteShare', () => {
  it('ページが決まっていなければ何も取りに行かない', () => {
    renderHook(() => useNoteShare(undefined, undefined));

    expect(hoisted.listPageGrants).not.toHaveBeenCalled();
    expect(hoisted.listGrantablePrincipals).not.toHaveBeenCalled();
  });

  it('主体 ID で突き合わせて表示名を付ける', async () => {
    const { result } = renderHook(() => useNoteShare(SLUG, PAGE));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.rows).toEqual([
      { principalId: 'pr-tanaka', role: 'editor', name: '田中 太郎', kind: 'user' },
    ]);
  });

  it('相手の一覧に居ない主体でも行を落とさない', async () => {
    // 引いた直後に主体が消えるとこうなる。ここで行を消すと、取り消せない権限が
    // 画面から見えないまま残る（誰が見られるのかを人が説明できなくなる）。
    hoisted.listPageGrants.mockResolvedValue([grant('pr-gone')]);
    const { result } = renderHook(() => useNoteShare(SLUG, PAGE));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.rows).toEqual([
      { principalId: 'pr-gone', role: 'editor', name: '', kind: 'unknown' },
    ]);
  });

  it('すでに権限を持っている相手は追加の候補から外す', async () => {
    const { result } = renderHook(() => useNoteShare(SLUG, PAGE));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.candidates.map((c) => c.id)).toEqual(['pr-dev']);
  });

  it('付与のあとは引き直す（画面と実態をずらさない）', async () => {
    // 弱い役割を張っても上位の役割で上書きされない、といった規則がサーバーにあるので、
    // 楽観的に画面だけ書き換えるとずれたまま残る。
    const { result } = renderHook(() => useNoteShare(SLUG, PAGE));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(hoisted.listPageGrants).toHaveBeenCalledTimes(1);

    hoisted.listPageGrants.mockResolvedValue([grant('pr-tanaka'), grant('pr-dev', 'viewer')]);
    await act(async () => {
      await result.current.grant('pr-dev', 'viewer');
    });

    expect(hoisted.grantPageRole).toHaveBeenCalledWith(SLUG, PAGE, 'pr-dev', 'viewer');
    expect(hoisted.listPageGrants).toHaveBeenCalledTimes(2);
    await waitFor(() => expect(result.current.rows).toHaveLength(2));
  });

  it('取り消しのあとも引き直す', async () => {
    const { result } = renderHook(() => useNoteShare(SLUG, PAGE));
    await waitFor(() => expect(result.current.loading).toBe(false));

    hoisted.listPageGrants.mockResolvedValue([]);
    await act(async () => {
      await result.current.revoke('pr-tanaka');
    });

    expect(hoisted.revokePageRole).toHaveBeenCalledWith(SLUG, PAGE, 'pr-tanaka');
    await waitFor(() => expect(result.current.rows).toHaveLength(0));
  });

  it('読み込みに失敗したら理由を出し、古い行を残さない', async () => {
    hoisted.listPageGrants.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useNoteShare(SLUG, PAGE));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toMatch(/権限を読めませんでした/);
    expect(result.current.rows).toHaveLength(0);
  });

  it('書き込みに失敗したら理由を出し、行はそのまま残す', async () => {
    const { result } = renderHook(() => useNoteShare(SLUG, PAGE));
    await waitFor(() => expect(result.current.loading).toBe(false));

    hoisted.grantPageRole.mockRejectedValue(new Error('boom'));
    await act(async () => {
      await result.current.grant('pr-dev', 'viewer');
    });

    expect(result.current.error).toMatch(/権限を変えられませんでした/);
    expect(result.current.saving).toBe(false);
    // 失敗した書き込みで画面を空にしない（何が張られているかは変わっていない）。
    expect(result.current.rows).toHaveLength(1);
  });

  it('速く開き直したとき、古い応答で新しい結果を上書きしない', async () => {
    // 1 本目をわざと遅らせ、2 本目のページの結果が出たあとに着地させる。
    let settleFirst: (v: unknown) => void = () => {};
    hoisted.listPageGrants.mockImplementationOnce(
      () => new Promise((resolve) => { settleFirst = resolve; }),
    );

    const { result, rerender } = renderHook(({ page }) => useNoteShare(SLUG, page), {
      initialProps: { page: 'p-old' },
    });

    hoisted.listPageGrants.mockResolvedValue([grant('pr-dev')]);
    rerender({ page: 'p-new' });
    await waitFor(() => expect(result.current.rows).toHaveLength(1));
    expect(result.current.rows[0].principalId).toBe('pr-dev');

    // 遅れて着地した古い応答は捨てる。
    await act(async () => {
      settleFirst([grant('pr-tanaka'), grant('pr-dev')]);
    });
    expect(result.current.rows.map((r) => r.principalId)).toEqual(['pr-dev']);
  });
});
