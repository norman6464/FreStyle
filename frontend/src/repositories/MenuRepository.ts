import apiClient from '../lib/axios';
import { CHAT, SCORES } from '../constants/apiRoutes';
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
      const { data } = await apiClient.get<ChatRoomsResponse>(CHAT.rooms);
      const count = Array.isArray(data?.chatUsers) ? data.chatUsers.length : 0;
      return { chatPartnerCount: count };
    } catch {
      return { chatPartnerCount: 0 };
    }
  },

  async fetchChatRooms(): Promise<ChatRoomsResponse> {
    const { data } = await apiClient.get(CHAT.rooms);
    return data;
  },

  async fetchScoreHistory(): Promise<ScoreHistory[]> {
    const { data } = await apiClient.get(SCORES.cards);
    return data;
  },
};
