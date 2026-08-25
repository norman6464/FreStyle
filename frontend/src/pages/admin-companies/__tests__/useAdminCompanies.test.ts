import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';

const { listStats, updateActive } = vi.hoisted(() => ({
  listStats: vi.fn(),
  updateActive: vi.fn(),
}));
vi.mock('@/entities/company', () => ({
  CompanyRepository: {
    listStats: () => listStats(),
    updateActive: (id: number, active: boolean) => updateActive(id, active),
  },
}));
vi.mock('@/shared/lib/logger', () => ({ logger: { error: vi.fn(), warn: vi.fn(), info: vi.fn() } }));

import { useAdminCompanies } from '../model/useAdminCompanies';

const acme = {
  id: 1,
  name: 'アクメ社',
  isActive: true,
  createdAt: '2026-06-01T00:00:00Z',
  memberTotal: 5,
  activeMembers: 4,
  traineeCount: 3,
};

beforeEach(() => {
  vi.clearAllMocks();
  listStats.mockResolvedValue([acme]);
  updateActive.mockResolvedValue(undefined);
});

async function mountHook() {
  const view = renderHook(() => useAdminCompanies());
  await waitFor(() => expect(view.result.current.loading).toBe(false));
  return view;
}

describe('useAdminCompanies', () => {
  it('マウント時に横断ビューを 1 回だけ取得する', async () => {
    const { result } = await mountHook();

    expect(listStats).toHaveBeenCalledTimes(1);
    expect(result.current.companies).toEqual([acme]);
    expect(result.current.error).toBeNull();
    expect(result.current.updatingId).toBeNull();
  });

  it('取得に失敗したときはエラーを立てて loading を落とす', async () => {
    listStats.mockRejectedValue(new Error('boom'));
    const { result } = await mountHook();

    expect(result.current.error).toBe('会社一覧の取得に失敗しました');
    expect(result.current.companies).toEqual([]);
  });

  it('切り替えは楽観的に反映し、一覧の再取得はしない', async () => {
    const { result } = await mountHook();

    await act(async () => {
      await result.current.setActive(1, false);
    });

    expect(updateActive).toHaveBeenCalledTimes(1);
    expect(updateActive).toHaveBeenCalledWith(1, false);
    expect(result.current.companies[0].isActive).toBe(false);
    expect(result.current.updatingId).toBeNull();
    expect(listStats).toHaveBeenCalledTimes(1);
  });

  it('切り替えに失敗したときは元の状態へ戻してエラーを出す', async () => {
    updateActive.mockRejectedValue(new Error('boom'));
    const { result } = await mountHook();

    await act(async () => {
      await result.current.setActive(1, false);
    });

    expect(result.current.companies[0].isActive).toBe(true);
    expect(result.current.error).toBe('会社状態の更新に失敗しました');
    expect(result.current.updatingId).toBeNull();
  });

  it('通信中は updatingId に対象の会社を入れる', async () => {
    let release: () => void = () => {};
    updateActive.mockReturnValue(new Promise<void>((resolve) => { release = resolve; }));
    const { result } = await mountHook();

    act(() => { result.current.setActive(1, false); });
    await waitFor(() => expect(result.current.updatingId).toBe(1));

    await act(async () => { release(); });
    await waitFor(() => expect(result.current.updatingId).toBeNull());
  });
});
