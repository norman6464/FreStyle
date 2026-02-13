import apiClient from '../lib/axios';

interface ChatStats {
  chatPartnerCount: number;
}

interface ChatRoomsResponse {
  chatUsers: Array<{ roomId: number; unreadCount: number }>;
}

interface ScoreHistory {
  sessionId: number;
  sessionTitle: string;
  overallScore: number;
  createdAt: string;
}

export const MenuRepository = {
  async fetchChatStats(): Promise<ChatStats> {
    const { data } = await apiClient.get('/api/chat/stats');
    return data;
  },

  async fetchChatRooms(): Promise<ChatRoomsResponse> {
    const { data } = await apiClient.get('/api/chat/rooms');
    return data;
  },

  async fetchScoreHistory(): Promise<ScoreHistory[]> {
    const { data } = await apiClient.get('/api/scores/history');
    return data;
  },
};
