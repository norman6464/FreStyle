import apiClient from '../lib/axios';
import { SCORES } from '../constants/apiRoutes';
import type { ScoreHistory } from '../types';

export const MenuRepository = {
  async fetchScoreHistory(): Promise<ScoreHistory[]> {
    const { data } = await apiClient.get(SCORES.cards);
    return data;
  },
};
