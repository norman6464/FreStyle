import apiClient from '../lib/axios';
import { WeeklyChallenge } from '../types';

export const WeeklyChallengeRepository = {
  async fetchCurrentChallenge(): Promise<WeeklyChallenge | null> {
    try {
      const response = await apiClient.get<WeeklyChallenge>('/api/weekly-challenge');
      return response.data;
    } catch {
      return null;
    }
  },

  async incrementProgress(): Promise<WeeklyChallenge | null> {
    try {
      const response = await apiClient.post<WeeklyChallenge>('/api/weekly-challenge/progress');
      return response.data;
    } catch {
      return null;
    }
  },
};
