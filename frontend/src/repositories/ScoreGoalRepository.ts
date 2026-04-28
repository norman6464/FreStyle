import apiClient from '../lib/axios';

// Go バックエンドの ScoreGoalHandler は { targetScore: number } を bind/return する。
// 旧 Spring Boot 時代の goalScore は廃止済み。
export const ScoreGoalRepository = {
  async fetchGoal(): Promise<number | null> {
    try {
      const response = await apiClient.get<{ targetScore: number }>('/api/v2/score-goals');
      return response.data?.targetScore ?? null;
    } catch {
      return null;
    }
  },

  async saveGoal(targetScore: number): Promise<void> {
    try {
      await apiClient.put('/api/v2/score-goals', { targetScore });
    } catch {
      // API失敗時は呼び出し元でlocalStorage保存済み
    }
  },
};
