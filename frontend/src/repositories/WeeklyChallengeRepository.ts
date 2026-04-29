import apiClient from '../lib/axios';
import { WEEKLY_CHALLENGE } from '../constants/apiRoutes';
import { WeeklyChallenge } from '../types';

export const WeeklyChallengeRepository = {
  async fetchCurrentChallenge(): Promise<WeeklyChallenge | null> {
    try {
      const response = await apiClient.get<WeeklyChallenge>(WEEKLY_CHALLENGE.current);
      return response.data;
    } catch {
      return null;
    }
  },

  async incrementProgress(): Promise<WeeklyChallenge | null> {
    try {
      const response = await apiClient.post<WeeklyChallenge>(WEEKLY_CHALLENGE.progress);
      return response.data;
    } catch {
      return null;
    }
  },
};
