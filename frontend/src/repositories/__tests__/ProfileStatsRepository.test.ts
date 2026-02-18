import { describe, it, expect, vi, beforeEach } from 'vitest';
import ProfileStatsRepository from '../ProfileStatsRepository';

const mockGet = vi.fn();
vi.mock('../../lib/axios', () => ({
  default: { get: (...args: unknown[]) => mockGet(...args) },
}));

describe('ProfileStatsRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchStatsが両APIを呼び出して結合結果を返す', async () => {
    mockGet
      .mockResolvedValueOnce({
        data: { totalSessions: 10, averageScore: 75.5, followerCount: 3, followingCount: 5 },
      })
      .mockResolvedValueOnce({
        data: { currentStreak: 7, longestStreak: 14, totalAchievedDays: 30 },
      });

    const stats = await ProfileStatsRepository.fetchStats();

    expect(mockGet).toHaveBeenCalledWith('/api/users/me/stats');
    expect(mockGet).toHaveBeenCalledWith('/api/daily-goals/streak');
    expect(stats).toEqual({
      totalSessions: 10,
      averageScore: 75.5,
      currentStreak: 7,
      longestStreak: 14,
      totalAchievedDays: 30,
      followerCount: 3,
      followingCount: 5,
    });
  });

  it('APIエラー時は例外がそのままスローされる', async () => {
    mockGet.mockRejectedValueOnce(new Error('Network Error'));

    await expect(ProfileStatsRepository.fetchStats()).rejects.toThrow('Network Error');
  });
});
