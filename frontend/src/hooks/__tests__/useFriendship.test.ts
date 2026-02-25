import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useFriendship } from '../useFriendship';

const mockGetFollowing = vi.fn();
const mockGetFollowers = vi.fn();
const mockFollow = vi.fn();
const mockUnfollow = vi.fn();

vi.mock('../../repositories/FriendshipRepository', () => ({
  FriendshipRepository: {
    getFollowing: (...args: unknown[]) => mockGetFollowing(...args),
    getFollowers: (...args: unknown[]) => mockGetFollowers(...args),
    follow: (...args: unknown[]) => mockFollow(...args),
    unfollow: (...args: unknown[]) => mockUnfollow(...args),
  },
}));

const mockFollowing = [
  { id: 1, userId: 2, username: 'ユーザーB', iconUrl: null, bio: null, mutual: true, createdAt: '2026-02-17T00:00:00' },
  { id: 2, userId: 3, username: 'ユーザーC', iconUrl: null, bio: null, mutual: false, createdAt: '2026-02-16T00:00:00' },
];

const mockFollowers = [
  { id: 3, userId: 4, username: 'フォロワーA', iconUrl: null, bio: null, mutual: false, createdAt: '2026-02-15T00:00:00' },
];

describe('useFriendship', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetFollowing.mockResolvedValue(mockFollowing);
    mockGetFollowers.mockResolvedValue(mockFollowers);
  });

  it('初期状態はloading=trueで空リスト', () => {
    const { result } = renderHook(() => useFriendship());
    expect(result.current.loading).toBe(true);
    expect(result.current.following).toEqual([]);
    expect(result.current.followers).toEqual([]);
  });

  it('マウント時にフォロー/フォロワー一覧を取得する', async () => {
    const { result } = renderHook(() => useFriendship());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.following).toHaveLength(2);
    expect(result.current.followers).toHaveLength(1);
  });

  it('ユーザーをフォローできる', async () => {
    const newFollow = { id: 10, userId: 5, username: '新しいフォロー', iconUrl: null, bio: null, mutual: false, createdAt: '2026-02-17T12:00:00' };
    mockFollow.mockResolvedValue(newFollow);
    mockGetFollowing.mockResolvedValueOnce(mockFollowing).mockResolvedValueOnce([...mockFollowing, newFollow]);
    mockGetFollowers.mockResolvedValue(mockFollowers);

    const { result } = renderHook(() => useFriendship());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.followUser(5);
    });

    expect(mockFollow).toHaveBeenCalledWith(5);
  });

  it('フォローを解除できる', async () => {
    mockUnfollow.mockResolvedValue(undefined);
    mockGetFollowing.mockResolvedValueOnce(mockFollowing).mockResolvedValueOnce([mockFollowing[1]]);
    mockGetFollowers.mockResolvedValue(mockFollowers);

    const { result } = renderHook(() => useFriendship());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.unfollowUser(2);
    });

    expect(mockUnfollow).toHaveBeenCalledWith(2);
  });

  it('API失敗時はerrorが設定され空リストを返す', async () => {
    mockGetFollowing.mockRejectedValue(new Error('API Error'));
    mockGetFollowers.mockRejectedValue(new Error('API Error'));

    const { result } = renderHook(() => useFriendship());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.following).toEqual([]);
    expect(result.current.followers).toEqual([]);
    expect(result.current.error).toBe('フレンド情報の取得に失敗しました');
  });

  it('フォロー失敗時にerrorが設定される', async () => {
    mockFollow.mockRejectedValue(new Error('Follow failed'));

    const { result } = renderHook(() => useFriendship());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.followUser(5);
    });

    expect(result.current.error).toBe('フォローに失敗しました');
  });

  it('フォロー解除失敗時にerrorが設定される', async () => {
    mockUnfollow.mockRejectedValue(new Error('Unfollow failed'));

    const { result } = renderHook(() => useFriendship());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.unfollowUser(2);
    });

    expect(result.current.error).toBe('フォロー解除に失敗しました');
  });
});
