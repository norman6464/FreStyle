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

  it('デバウンス後に検索クエリ付きで検索が実行される', async () => {
    const mockUsers = [{ id: 2, name: '検索結果', email: 'found@example.com' }];
    vi.mocked(UserSearchRepository.searchUsers).mockResolvedValue(mockUsers);

    const { result } = renderHook(() => useUserSearch());

    act(() => {
      result.current.setSearchQuery('検索');
    });

    // デバウンス前はまだ反映されていない
    expect(result.current.debounceQuery).toBe('');

    // デバウンス後に検索が実行される
    await waitFor(() => {
      expect(result.current.debounceQuery).toBe('検索');
    }, { timeout: 2000 });

    await waitFor(() => {
      expect(UserSearchRepository.searchUsers).toHaveBeenCalledWith('検索');
      expect(result.current.users).toEqual(mockUsers);
    });
  });

  it('検索成功時にerrorがクリアされる', async () => {
    // まずエラーを発生させる
    vi.mocked(UserSearchRepository.searchUsers).mockRejectedValueOnce(new Error('エラー'));

    const { result } = renderHook(() => useUserSearch());

    await waitFor(() => {
      expect(result.current.error).toBe('エラー');
    });

    // 次の検索が成功するようにモックを設定
    vi.mocked(UserSearchRepository.searchUsers).mockResolvedValue([{ id: 1, name: 'OK', email: 'ok@example.com' }]);

    // searchQueryを更新してデバウンス後に再検索をトリガー
    act(() => {
      result.current.setSearchQuery('新しい検索');
    });

    await waitFor(() => {
      expect(result.current.error).toBeNull();
    }, { timeout: 2000 });
  });

  it('アンマウント時にキャンセルされた検索結果は反映されない', async () => {
    let resolveSearch: (value: any) => void;
    vi.mocked(UserSearchRepository.searchUsers).mockImplementation(
      () => new Promise((resolve) => { resolveSearch = resolve; })
    );

    const { result, unmount } = renderHook(() => useUserSearch());

    // アンマウント前にプロミスが未解決の状態
    unmount();

    // プロミスを解決してもusersは更新されないことを確認
    await act(async () => {
      resolveSearch!([{ id: 1, name: 'Late Result', email: 'late@example.com' }]);
    });

    // アンマウント後なのでresult.currentにアクセスしてもエラーにならないことを確認
    expect(result.current.users).toEqual([]);
  });
});
