import { useState, useEffect, useCallback } from 'react';
import ChatRepository from '../repositories/ChatRepository';
import { ChatUser } from '../types';
import { useDebounce } from './useDebounce';

/**
 * チャットリスト管理フック
 *
 * ChatListPage からビジネスロジックを分離し、Repository 経由で API 呼び出しを行う。
 * SockJS / STOMP ベースの旧 unread リアルタイム購読は廃止し、
 * Page 表示時 fetch + 検索デバウンスのシンプル構成にした。
 * リアルタイム unread 更新は別 issue で raw WebSocket 化する予定。
 */
export function useChatList() {
  const [chatUsers, setChatUsers] = useState<ChatUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [userId, setUserId] = useState<number | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const debouncedSearchQuery = useDebounce(searchQuery, 300);

  const fetchChatUsers = useCallback(async (query?: string) => {
    try {
      setLoading(true);
      const users = await ChatRepository.fetchChatUsers(query);
      setChatUsers(users);
    } catch {
      // サイレントに処理
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchUserId = useCallback(async () => {
    try {
      const user = await ChatRepository.fetchCurrentUser();
      setUserId(user.id);
    } catch {
      // サイレントに処理
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

  // デバウンスされた検索クエリが変わったらフェッチ
  useEffect(() => {
    fetchChatUsers(debouncedSearchQuery || undefined);
  }, [debouncedSearchQuery, fetchChatUsers]);

  return {
    chatUsers,
    loading,
    userId,
    searchQuery,
    setSearchQuery,
    fetchChatUsers,
    updateUnreadCount,
  };
}
