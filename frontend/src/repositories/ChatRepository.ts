import apiClient from '../lib/axios';
import { ChatUser } from '../types';

/**
 * Chatリポジトリ
 *
 * チャット関連のAPI呼び出しを抽象化し、
 * fetch()直接呼び出しを排除してaxiosインターセプターによる
 * 自動トークンリフレッシュを活用する。
 */
const ChatRepository = {
  async fetchChatUsers(query?: string): Promise<ChatUser[]> {
    const params: Record<string, string> = {};
    if (query) params.query = query;
    const res = await apiClient.get('/api/chat/rooms', { params });
    return res.data.chatUsers || [];
  },

  async fetchCurrentUser(): Promise<{ id: number; name: string }> {
    const res = await apiClient.get('/api/auth/cognito/me');
    return res.data;
  },

  async createRoom(userId: number): Promise<{ roomId: number }> {
    const res = await apiClient.post(`/api/chat/users/${userId}/create`);
    return res.data;
  },
};

export default ChatRepository;
