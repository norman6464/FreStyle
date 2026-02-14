import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useChatList } from '../useChatList';
import ChatRepository from '../../repositories/ChatRepository';

vi.mock('../../repositories/ChatRepository');

const mockedRepo = vi.mocked(ChatRepository);

describe('useChatList', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedRepo.fetchChatUsers.mockResolvedValue([]);
    mockedRepo.fetchCurrentUser.mockResolvedValue({ id: 1, name: 'テスト' });
  });

  it('初期ロード時にチャットユーザーを取得する', async () => {
    const mockUsers = [{ roomId: 1, userId: 1, name: 'ユーザー1', unreadCount: 0 }];
    mockedRepo.fetchChatUsers.mockResolvedValue(mockUsers as any);

    const { result } = renderHook(() => useChatList());

    await waitFor(() => {
      expect(result.current.chatUsers).toEqual(mockUsers);
    });
  });

  it('ユーザーIDを取得する', async () => {
    const { result } = renderHook(() => useChatList());

    await waitFor(() => {
      expect(result.current.userId).toBe(1);
    });
  });

  it('ローディング状態を管理する', async () => {
    const { result } = renderHook(() => useChatList());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('updateUnreadCountで未読数が更新される', async () => {
    const mockUsers = [
      { roomId: 1, userId: 1, name: 'ユーザー1', unreadCount: 2 },
      { roomId: 2, userId: 2, name: 'ユーザー2', unreadCount: 0 },
    ];
    mockedRepo.fetchChatUsers.mockResolvedValue(mockUsers as any);

    const { result } = renderHook(() => useChatList());

    await waitFor(() => {
      expect(result.current.chatUsers).toHaveLength(2);
    });

    act(() => {
      result.current.updateUnreadCount(1, 3);
    });

    expect(result.current.chatUsers[0].unreadCount).toBe(5);
    expect(result.current.chatUsers[1].unreadCount).toBe(0);
  });

  it('fetchChatUsers失敗時にエラーをサイレントに処理する', async () => {
    mockedRepo.fetchChatUsers.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useChatList());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.chatUsers).toEqual([]);
  });

  it('fetchUserId失敗時にuserIdがnullのまま', async () => {
    mockedRepo.fetchCurrentUser.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useChatList());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.userId).toBeNull();
  });

  it('fetchChatUsersにクエリパラメータを渡せる', async () => {
    const { result } = renderHook(() => useChatList());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.fetchChatUsers('検索ワード');
    });

    expect(mockedRepo.fetchChatUsers).toHaveBeenCalledWith('検索ワード');
  });
});
