import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useLessonProgress } from '../useLessonProgress';
import LessonProgressRepository from '../../repositories/LessonProgressRepository';
import type { UserLessonProgress } from '../../types';

vi.mock('../../repositories/LessonProgressRepository', () => ({
  default: {
    list: vi.fn(),
    complete: vi.fn(),
    incomplete: vi.fn(),
  },
}));

const mocks = vi.mocked(LessonProgressRepository);

const row = (teachingMaterialId: number): UserLessonProgress => ({
  id: teachingMaterialId,
  userId: 1,
  teachingMaterialId,
  courseId: 5,
  completedAt: '2026-06-16T00:00:00Z',
  createdAt: '2026-06-16T00:00:00Z',
});

describe('useLessonProgress', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('enabled=false のときは API を叩かず空集合', async () => {
    const { result } = renderHook(() => useLessonProgress(false));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(mocks.list).not.toHaveBeenCalled();
    expect(result.current.completedIds.size).toBe(0);
  });

  it('enabled=true のとき完了一覧を読み込んで集合化する', async () => {
    mocks.list.mockResolvedValue([row(10), row(11)]);
    const { result } = renderHook(() => useLessonProgress(true));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(mocks.list).toHaveBeenCalledTimes(1);
    expect(result.current.isCompleted(10)).toBe(true);
    expect(result.current.isCompleted(11)).toBe(true);
    expect(result.current.isCompleted(12)).toBe(false);
  });

  it('完了トグルは楽観的更新し complete を呼ぶ', async () => {
    mocks.list.mockResolvedValue([]);
    mocks.complete.mockResolvedValue();
    const { result } = renderHook(() => useLessonProgress(true));
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.toggle(10, true);
    });
    expect(mocks.complete).toHaveBeenCalledWith(10);
    expect(result.current.isCompleted(10)).toBe(true);
  });

  it('完了取消は incomplete を呼んで集合から外す', async () => {
    mocks.list.mockResolvedValue([row(10)]);
    mocks.incomplete.mockResolvedValue();
    const { result } = renderHook(() => useLessonProgress(true));
    await waitFor(() => expect(result.current.isCompleted(10)).toBe(true));

    await act(async () => {
      await result.current.toggle(10, false);
    });
    expect(mocks.incomplete).toHaveBeenCalledWith(10);
    expect(result.current.isCompleted(10)).toBe(false);
  });

  it('API 失敗時は元の状態へロールバックする', async () => {
    mocks.list.mockResolvedValue([]);
    mocks.complete.mockRejectedValue(new Error('network'));
    const { result } = renderHook(() => useLessonProgress(true));
    await waitFor(() => expect(result.current.loading).toBe(false));

    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.toggle(10, true);
    });
    expect(ok).toBe(false);
    expect(result.current.isCompleted(10)).toBe(false); // ロールバック済み
  });
});
