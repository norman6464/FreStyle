import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useExerciseList } from '../useExerciseList';
import ExerciseRepository from '../../repositories/ExerciseRepository';
import { ExercisePage } from '../../types';

vi.mock('../../repositories/ExerciseRepository', () => ({
  default: { listExercises: vi.fn() },
}));

const mockList = vi.mocked(ExerciseRepository.listExercises);

const emptyPage: ExercisePage = { items: [], hasNext: false, offset: 0, limit: 20 };

function makePage(items: ExercisePage['items'], hasNext = false): ExercisePage {
  return { items, hasNext, offset: 0, limit: 20 };
}

const baseItem = {
  id: 1, slug: 'a', category: '基礎', orderIndex: 1, language: 'php',
  title: '', difficulty: 1, mode: 'qa' as const, isPublished: true,
  status: '' as const, stats: { totalSubmissions: 0, solvedUsers: 0 },
};

describe('useExerciseList', () => {
  beforeEach(() => vi.clearAllMocks());

  it('初期化時に PHP の一覧を取得する', async () => {
    mockList.mockResolvedValue(emptyPage);
    const { result } = renderHook(() => useExerciseList());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(mockList).toHaveBeenCalledWith('php', 0, 20);
  });

  it('language を切り替えると items をリセットして再 fetch する', async () => {
    mockList.mockResolvedValue(emptyPage);
    const { result } = renderHook(() => useExerciseList());
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.setLanguage(''));
    await waitFor(() => expect(mockList).toHaveBeenCalledTimes(2));
    expect(mockList).toHaveBeenLastCalledWith(undefined, 0, 20);
  });

  it('取得失敗時は error をセットする', async () => {
    mockList.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useExerciseList());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe('演習問題の取得に失敗しました');
  });

  it('カテゴリのユニーク一覧を返す', async () => {
    mockList.mockResolvedValue(makePage([
      { ...baseItem, id: 1, slug: 'a', category: '基礎' },
      { ...baseItem, id: 2, slug: 'b', category: '基礎' },
      { ...baseItem, id: 3, slug: 'c', category: '応用', difficulty: 2 },
    ]));
    const { result } = renderHook(() => useExerciseList());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.categories).toEqual(['基礎', '応用']);
  });

  it('hasNext が true のとき loadMore を呼ぶと追記される', async () => {
    mockList
      .mockResolvedValueOnce(makePage([{ ...baseItem, id: 1, slug: 'p1' }], true))
      .mockResolvedValueOnce(makePage([{ ...baseItem, id: 2, slug: 'p2' }], false));

    const { result } = renderHook(() => useExerciseList());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.exercises).toHaveLength(1);

    act(() => result.current.loadMore());
    await waitFor(() => expect(result.current.loadingMore).toBe(false));
    expect(result.current.exercises).toHaveLength(2);
    expect(result.current.hasNext).toBe(false);
  });

  it('hasNext が false のとき loadMore は何もしない', async () => {
    mockList.mockResolvedValue(emptyPage);
    const { result } = renderHook(() => useExerciseList());
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.loadMore());
    await waitFor(() => expect(result.current.loadingMore).toBe(false));
    expect(mockList).toHaveBeenCalledTimes(1);
  });
});
