import apiClient from '../lib/axios';
import { Ranking } from '../types';

export const RankingRepository = {
  async fetchRanking(period: string = 'weekly'): Promise<Ranking> {
    const response = await apiClient.get<Ranking>('/api/ranking', {
      params: { period },
    });
    return response.data;
  },
};
