import { describe, it, expect, vi, afterEach } from 'vitest';
import apiClient from '../../lib/axios';
import { MenuRepository } from '../MenuRepository';

vi.mock('../../lib/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('MenuRepository', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('スコア履歴を取得できる', async () => {
    const mockScores = [{ sessionId: 1, overallScore: 7.5 }];
    mockedApiClient.get.mockResolvedValue({ data: mockScores });

    const result = await MenuRepository.fetchScoreHistory();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/score-cards');
    expect(result).toEqual(mockScores);
  });
});
