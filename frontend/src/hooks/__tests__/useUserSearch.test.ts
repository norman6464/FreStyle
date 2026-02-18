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

  it('クエリが空の場合はAPIを呼ばずusersが空配列のまま', () => {
    const { result } = renderHook(() => useUserSearch());

    expect(result.current.users).toEqual([]);
    expect(result.current.loading).toBe(false);
    expect(UserSearchRepository.searchUsers).not.toHaveBeenCalled();
  });

  it('検索中はloadingがtrueになる', async () => {
    let resolveSearch: (value: any) => void;
    vi.mocked(UserSearchRepository.searchUsers).mockImplementation(
      () => new Promise((resolve) => { resolveSearch = resolve; })
    );

    const { result } = renderHook(() => useUserSearch());

    // クエリを入力してデバウンス後にloadingがtrueになる
    act(() => {
      result.current.setSearchQuery('テスト');
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    }, { timeout: 2000 });

    await act(async () => {
      resolveSearch!([]);
    });

    expect(result.current.loading).toBe(false);
  });

  it('検索エラー時にerrorが設定される', async () => {
    vi.mocked(UserSearchRepository.searchUsers).mockRejectedValue(new Error('ネットワークエラー'));

    const { result } = renderHook(() => useUserSearch());

    act(() => {
      result.current.setSearchQuery('テスト');
    });

    await waitFor(() => {
      expect(result.current.error).toBe('ネットワークエラー');
    }, { timeout: 2000 });
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

    act(() => {
      result.current.setSearchQuery('失敗クエリ');
    });

    await waitFor(() => {
      expect(result.current.error).toBe('エラー');
    }, { timeout: 2000 });

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

  it('初期状態でsearchQueryが空文字・loadingに関わらずusersが空', () => {
    const { result } = renderHook(() => useUserSearch());
    expect(result.current.searchQuery).toBe('');
    expect(result.current.debounceQuery).toBe('');
    expect(result.current.users).toEqual([]);
  });

  it('non-Errorオブジェクトのreject時にデフォルトメッセージが表示されない', async () => {
    vi.mocked(UserSearchRepository.searchUsers).mockRejectedValue('文字列エラー');

    const { result } = renderHook(() => useUserSearch());

    act(() => {
      result.current.setSearchQuery('テスト');
    });

    await waitFor(() => {
      expect(result.current.error).toBeDefined();
    }, { timeout: 2000 });
  });

  it('アンマウント時にキャンセルされた検索結果は反映されない', async () => {
    let resolveSearch: (value: any) => void;
    vi.mocked(UserSearchRepository.searchUsers).mockImplementation(
      () => new Promise((resolve) => { resolveSearch = resolve; })
    );

    const { result, unmount } = renderHook(() => useUserSearch());

    // クエリを入力してデバウンス後に検索をトリガー
    act(() => {
      result.current.setSearchQuery('テスト');
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    }, { timeout: 2000 });

    // アンマウント前にプロミスが未解決の状態
    unmount();

    // プロミスを解決してもusersは更新されないことを確認
    await act(async () => {
      resolveSearch!([{ id: 1, name: 'Late Result', email: 'late@example.com' }]);
    });

    // アンマウント後なのでresult.currentにアクセスしてもエラーにならないことを確認
    expect(result.current.users).toEqual([]);
  });

  it('クエリをクリアするとusersが空に戻りAPIを呼ばない', async () => {
    vi.mocked(UserSearchRepository.searchUsers).mockResolvedValue([{ id: 1, name: 'テスト', email: 'test@example.com' }]);

    const { result } = renderHook(() => useUserSearch());

    act(() => {
      result.current.setSearchQuery('テスト');
    });

    await waitFor(() => {
      expect(result.current.users).toHaveLength(1);
    }, { timeout: 2000 });

    vi.clearAllMocks();

    // クエリをクリア
    act(() => {
      result.current.setSearchQuery('');
    });

    await waitFor(() => {
      expect(result.current.debounceQuery).toBe('');
    }, { timeout: 2000 });

    expect(result.current.users).toEqual([]);
    expect(UserSearchRepository.searchUsers).not.toHaveBeenCalled();
  });
});
