import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useLearningReport } from '../useLearningReport';

const mockGetAll = vi.fn();
const mockGetMonthly = vi.fn();
const mockGenerate = vi.fn();

vi.mock('@/entities/learning-report/api/learningReportRepository', () => ({
  LearningReportRepository: {
    getAll: (...args: unknown[]) => mockGetAll(...args),
    getMonthly: (...args: unknown[]) => mockGetMonthly(...args),
    generate: (...args: unknown[]) => mockGenerate(...args),
  },
}));

const mockReports = [
  { id: 1, userId: 7, periodFrom: '2026-02-01T00:00:00Z', periodTo: '2026-03-01T00:00:00Z', status: 'pending', createdAt: '2026-02-28T00:00:00Z' },
  { id: 2, userId: 7, periodFrom: '2026-01-01T00:00:00Z', periodTo: '2026-02-01T00:00:00Z', status: 'ready', s3Key: 'reports/2026-01.pdf', createdAt: '2026-01-31T00:00:00Z' },
];

describe('useLearningReport', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetAll.mockResolvedValue(mockReports);
    mockGenerate.mockResolvedValue({ status: 'processing' });
  });

  it('初期状態はloading=trueで空のレポートリスト', () => {
    const { result } = renderHook(() => useLearningReport());
    expect(result.current.loading).toBe(true);
    expect(result.current.reports).toEqual([]);
    expect(result.current.generating).toBe(false);
  });

  it('マウント時にレポート一覧を取得する', async () => {
    const { result } = renderHook(() => useLearningReport());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.reports).toHaveLength(2);
    expect(mockGetAll).toHaveBeenCalled();
  });

  it('generateReportで202レスポンス後にgenerating=trueになる', async () => {
    const { result } = renderHook(() => useLearningReport());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.generateReport(2026, 2);
    });

    expect(result.current.generating).toBe(true);
    expect(mockGenerate).toHaveBeenCalledWith(2026, 2);
  });

  it('ポーリングでレポートが増えたらgenerating=falseになる', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });

    const { result } = renderHook(() => useLearningReport());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.generateReport(2026, 3);
    });

    expect(result.current.generating).toBe(true);

    // ポーリングでレポートが増える
    const newReport = { id: 3, year: 2026, month: 3, totalSessions: 3, averageScore: 70.0, previousAverageScore: 80.0, scoreChange: -10.0, bestAxis: '論理的構成力', worstAxis: '配慮表現', practiceDays: 2, createdAt: '2026-04-01T00:00:00' };
    mockGetAll.mockResolvedValue([...mockReports, newReport]);

    await act(async () => {
      vi.advanceTimersByTime(5000);
    });

    await waitFor(() => {
      expect(result.current.generating).toBe(false);
    });

    expect(result.current.reports).toHaveLength(3);

    vi.useRealTimers();
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
