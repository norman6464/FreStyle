import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useLearningReport } from '../useLearningReport';

const mockGetAll = vi.fn();
const mockGetMonthly = vi.fn();
const mockGenerate = vi.fn();

vi.mock('../../repositories/LearningReportRepository', () => ({
  LearningReportRepository: {
    getAll: (...args: unknown[]) => mockGetAll(...args),
    getMonthly: (...args: unknown[]) => mockGetMonthly(...args),
    generate: (...args: unknown[]) => mockGenerate(...args),
  },
}));

const mockReports = [
  { id: 1, year: 2026, month: 2, totalSessions: 8, averageScore: 80.0, previousAverageScore: 75.0, scoreChange: 5.0, bestAxis: '論理的構成力', worstAxis: '配慮表現', practiceDays: 5, createdAt: '2026-03-01T00:00:00' },
  { id: 2, year: 2026, month: 1, totalSessions: 5, averageScore: 75.0, previousAverageScore: null, scoreChange: null, bestAxis: '要約力', worstAxis: '提案力', practiceDays: 3, createdAt: '2026-02-01T00:00:00' },
];

describe('useLearningReport', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetAll.mockResolvedValue(mockReports);
    mockGenerate.mockResolvedValue(mockReports[0]);
  });

  it('初期状態はloading=trueで空のレポートリスト', () => {
    const { result } = renderHook(() => useLearningReport());
    expect(result.current.loading).toBe(true);
    expect(result.current.reports).toEqual([]);
  });

  it('マウント時にレポート一覧を取得する', async () => {
    const { result } = renderHook(() => useLearningReport());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.reports).toHaveLength(2);
    expect(mockGetAll).toHaveBeenCalled();
  });

  it('レポートを生成できる', async () => {
    mockGetAll
      .mockResolvedValueOnce(mockReports)
      .mockResolvedValueOnce(mockReports);

    const { result } = renderHook(() => useLearningReport());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.generateReport(2026, 2);
    });

    expect(mockGenerate).toHaveBeenCalledWith(2026, 2);
  });

  it('API失敗時はエラーなく空リストを返す', async () => {
    mockGetAll.mockRejectedValue(new Error('API Error'));

    const { result } = renderHook(() => useLearningReport());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.reports).toEqual([]);
  });
});
