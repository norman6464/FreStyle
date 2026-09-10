import { renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useWorkspaceMembers } from '../useWorkspaceMembers';

const hoisted = vi.hoisted(() => ({
  fetchMembers: vi.fn(),
}));

vi.mock('@/entities/kb', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/kb')>();
  return {
    ...actual,
    KbRepository: { fetchMembers: hoisted.fetchMembers },
  };
});

const SLUG = 'acme';

beforeEach(() => {
  vi.clearAllMocks();
});

describe('useWorkspaceMembers', () => {
  it('宛先が揃ったら取得する', async () => {
    hoisted.fetchMembers.mockResolvedValue([{ principalId: 'p-1', userId: 1, name: '田中 太郎' }]);
    const { result } = renderHook(() => useWorkspaceMembers(SLUG));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.members).toHaveLength(1);
    expect(hoisted.fetchMembers).toHaveBeenCalledWith(SLUG);
  });

  it('取得失敗は文言を出す', async () => {
    hoisted.fetchMembers.mockRejectedValue(new Error('network'));
    const { result } = renderHook(() => useWorkspaceMembers(SLUG));
    await waitFor(() => expect(result.current.error).not.toBeNull());
  });

  it('宛先が揃っていなければ何もしない', () => {
    const { result } = renderHook(() => useWorkspaceMembers(undefined));
    expect(result.current.members).toEqual([]);
    expect(hoisted.fetchMembers).not.toHaveBeenCalled();
  });
});
