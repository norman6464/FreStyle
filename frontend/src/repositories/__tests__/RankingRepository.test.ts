import { describe, it, expect, vi, beforeEach } from 'vitest';
import apiClient from '../../lib/axios';
import { RankingRepository } from '../RankingRepository';

vi.mock('../../lib/axios');
const mockedApiClient = vi.mocked(apiClient);

describe('RankingRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('fetchRanking', () => {
    it('デフォルトでweeklyのランキングを取得する', async () => {
      const mockData = { entries: [], myRanking: null };
      mockedApiClient.get.mockResolvedValue({ data: mockData });

      const result = await RankingRepository.fetchRanking();

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/ranking', {
        params: { period: 'weekly' },
      });
      expect(result).toEqual(mockData);
    });

    it('指定した期間でランキングを取得する', async () => {
      const mockData = { entries: [], myRanking: null };
      mockedApiClient.get.mockResolvedValue({ data: mockData });

      await RankingRepository.fetchRanking('monthly');

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/ranking', {
        params: { period: 'monthly' },
      });
    });
  });
});
