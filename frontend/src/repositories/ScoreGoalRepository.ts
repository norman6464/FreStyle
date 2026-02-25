import apiClient from '../lib/axios';

export const ScoreGoalRepository = {
  async fetchGoal(): Promise<number | null> {
    try {
      const response = await apiClient.get<{ goalScore: number }>('/api/score-goal');
      return response.data?.goalScore ?? null;
    } catch {
      return null;
    }
  },

  async saveGoal(goalScore: number): Promise<void> {
    try {
      await apiClient.put('/api/score-goal', { goalScore });
    } catch {
      // API失敗時は呼び出し元でlocalStorage保存済み
    }
  },
};
