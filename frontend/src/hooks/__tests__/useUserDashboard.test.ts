import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useUserDashboard } from '../useUserDashboard';
import DashboardRepository from '../../repositories/DashboardRepository';
import type { UserDashboard } from '@/types';

vi.mock('../../repositories/DashboardRepository');

const mockGet = vi.mocked(DashboardRepository.get);

const sample: UserDashboard = {
  streak: 1,
  totalExercises: 3,
  totalCorrect: 2,
  totalLessons: 4,
  recentActivity: [],
  recentChapterViews: [],
};

describe('useUserDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('enabled=true のとき取得してダッシュボードを返す', async () => {
    mockGet.mockResolvedValue(sample);
    const { result } = renderHook(() => useUserDashboard());

    expect(result.current.loading).toBe(true);
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.dashboard).toEqual(sample);
    expect(result.current.error).toBeNull();
    expect(mockGet).toHaveBeenCalledTimes(1);
  });

  it('取得に失敗したらエラーメッセージを立てる', async () => {
    mockGet.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useUserDashboard());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.dashboard).toBeNull();
    expect(result.current.error).toBe('ダッシュボードの取得に失敗しました');
  });

  it('enabled=false のときはリクエストを発行しない', () => {
    const { result } = renderHook(() => useUserDashboard({ enabled: false }));

    expect(result.current.loading).toBe(false);
    expect(result.current.dashboard).toBeNull();
    expect(mockGet).not.toHaveBeenCalled();
  });
});
