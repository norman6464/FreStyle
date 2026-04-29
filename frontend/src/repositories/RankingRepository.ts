import apiClient from '../lib/axios';
import { RANKING } from '../constants/apiRoutes';
import { Ranking } from '../types';

export const RankingRepository = {
  async fetchRanking(period: string = 'weekly'): Promise<Ranking> {
    const response = await apiClient.get<Ranking>(RANKING, {
      params: { period },
    });
    return response.data;
  },
};
