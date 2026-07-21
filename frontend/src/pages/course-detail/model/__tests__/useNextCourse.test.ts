import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useNextCourse } from '../useNextCourse';
import { CourseRepository } from '@/entities/course';
import type { CourseWithProgress } from '@/entities/course';

vi.mock('@/entities/course/api/courseRepository', () => ({
  default: {
    list: vi.fn(),
  },
}));

const mockList = vi.mocked(CourseRepository.list);

function course(id: number, title: string): CourseWithProgress {
  return {
    id,
    companyId: 10,
    createdByUserId: 1,
    title,
    description: '',
    category: 'architecture',
    sortOrder: id * 10,
    isPublished: true,
    materialCount: 5,
    completedCount: 0,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-01T00:00:00Z',
  };
}

describe('useNextCourse', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('一覧の並び順で次のコースを返す', async () => {
    mockList.mockResolvedValue([course(7, 'クリーンアーキテクチャ'), course(8, 'ヘキサゴナル'), course(9, 'レイヤード')]);
    const { result } = renderHook(() => useNextCourse(7, true));
    await waitFor(() => expect(result.current.nextCourse?.id).toBe(8));
    expect(result.current.nextCourse?.title).toBe('ヘキサゴナル');
  });

  it('並び順で最後のコースでは null', async () => {
    mockList.mockResolvedValue([course(7, 'A'), course(8, 'B')]);
    const { result } = renderHook(() => useNextCourse(8, true));
    await waitFor(() => expect(mockList).toHaveBeenCalled());
    expect(result.current.nextCourse).toBeNull();
  });

  it('一覧に自分が見つからない場合は null', async () => {
    mockList.mockResolvedValue([course(7, 'A'), course(8, 'B')]);
    const { result } = renderHook(() => useNextCourse(999, true));
    await waitFor(() => expect(mockList).toHaveBeenCalled());
    expect(result.current.nextCourse).toBeNull();
  });

  it('enabled=false(管理ロール)では API を叩かない', async () => {
    const { result } = renderHook(() => useNextCourse(7, false));
    await new Promise((r) => setTimeout(r, 10));
    expect(mockList).not.toHaveBeenCalled();
    expect(result.current.nextCourse).toBeNull();
  });

  it('API が null を返しても null のまま続行する (FRESTYLE-70 防御)', async () => {
    mockList.mockResolvedValue(null as unknown as CourseWithProgress[]);
    const { result } = renderHook(() => useNextCourse(7, true));
    await waitFor(() => expect(mockList).toHaveBeenCalled());
    expect(result.current.nextCourse).toBeNull();
  });

  it('取得失敗時は null のまま黙って続行する', async () => {
    mockList.mockRejectedValue(new Error('network'));
    const { result } = renderHook(() => useNextCourse(7, true));
    await waitFor(() => expect(mockList).toHaveBeenCalled());
    expect(result.current.nextCourse).toBeNull();
  });

  it('コース切替の再取得中は前回の値を残さない(即クリア)', async () => {
    mockList.mockResolvedValueOnce([course(7, 'A'), course(8, 'B')]);
    const { result, rerender } = renderHook(({ id }: { id: number }) => useNextCourse(id, true), {
      initialProps: { id: 7 },
    });
    await waitFor(() => expect(result.current.nextCourse?.id).toBe(8));

    // 2 回目の取得は解決させないままにして、切替直後の状態を観察する。
    mockList.mockReturnValue(new Promise(() => {}));
    rerender({ id: 8 });
    expect(result.current.nextCourse).toBeNull();
  });
});
