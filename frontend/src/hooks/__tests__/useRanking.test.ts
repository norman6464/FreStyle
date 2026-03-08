import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useRanking } from '../useRanking';
import { RankingRepository } from '../../repositories/RankingRepository';

vi.mock('../../repositories/RankingRepository');
const mockedRepo = vi.mocked(RankingRepository);

describe('useRanking', () => {
  const mockRanking = {
    entries: [
      { rank: 1, userId: 1, username: 'user1', iconUrl: null, averageScore: 9.0, sessionCount: 5 },
      { rank: 2, userId: 2, username: 'user2', iconUrl: null, averageScore: 8.0, sessionCount: 3 },
    ],
    myRanking: { rank: 2, userId: 2, username: 'user2', iconUrl: null, averageScore: 8.0, sessionCount: 3 },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockedRepo.fetchRanking.mockResolvedValue(mockRanking);
  });

  it('初期ロード時にweeklyのランキングを取得する', async () => {
    const { result } = renderHook(() => useRanking());

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.ranking).toEqual(mockRanking);
    expect(result.current.period).toBe('weekly');
    expect(mockedRepo.fetchRanking).toHaveBeenCalledWith('weekly');
  });

  it('期間を変更するとデータを再取得する', async () => {
    const { result } = renderHook(() => useRanking());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    act(() => {
      result.current.changePeriod('monthly');
    });

    await waitFor(() => {
      expect(mockedRepo.fetchRanking).toHaveBeenCalledWith('monthly');
    });
  });

  it('エラー時にエラーメッセージを設定する', async () => {
    mockedRepo.fetchRanking.mockRejectedValue(new Error('API Error'));

    const { result } = renderHook(() => useRanking());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe('ランキングの取得に失敗しました');
  });
});
