import { useState, useEffect, useCallback, useRef } from 'react';
import SockJS from 'sockjs-client';
import { Client } from '@stomp/stompjs';
import ChatRepository from '../repositories/ChatRepository';
import { ChatUser } from '../types';

/**
 * チャットリスト管理フック
 *
 * ChatListPageからビジネスロジックを分離し、
 * Repository経由でAPI呼び出しを行う。
 * WebSocket購読・デバウンス検索も含む。
 */
export function useChatList() {
  const [chatUsers, setChatUsers] = useState<ChatUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [userId, setUserId] = useState<number | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const stompClientRef = useRef<Client | null>(null);
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

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

  // リアルタイム未読数更新のWebSocket購読
  useEffect(() => {
    if (!userId) return;

    const client = new Client({
      webSocketFactory: () =>
        new SockJS(`${API_BASE_URL}/ws/chat`, undefined, { withCredentials: true }),
      reconnectDelay: 5000,
      onConnect: () => {
        client.subscribe(`/topic/unread/${userId}`, (message) => {
          const data = JSON.parse(message.body);
          if (data.type === 'unread_update') {
            updateUnreadCount(data.roomId, data.increment);
          }
        });
      },
    });

    stompClientRef.current = client;
    client.activate();
    return () => { client.deactivate(); };
  }, [userId, updateUnreadCount, API_BASE_URL]);

  // 検索クエリ変更時にデバウンス検索
  useEffect(() => {
    const timer = setTimeout(() => {
      fetchChatUsers(searchQuery || undefined);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery, fetchChatUsers]);

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
