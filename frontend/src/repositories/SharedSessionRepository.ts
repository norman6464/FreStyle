import apiClient from '../lib/axios';
import { SharedSession } from '../types';

export const SharedSessionRepository = {
  async fetchPublicSessions(): Promise<SharedSession[]> {
    const response = await apiClient.get<SharedSession[]>('/api/v2/shared-sessions');
    return response.data;
  },

  async shareSession(sessionId: number, description?: string): Promise<SharedSession> {
    const response = await apiClient.post<SharedSession>('/api/v2/shared-sessions', {
      sessionId,
      description,
    });
    return response.data;
  },

  async unshareSession(sessionId: number): Promise<void> {
    await apiClient.delete(`/api/v2/shared-sessions/${sessionId}`);
  },
};
