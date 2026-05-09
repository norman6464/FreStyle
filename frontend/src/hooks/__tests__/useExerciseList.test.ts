import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useExerciseList } from '../useExerciseList';
import ExerciseRepository from '../../repositories/ExerciseRepository';

vi.mock('../../repositories/ExerciseRepository', () => ({
  default: { listExercises: vi.fn() },
}));

const mockList = vi.mocked(ExerciseRepository.listExercises);

describe('useExerciseList', () => {
  beforeEach(() => vi.clearAllMocks());

  it('初期化時に PHP の一覧を取得する', async () => {
    mockList.mockResolvedValue([]);
    const { result } = renderHook(() => useExerciseList());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(mockList).toHaveBeenCalledWith('php');
  });

  it('language を切り替えると再 fetch する', async () => {
    mockList.mockResolvedValue([]);
    const { result } = renderHook(() => useExerciseList());
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.setLanguage(''));
    await waitFor(() => expect(mockList).toHaveBeenCalledTimes(2));
    expect(mockList).toHaveBeenLastCalledWith(undefined);
  });

  it('取得失敗時は error をセットする', async () => {
    mockList.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useExerciseList());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe('演習問題の取得に失敗しました');
  });

  it('カテゴリのユニーク一覧を返す', async () => {
    mockList.mockResolvedValue([
      { id: 1, slug: 'a', category: '基礎', orderIndex: 1, language: 'php', title: '', description: '', starterCode: '', hintText: '', expectedOutput: '', difficulty: 1, isPublished: true, createdAt: '', updatedAt: '', status: '' as const, stats: { totalSubmissions: 0, solvedUsers: 0 } },
      { id: 2, slug: 'b', category: '基礎', orderIndex: 2, language: 'php', title: '', description: '', starterCode: '', hintText: '', expectedOutput: '', difficulty: 1, isPublished: true, createdAt: '', updatedAt: '', status: '' as const, stats: { totalSubmissions: 0, solvedUsers: 0 } },
      { id: 3, slug: 'c', category: '応用', orderIndex: 3, language: 'php', title: '', description: '', starterCode: '', hintText: '', expectedOutput: '', difficulty: 2, isPublished: true, createdAt: '', updatedAt: '', status: '' as const, stats: { totalSubmissions: 0, solvedUsers: 0 } },
    ]);
    const { result } = renderHook(() => useExerciseList());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.categories).toEqual(['基礎', '応用']);
  });
});
