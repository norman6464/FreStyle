import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useUserSearch } from '../useUserSearch';
import UserSearchRepository from '../../repositories/UserSearchRepository';

vi.mock('../../repositories/UserSearchRepository');

describe('useUserSearch', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(UserSearchRepository.searchUsers).mockResolvedValue([]);
  });

  it('初期状態ではusersが空配列である', () => {
    const { result } = renderHook(() => useUserSearch());

    expect(result.current.users).toEqual([]);
    expect(result.current.error).toBeNull();
  });

  it('初回マウント時にユーザー検索が実行される', async () => {
    const mockUsers = [{ id: 1, name: 'テスト', email: 'test@example.com' }];
    vi.mocked(UserSearchRepository.searchUsers).mockResolvedValue(mockUsers);

    const { result } = renderHook(() => useUserSearch());

    await waitFor(() => {
      expect(result.current.users).toEqual(mockUsers);
    });

    expect(UserSearchRepository.searchUsers).toHaveBeenCalled();
  });

  it('検索エラー時にerrorが設定される', async () => {
    vi.mocked(UserSearchRepository.searchUsers).mockRejectedValue(new Error('ネットワークエラー'));

    const { result } = renderHook(() => useUserSearch());

    await waitFor(() => {
      expect(result.current.error).toBe('ネットワークエラー');
    });
  });

  it('setSearchQueryで検索クエリを更新できる', () => {
    const { result } = renderHook(() => useUserSearch());

    act(() => {
      result.current.setSearchQuery('テスト');
    });

    expect(result.current.searchQuery).toBe('テスト');
  });
});
