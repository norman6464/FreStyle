import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../lib/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

import apiClient from '../../lib/axios';
import { LearningReportRepository } from '../LearningReportRepository';

const mockedGet = vi.mocked(apiClient.get);
const mockedPost = vi.mocked(apiClient.post);

describe('LearningReportRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getAll: レポート一覧を取得する', async () => {
    const reports = [
      { id: 1, year: 2024, month: 6, totalSessions: 10, averageScore: 7.5, practiceDays: 8 },
    ];
    mockedGet.mockResolvedValue({ data: reports });

    const result = await LearningReportRepository.getAll();
    expect(result).toEqual(reports);
    expect(mockedGet).toHaveBeenCalledWith('/api/reports');
  });

  it('getMonthly: 指定月のレポートを取得する', async () => {
    const report = { id: 1, year: 2024, month: 6, totalSessions: 10, averageScore: 7.5, practiceDays: 8 };
    mockedGet.mockResolvedValue({ data: report });

    const result = await LearningReportRepository.getMonthly(2024, 6);
    expect(result).toEqual(report);
    expect(mockedGet).toHaveBeenCalledWith('/api/reports/2024/6');
  });

  it('getMonthly: 該当レポートがない場合はnullを返す', async () => {
    mockedGet.mockRejectedValue(new Error('Not Found'));

    const result = await LearningReportRepository.getMonthly(2024, 12);
    expect(result).toBeNull();
  });

  it('generate: レポート生成リクエストを送る', async () => {
    mockedPost.mockResolvedValue({ data: { status: 'generated' } });

    const result = await LearningReportRepository.generate(2024, 6);
    expect(result).toEqual({ status: 'generated' });
    expect(mockedPost).toHaveBeenCalledWith('/api/reports/generate', { year: 2024, month: 6 });
  });
});
