import apiClient from '../lib/axios';

export const ScoreGoalRepository = {
  async fetchGoal(): Promise<number | null> {
    try {
      const response = await apiClient.get<{ goalScore: number }>('/api/v2/score-goals');
      return response.data?.goalScore ?? null;
    } catch {
      return null;
    }
  },

  async saveGoal(goalScore: number): Promise<void> {
    try {
      await apiClient.put('/api/v2/score-goals', { goalScore });
    } catch {
      // API失敗時は呼び出し元でlocalStorage保存済み
    }
  },
};
