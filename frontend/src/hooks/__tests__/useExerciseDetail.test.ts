import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useExerciseDetail } from '../useExerciseDetail';
import ExerciseRepository from '../../repositories/ExerciseRepository';

vi.mock('../../repositories/ExerciseRepository', () => ({
  default: {
    getDetail: vi.fn(),
    execute: vi.fn(),
    submit: vi.fn(),
    listSubmissions: vi.fn(),
  },
}));

const mocks = vi.mocked(ExerciseRepository);

const baseExercise = {
  id: 1, slug: 'php-1', language: 'php', orderIndex: 1, category: '基礎',
  title: 'Hello', description: 'desc', starterCode: '<?php echo "hi";', hintText: '',
  expectedOutput: 'hi', difficulty: 1, isPublished: true,
  createdAt: '', updatedAt: '',
};

describe('useExerciseDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.listSubmissions.mockResolvedValue([]);
  });

  it('slug を渡すと detail と starterCode をロードする', async () => {
    mocks.getDetail.mockResolvedValue({ exercise: baseExercise, examples: [] });
    const { result } = renderHook(() => useExerciseDetail('php-1'));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.detail?.exercise.slug).toBe('php-1');
    expect(result.current.code).toBe('<?php echo "hi";');
    expect(mocks.listSubmissions).toHaveBeenCalledWith('php-1');
  });

  it('runCode は実行結果を保存する', async () => {
    mocks.getDetail.mockResolvedValue({ exercise: baseExercise, examples: [] });
    mocks.execute.mockResolvedValue({ stdout: 'hi', stderr: '', exitCode: 0 });
    const { result } = renderHook(() => useExerciseDetail('php-1'));
    await waitFor(() => expect(result.current.loading).toBe(false));
    await act(async () => {
      await result.current.runCode();
    });
    expect(result.current.executionResult?.stdout).toBe('hi');
    expect(mocks.execute).toHaveBeenCalledWith('<?php echo "hi";', 'php');
  });

  it('submitCode は採点結果を保存して履歴を再取得する', async () => {
    mocks.getDetail.mockResolvedValue({ exercise: baseExercise, examples: [] });
    mocks.submit.mockResolvedValue({ submissionId: 9, isCorrect: true, results: [] });
    const { result } = renderHook(() => useExerciseDetail('php-1'));
    await waitFor(() => expect(result.current.loading).toBe(false));
    await act(async () => {
      await result.current.submitCode();
    });
    expect(result.current.submitResult?.isCorrect).toBe(true);
    expect(mocks.submit).toHaveBeenCalledWith('php-1', '<?php echo "hi";');
    // initial load + submit 後の再取得で 2 回
    expect(mocks.listSubmissions).toHaveBeenCalledTimes(2);
  });

  it('submitCode 失敗時に submitError をセットする', async () => {
    mocks.getDetail.mockResolvedValue({ exercise: baseExercise, examples: [] });
    mocks.submit.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useExerciseDetail('php-1'));
    await waitFor(() => expect(result.current.loading).toBe(false));
    await act(async () => {
      await result.current.submitCode();
    });
    expect(result.current.submitError).toBe('提出に失敗しました');
  });

  it('runCode 失敗時はエラー stderr をセット', async () => {
    mocks.getDetail.mockResolvedValue({ exercise: baseExercise, examples: [] });
    mocks.execute.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useExerciseDetail('php-1'));
    await waitFor(() => expect(result.current.loading).toBe(false));
    await act(async () => {
      await result.current.runCode();
    });
    expect(result.current.executionResult?.exitCode).toBe(1);
    expect(result.current.executionResult?.stderr).toMatch(/失敗/);
  });

  it('resetCode は starterCode に戻す', async () => {
    mocks.getDetail.mockResolvedValue({ exercise: baseExercise, examples: [] });
    const { result } = renderHook(() => useExerciseDetail('php-1'));
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.setCode('改造済み'));
    expect(result.current.code).toBe('改造済み');
    act(() => result.current.resetCode());
    expect(result.current.code).toBe('<?php echo "hi";');
  });
});
