import apiClient from '../lib/axios';
import type { ScoreHistory } from '../types';

interface ChatStats {
  chatPartnerCount: number;
}

interface ChatRoomsResponse {
  chatUsers: Array<{ roomId: number; unreadCount: number }>;
}

export const MenuRepository = {
  async fetchChatStats(): Promise<ChatStats> {
    const { data } = await apiClient.get('/api/v2/chat/stats');
    return data;
  },

  async fetchChatRooms(): Promise<ChatRoomsResponse> {
    const { data } = await apiClient.get('/api/v2/chat/rooms');
    return data;
  },

  async fetchScoreHistory(): Promise<ScoreHistory[]> {
    const { data } = await apiClient.get('/api/v2/scores/history');
    return data;
  },
};
