import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useWeeklyChallenge } from '../useWeeklyChallenge';
import { WeeklyChallengeRepository } from '../../repositories/WeeklyChallengeRepository';

vi.mock('../../repositories/WeeklyChallengeRepository');
const mockedRepo = vi.mocked(WeeklyChallengeRepository);

describe('useWeeklyChallenge', () => {
  const mockChallenge = { id: 1, title: 'Test', description: 'desc', category: 'meeting', targetSessions: 3, completedSessions: 1, isCompleted: false, weekStart: '2024-01-01', weekEnd: '2024-01-07' };

  beforeEach(() => {
    vi.clearAllMocks();
    mockedRepo.fetchCurrentChallenge.mockResolvedValue(mockChallenge);
    mockedRepo.incrementProgress.mockResolvedValue({ ...mockChallenge, completedSessions: 2 });
  });

  it('チャレンジを取得する', async () => {
    const { result } = renderHook(() => useWeeklyChallenge());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.challenge).toEqual(mockChallenge);
  });

  it('進捗をインクリメントする', async () => {
    const { result } = renderHook(() => useWeeklyChallenge());
    await waitFor(() => expect(result.current.loading).toBe(false));
    await act(async () => { await result.current.incrementProgress(); });
    expect(result.current.challenge?.completedSessions).toBe(2);
  });

  it('チャレンジがない場合はnullを返す', async () => {
    mockedRepo.fetchCurrentChallenge.mockResolvedValue(null);
    const { result } = renderHook(() => useWeeklyChallenge());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.challenge).toBeNull();
  });
});
