import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
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

    // 最終的にloadingはfalseになる
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });
});
