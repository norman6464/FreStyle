import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useProfileStats } from '../useProfileStats';

vi.mock('../../repositories/ProfileStatsRepository', () => ({
  default: {
    fetchStats: vi.fn(),
  },
}));

import ProfileStatsRepository from '../../repositories/ProfileStatsRepository';

const mockFetchStats = ProfileStatsRepository.fetchStats as ReturnType<typeof vi.fn>;

describe('useProfileStats', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('初期状態はloading=trueでstats=null', () => {
    mockFetchStats.mockReturnValue(new Promise(() => {})); // never resolves

    const { result } = renderHook(() => useProfileStats());

    expect(result.current.loading).toBe(true);
    expect(result.current.stats).toBeNull();
  });

  it('統計データを正常に取得する', async () => {
    const mockStats = {
      totalSessions: 10,
      averageScore: 7.5,
      currentStreak: 3,
      longestStreak: 5,
      totalAchievedDays: 15,
      followerCount: 2,
      followingCount: 4,
    };
    mockFetchStats.mockResolvedValue(mockStats);

    const { result } = renderHook(() => useProfileStats());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.stats).toEqual(mockStats);
  });

  it('取得エラー時はstats=nullでloading=false', async () => {
    mockFetchStats.mockRejectedValue(new Error('API error'));

    const { result } = renderHook(() => useProfileStats());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.stats).toBeNull();
  });
});
