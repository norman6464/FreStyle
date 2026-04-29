import apiClient from '../lib/axios';
import { LEARNING_REPORTS } from '../constants/apiRoutes';
import type { LearningReport } from '../types';

// Go バックエンドの正規パスは /learning-reports（旧 Spring Boot 時代の /reports は廃止）。
export const LearningReportRepository = {
  async getAll(): Promise<LearningReport[]> {
    const response = await apiClient.get<LearningReport[]>(LEARNING_REPORTS.list);
    return response.data;
  },

  async getMonthly(year: number, month: number): Promise<LearningReport | null> {
    try {
      const response = await apiClient.get<LearningReport>(LEARNING_REPORTS.yearMonth(year, month));
      return response.data;
    } catch {
      return null;
    }
  },

  async generate(year: number, month: number): Promise<{ status: string }> {
    const response = await apiClient.post<{ status: string }>(LEARNING_REPORTS.generate, { year, month });
    return response.data;
  },
};
