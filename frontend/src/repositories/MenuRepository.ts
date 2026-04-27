import apiClient from '../lib/axios';
import type { ScoreHistory } from '../types';

interface ChatStats {
  chatPartnerCount: number;
}

interface ChatRoomsResponse {
  chatUsers: Array<{ roomId: number; unreadCount: number }>;
}

export const MenuRepository = {
  // /chat/stats は Go バックエンドに未実装。chat/rooms から件数を計算する暫定実装。
  async fetchChatStats(): Promise<ChatStats> {
    try {
      const { data } = await apiClient.get<ChatRoomsResponse>('/api/v2/chat/rooms');
      const count = Array.isArray(data?.chatUsers) ? data.chatUsers.length : 0;
      return { chatPartnerCount: count };
    } catch {
      return { chatPartnerCount: 0 };
    }
  },

  async fetchChatRooms(): Promise<ChatRoomsResponse> {
    const { data } = await apiClient.get('/api/v2/chat/rooms');
    return data;
  },

  async fetchScoreHistory(): Promise<ScoreHistory[]> {
    const { data } = await apiClient.get('/api/v2/score-cards');
    return data;
  },
};
