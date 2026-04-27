import apiClient from '../lib/axios';
import type { LearningReport } from '../types';

export const LearningReportRepository = {
  async getAll(): Promise<LearningReport[]> {
    const response = await apiClient.get<LearningReport[]>('/api/v2/reports');
    return response.data;
  },

  async getMonthly(year: number, month: number): Promise<LearningReport | null> {
    try {
      const response = await apiClient.get<LearningReport>(`/api/v2/reports/${year}/${month}`);
      return response.data;
    } catch {
      return null;
    }
  },

  async generate(year: number, month: number): Promise<{ status: string }> {
    const response = await apiClient.post<{ status: string }>('/api/v2/reports/generate', { year, month });
    return response.data;
  },
};
