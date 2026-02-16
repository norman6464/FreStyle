import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useDailyGoal } from '../useDailyGoal';

const mockGetToday = vi.fn();
const mockSetTarget = vi.fn();
const mockIncrementCompleted = vi.fn();

vi.mock('../../repositories/DailyGoalRepository', () => ({
  DailyGoalRepository: {
    getToday: (...args: unknown[]) => mockGetToday(...args),
    setTarget: (...args: unknown[]) => mockSetTarget(...args),
    incrementCompleted: (...args: unknown[]) => mockIncrementCompleted(...args),
  },
}));

describe('useDailyGoal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetToday.mockResolvedValue({ date: '2026-02-16', target: 3, completed: 0 });
    mockSetTarget.mockResolvedValue(undefined);
    mockIncrementCompleted.mockResolvedValue({ date: '2026-02-16', target: 3, completed: 1 });
  });

  it('初期状態はloading=trueでデフォルト値', () => {
    const { result } = renderHook(() => useDailyGoal());
    expect(result.current.loading).toBe(true);
    expect(result.current.goal.target).toBe(3);
  });

  it('マウント時にAPIから今日の目標を取得する', async () => {
    mockGetToday.mockResolvedValue({ date: '2026-02-16', target: 5, completed: 2 });

    const { result } = renderHook(() => useDailyGoal());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.goal.target).toBe(5);
    expect(result.current.goal.completed).toBe(2);
  });

  it('目標達成時にisAchievedがtrueになる', async () => {
    mockGetToday.mockResolvedValue({ date: '2026-02-16', target: 3, completed: 3 });

    const { result } = renderHook(() => useDailyGoal());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.isAchieved).toBe(true);
  });

  it('目標数を更新できる', async () => {
    mockGetToday
      .mockResolvedValueOnce({ date: '2026-02-16', target: 3, completed: 0 })
      .mockResolvedValueOnce({ date: '2026-02-16', target: 5, completed: 0 });

    const { result } = renderHook(() => useDailyGoal());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.setTarget(5);
    });

    expect(mockSetTarget).toHaveBeenCalledWith(5);
    expect(result.current.goal.target).toBe(5);
  });

  it('完了数をインクリメントできる', async () => {
    mockGetToday.mockResolvedValue({ date: '2026-02-16', target: 3, completed: 0 });

    const { result } = renderHook(() => useDailyGoal());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.incrementCompleted();
    });

    expect(mockIncrementCompleted).toHaveBeenCalled();
    expect(result.current.goal.completed).toBe(1);
  });

  it('進捗率が正しく計算される', async () => {
    mockGetToday.mockResolvedValue({ date: '2026-02-16', target: 4, completed: 2 });

    const { result } = renderHook(() => useDailyGoal());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.progress).toBe(50);
  });

  it('target=0の場合progressが100になる', async () => {
    mockGetToday.mockResolvedValue({ date: '2026-02-16', target: 0, completed: 0 });

    const { result } = renderHook(() => useDailyGoal());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.progress).toBe(100);
  });
});
