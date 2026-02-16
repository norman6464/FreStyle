import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { DailyGoalRepository } from '../DailyGoalRepository';

const mockGet = vi.fn();
const mockPut = vi.fn();
const mockPost = vi.fn();

vi.mock('../../lib/axios', () => ({
  default: {
    get: (...args: unknown[]) => mockGet(...args),
    put: (...args: unknown[]) => mockPut(...args),
    post: (...args: unknown[]) => mockPost(...args),
  },
}));

function createMockStorage(): Storage {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
    removeItem: vi.fn((key: string) => { delete store[key]; }),
    clear: vi.fn(() => { store = {}; }),
    get length() { return Object.keys(store).length; },
    key: vi.fn((index: number) => Object.keys(store)[index] ?? null),
  };
}

describe('DailyGoalRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  describe('API正常系', () => {
    it('getToday: APIから今日のゴールを取得する', async () => {
      mockGet.mockResolvedValue({ data: { date: '2026-02-16', target: 5, completed: 2 } });

      const result = await DailyGoalRepository.getToday();

      expect(result.target).toBe(5);
      expect(result.completed).toBe(2);
      expect(mockGet).toHaveBeenCalledWith('/api/daily-goals/today');
    });

    it('setTarget: APIで目標回数を設定する', async () => {
      mockPut.mockResolvedValue({});

      await DailyGoalRepository.setTarget(7);

      expect(mockPut).toHaveBeenCalledWith('/api/daily-goals/target', { target: 7 });
    });

    it('incrementCompleted: APIで完了数をインクリメントする', async () => {
      mockPost.mockResolvedValue({ data: { date: '2026-02-16', target: 3, completed: 1 } });

      const result = await DailyGoalRepository.incrementCompleted();

      expect(result.completed).toBe(1);
      expect(mockPost).toHaveBeenCalledWith('/api/daily-goals/increment');
    });
  });

  describe('APIエラー時のlocalStorageフォールバック', () => {
    it('getToday: APIエラー時はlocalStorageから取得する', async () => {
      mockGet.mockRejectedValue(new Error('Network Error'));
      const today = new Date().toISOString().split('T')[0];
      localStorage.setItem('freestyle_daily_goal', JSON.stringify({ date: today, target: 4, completed: 1 }));

      const result = await DailyGoalRepository.getToday();

      expect(result.target).toBe(4);
      expect(result.completed).toBe(1);
    });

    it('getToday: APIエラー・localStorage空の場合はデフォルトを返す', async () => {
      mockGet.mockRejectedValue(new Error('Network Error'));

      const result = await DailyGoalRepository.getToday();

      expect(result.target).toBe(3);
      expect(result.completed).toBe(0);
    });

    it('setTarget: APIエラー時はlocalStorageに保存する', async () => {
      mockPut.mockRejectedValue(new Error('Network Error'));

      await DailyGoalRepository.setTarget(5);

      expect(localStorage.setItem).toHaveBeenCalled();
    });

    it('incrementCompleted: APIエラー時はlocalStorageを更新する', async () => {
      mockPost.mockRejectedValue(new Error('Network Error'));
      const today = new Date().toISOString().split('T')[0];
      localStorage.setItem('freestyle_daily_goal', JSON.stringify({ date: today, target: 3, completed: 1 }));

      const result = await DailyGoalRepository.incrementCompleted();

      expect(result.completed).toBe(2);
    });
  });
});
