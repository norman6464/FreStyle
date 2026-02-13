import { useState, useEffect, useCallback } from 'react';
import ChatRepository from '../repositories/ChatRepository';
import { ChatUser } from '../types';

/**
 * チャットリスト管理フック
 *
 * ChatListPageからビジネスロジックを分離し、
 * Repository経由でAPI呼び出しを行う。
 */
export function useChatList() {
  const [chatUsers, setChatUsers] = useState<ChatUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [userId, setUserId] = useState<number | null>(null);

  const fetchChatUsers = useCallback(async (query?: string) => {
    try {
      setLoading(true);
      const users = await ChatRepository.fetchChatUsers(query);
      setChatUsers(users);
    } catch (e) {
      console.error('チャットユーザー取得失敗', e);
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchUserId = useCallback(async () => {
    try {
      const user = await ChatRepository.fetchCurrentUser();
      setUserId(user.id);
    } catch (e) {
      console.error('ユーザー情報取得エラー:', e);
    }
  }, []);

  const updateUnreadCount = useCallback((roomId: number, increment: number) => {
    setChatUsers((prev) =>
      prev.map((u) =>
        u.roomId === roomId
          ? { ...u, unreadCount: (u.unreadCount || 0) + increment }
          : u
      )
    );
  }, []);

  useEffect(() => {
    fetchUserId();
    fetchChatUsers();
  }, [fetchUserId, fetchChatUsers]);

  return {
    chatUsers,
    loading,
    userId,
    fetchChatUsers,
    updateUnreadCount,
  };
}
