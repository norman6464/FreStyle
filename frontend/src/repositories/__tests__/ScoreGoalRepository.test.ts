import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../lib/axios', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
  },
}));

import apiClient from '../../lib/axios';

const mockedGet = vi.mocked(apiClient.get);
const mockedPut = vi.mocked(apiClient.put);

describe('ScoreGoalRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchGoal: API成功時にgoalScoreを返す', async () => {
    mockedGet.mockResolvedValue({ data: { goalScore: 9.0 } });
    const { ScoreGoalRepository } = await import('../ScoreGoalRepository');
    const result = await ScoreGoalRepository.fetchGoal();
    expect(result).toBe(9.0);
    expect(mockedGet).toHaveBeenCalledWith('/api/score-goal');
  });

  it('fetchGoal: API失敗時にnullを返す', async () => {
    mockedGet.mockRejectedValue(new Error('Network error'));
    const { ScoreGoalRepository } = await import('../ScoreGoalRepository');
    const result = await ScoreGoalRepository.fetchGoal();
    expect(result).toBeNull();
  });

  it('saveGoal: APIにPUTリクエストを送る', async () => {
    mockedPut.mockResolvedValue({ data: {} });
    const { ScoreGoalRepository } = await import('../ScoreGoalRepository');
    await ScoreGoalRepository.saveGoal(7.5);
    expect(mockedPut).toHaveBeenCalledWith('/api/score-goal', { goalScore: 7.5 });
  });

  it('saveGoal: API失敗時にエラーをスローしない', async () => {
    mockedPut.mockRejectedValue(new Error('Network error'));
    const { ScoreGoalRepository } = await import('../ScoreGoalRepository');
    await expect(ScoreGoalRepository.saveGoal(7.5)).resolves.toBeUndefined();
  });
});
