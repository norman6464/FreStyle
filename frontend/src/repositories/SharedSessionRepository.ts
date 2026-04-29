import apiClient from '../lib/axios';
import { SHARED_SESSIONS } from '../constants/apiRoutes';
import { SharedSession } from '../types';

export const SharedSessionRepository = {
  async fetchPublicSessions(): Promise<SharedSession[]> {
    const response = await apiClient.get<SharedSession[]>(SHARED_SESSIONS.list);
    return response.data;
  },

  async shareSession(sessionId: number, description?: string): Promise<SharedSession> {
    const response = await apiClient.post<SharedSession>(SHARED_SESSIONS.list, {
      sessionId,
      description,
    });
    return response.data;
  },

  async unshareSession(sessionId: number): Promise<void> {
    await apiClient.delete(SHARED_SESSIONS.byId(sessionId));
  },
};
