import { describe, it, expect, vi, afterEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useMenuData } from '../useMenuData';
import { MenuRepository } from '../../repositories/MenuRepository';

vi.mock('../../repositories/MenuRepository');

const mockedRepo = vi.mocked(MenuRepository);

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

describe('useMenuData', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('スコアデータを取得する', async () => {
    mockedRepo.fetchScoreHistory.mockResolvedValue([
      { sessionId: 1, sessionTitle: 'テスト', overallScore: 7.5, createdAt: '2026-02-13T00:00:00' },
      { sessionId: 2, sessionTitle: 'テスト2', overallScore: 8.5, createdAt: '2026-02-12T00:00:00' },
    ]);

    const { result } = renderHook(() => useMenuData());

    await waitFor(() => {
      expect(result.current.latestScore).not.toBeNull();
    });

    expect(result.current.latestScore?.overallScore).toBe(7.5);
    expect(result.current.totalSessions).toBe(2);
    expect(result.current.averageScore).toBe(8.0);
    expect(result.current.uniqueDays).toBe(2);
  });

  it('APIエラー時にクラッシュしない', async () => {
    mockedRepo.fetchScoreHistory.mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useMenuData());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.latestScore).toBeNull();
  });

  it('スコアがない場合の初期値が正しい', async () => {
    mockedRepo.fetchScoreHistory.mockResolvedValue([]);

    const { result } = renderHook(() => useMenuData());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.latestScore).toBeNull();
    expect(result.current.allScores).toEqual([]);
    expect(result.current.totalSessions).toBe(0);
    expect(result.current.averageScore).toBe(0);
    expect(result.current.uniqueDays).toBe(0);
  });

  it('同日のセッションはuniqueDays=1になる', async () => {
    mockedRepo.fetchScoreHistory.mockResolvedValue([
      { sessionId: 1, sessionTitle: 'A', overallScore: 7.0, createdAt: '2026-02-13T10:00:00' },
      { sessionId: 2, sessionTitle: 'B', overallScore: 9.0, createdAt: '2026-02-13T14:00:00' },
    ]);

    const { result } = renderHook(() => useMenuData());

    await waitFor(() => {
      expect(result.current.totalSessions).toBe(2);
    });

    expect(result.current.uniqueDays).toBe(1);
    expect(result.current.averageScore).toBe(8.0);
  });

  it('sessionsThisWeekが今週のセッション数を正しく算出する', async () => {
    const now = new Date();
    const day = now.getDay();
    const monday = new Date(now);
    monday.setDate(now.getDate() - (day === 0 ? 6 : day - 1));
    monday.setHours(0, 0, 0, 0);

    const thisWeekDate = new Date(monday);
    thisWeekDate.setDate(monday.getDate() + 1);

    const lastWeekDate = new Date(monday);
    lastWeekDate.setDate(monday.getDate() - 3);

    mockedRepo.fetchScoreHistory.mockResolvedValue([
      { sessionId: 1, sessionTitle: 'A', overallScore: 7.0, createdAt: thisWeekDate.toISOString() },
      { sessionId: 2, sessionTitle: 'B', overallScore: 8.0, createdAt: thisWeekDate.toISOString() },
      { sessionId: 3, sessionTitle: 'C', overallScore: 6.0, createdAt: lastWeekDate.toISOString() },
    ]);

    const { result } = renderHook(() => useMenuData());

    await waitFor(() => {
      expect(result.current.totalSessions).toBe(3);
    });

    expect(result.current.sessionsThisWeek).toBe(2);
  });

  it('practiceDatesが重複なしの日付配列を返す', async () => {
    mockedRepo.fetchScoreHistory.mockResolvedValue([
      { sessionId: 1, sessionTitle: 'A', overallScore: 7.0, createdAt: '2026-02-13T10:00:00' },
      { sessionId: 2, sessionTitle: 'B', overallScore: 8.0, createdAt: '2026-02-13T14:00:00' },
      { sessionId: 3, sessionTitle: 'C', overallScore: 6.0, createdAt: '2026-02-12T09:00:00' },
    ]);

    const { result } = renderHook(() => useMenuData());

    await waitFor(() => {
      expect(result.current.totalSessions).toBe(3);
    });

    expect(result.current.practiceDates).toHaveLength(2);
    expect(result.current.practiceDates).toContain('2026-02-13');
    expect(result.current.practiceDates).toContain('2026-02-12');
  });
});
