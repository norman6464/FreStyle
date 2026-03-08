import { describe, it, expect, vi, beforeEach } from 'vitest';
import apiClient from '../../lib/axios';
import { WeeklyChallengeRepository } from '../WeeklyChallengeRepository';

vi.mock('../../lib/axios');
const mockedApiClient = vi.mocked(apiClient);

describe('WeeklyChallengeRepository', () => {
  beforeEach(() => vi.clearAllMocks());

  it('今週のチャレンジを取得する', async () => {
    const mockData = { id: 1, title: 'Test', targetSessions: 3, completedSessions: 0, isCompleted: false };
    mockedApiClient.get.mockResolvedValue({ data: mockData });
    const result = await WeeklyChallengeRepository.fetchCurrentChallenge();
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/weekly-challenge');
    expect(result).toEqual(mockData);
  });

  it('進捗をインクリメントする', async () => {
    const mockData = { id: 1, completedSessions: 1 };
    mockedApiClient.post.mockResolvedValue({ data: mockData });
    const result = await WeeklyChallengeRepository.incrementProgress();
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/weekly-challenge/progress');
    expect(result).toEqual(mockData);
  });

  it('エラー時にnullを返す', async () => {
    mockedApiClient.get.mockRejectedValue(new Error('fail'));
    const result = await WeeklyChallengeRepository.fetchCurrentChallenge();
    expect(result).toBeNull();
  });
});
