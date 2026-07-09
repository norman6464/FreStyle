import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useCourseCompletionCounts } from '../useCourseCompletionCounts';
import LessonProgressRepository from '../../repositories/LessonProgressRepository';
import type { UserLessonProgress } from '../../types';

vi.mock('../../repositories/LessonProgressRepository', () => ({
  default: {
    list: vi.fn(),
    complete: vi.fn(),
    incomplete: vi.fn(),
  },
}));

const mockList = vi.mocked(LessonProgressRepository.list);

function row(teachingMaterialId: number, courseId: number): UserLessonProgress {
  return {
    id: teachingMaterialId,
    userId: 1,
    teachingMaterialId,
    courseId,
    completedAt: '2026-07-01T00:00:00Z',
    createdAt: '2026-07-01T00:00:00Z',
  };
}

describe('useCourseCompletionCounts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('完了行をコース別に集計する', async () => {
    mockList.mockResolvedValue([row(11, 1), row(12, 1), row(21, 2)]);
    const { result } = renderHook(() => useCourseCompletionCounts(true));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.countByCourse.get(1)).toBe(2);
    expect(result.current.countByCourse.get(2)).toBe(1);
    expect(result.current.countByCourse.get(99)).toBeUndefined();
  });

  it('enabled=false のときは API を叩かず空 Map', async () => {
    const { result } = renderHook(() => useCourseCompletionCounts(false));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(mockList).not.toHaveBeenCalled();
    expect(result.current.countByCourse.size).toBe(0);
  });

  it('API が null を返しても空 Map で続行する (FRESTYLE-70 防御)', async () => {
    mockList.mockResolvedValue(null as unknown as UserLessonProgress[]);
    const { result } = renderHook(() => useCourseCompletionCounts(true));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.countByCourse.size).toBe(0);
  });

  it('取得失敗時は空 Map のまま黙って続行する', async () => {
    mockList.mockRejectedValue(new Error('network'));
    const { result } = renderHook(() => useCourseCompletionCounts(true));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.countByCourse.size).toBe(0);
  });
});
