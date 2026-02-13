import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { DailyGoalRepository } from '../DailyGoalRepository';

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
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('初期状態ではデフォルト目標を返す', () => {
    const goal = DailyGoalRepository.getToday();
    expect(goal.target).toBe(3);
    expect(goal.completed).toBe(0);
    expect(goal.date).toBeDefined();
  });

  it('目標数を設定できる', () => {
    DailyGoalRepository.setTarget(5);
    const goal = DailyGoalRepository.getToday();
    expect(goal.target).toBe(5);
  });

  it('完了数を増加できる', () => {
    DailyGoalRepository.incrementCompleted();
    const goal = DailyGoalRepository.getToday();
    expect(goal.completed).toBe(1);

    DailyGoalRepository.incrementCompleted();
    expect(DailyGoalRepository.getToday().completed).toBe(2);
  });

  it('日付が変わるとリセットされる', () => {
    DailyGoalRepository.incrementCompleted();
    expect(DailyGoalRepository.getToday().completed).toBe(1);

    // 別の日付のデータを直接セット
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    const yesterdayStr = yesterday.toISOString().split('T')[0];
    localStorage.setItem(
      'freestyle_daily_goal',
      JSON.stringify({ date: yesterdayStr, target: 3, completed: 2 })
    );

    const goal = DailyGoalRepository.getToday();
    expect(goal.completed).toBe(0);
    expect(goal.target).toBe(3);
  });
});
