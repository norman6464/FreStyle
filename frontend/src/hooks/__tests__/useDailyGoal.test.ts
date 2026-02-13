import { describe, it, expect, vi, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDailyGoal } from '../useDailyGoal';
import { DailyGoalRepository } from '../../repositories/DailyGoalRepository';

vi.mock('../../repositories/DailyGoalRepository');

const mockedRepo = vi.mocked(DailyGoalRepository);

describe('useDailyGoal', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('初期状態で今日の目標を取得する', () => {
    mockedRepo.getToday.mockReturnValue({ date: '2026-02-13', target: 3, completed: 1 });

    const { result } = renderHook(() => useDailyGoal());

    expect(result.current.goal.target).toBe(3);
    expect(result.current.goal.completed).toBe(1);
    expect(result.current.isAchieved).toBe(false);
  });

  it('目標達成時にisAchievedがtrueになる', () => {
    mockedRepo.getToday.mockReturnValue({ date: '2026-02-13', target: 3, completed: 3 });

    const { result } = renderHook(() => useDailyGoal());

    expect(result.current.isAchieved).toBe(true);
  });

  it('目標数を更新できる', () => {
    mockedRepo.getToday
      .mockReturnValueOnce({ date: '2026-02-13', target: 3, completed: 0 })
      .mockReturnValueOnce({ date: '2026-02-13', target: 5, completed: 0 });

    const { result } = renderHook(() => useDailyGoal());

    act(() => {
      result.current.setTarget(5);
    });

    expect(mockedRepo.setTarget).toHaveBeenCalledWith(5);
    expect(result.current.goal.target).toBe(5);
  });

  it('完了数を増加できる', () => {
    mockedRepo.getToday
      .mockReturnValueOnce({ date: '2026-02-13', target: 3, completed: 0 })
      .mockReturnValueOnce({ date: '2026-02-13', target: 3, completed: 1 });

    const { result } = renderHook(() => useDailyGoal());

    act(() => {
      result.current.incrementCompleted();
    });

    expect(mockedRepo.incrementCompleted).toHaveBeenCalled();
    expect(result.current.goal.completed).toBe(1);
  });

  it('進捗率が正しく計算される', () => {
    mockedRepo.getToday.mockReturnValue({ date: '2026-02-13', target: 4, completed: 2 });

    const { result } = renderHook(() => useDailyGoal());

    expect(result.current.progress).toBe(50);
  });
});
