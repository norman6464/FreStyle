import { renderHook, act, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useExerciseEditor } from '../useExerciseEditor';
import ExerciseRepository from '../../repositories/ExerciseRepository';
import type { MasterExercise, CodeExecutionResult } from '../../types';

vi.mock('../../repositories/ExerciseRepository');

const mockExercises: MasterExercise[] = [
  {
    id: 1,
    slug: 'php-1',
    language: 'php',
    orderIndex: 1,
    category: '基礎',
    title: 'こんにちは世界',
    description: 'Hello World を表示しましょう',
    starterCode: '<?php\necho "Hello, World!\\n";',
    hintText: 'echo を使います',
    expectedOutput: 'Hello, World!',
    difficulty: 1,
    isPublished: true,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 2,
    slug: 'php-2',
    language: 'php',
    orderIndex: 2,
    category: '基礎',
    title: '変数と文字列',
    description: '変数を使いましょう',
    starterCode: '<?php\n$name = "test";',
    hintText: '$ で変数を宣言します',
    expectedOutput: 'test',
    difficulty: 1,
    isPublished: true,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  },
];

describe('useExerciseEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(ExerciseRepository.listExercises).mockResolvedValue(mockExercises);
  });

  it('初期ロード時に演習一覧を取得し最初の問題を選択する', async () => {
    const { result } = renderHook(() => useExerciseEditor());

    expect(result.current.loadingExercises).toBe(true);

    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    expect(result.current.exercises).toHaveLength(2);
    expect(result.current.selectedExercise?.id).toBe(1);
    expect(result.current.code).toBe(mockExercises[0].starterCode);
  });

  it('listExercises はデフォルトで language="php" を渡す', async () => {
    const { result } = renderHook(() => useExerciseEditor());
    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    expect(ExerciseRepository.listExercises).toHaveBeenCalledWith('php');
  });

  it('language を引数で切り替えられる', async () => {
    const { result } = renderHook(() => useExerciseEditor('sql'));
    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    expect(ExerciseRepository.listExercises).toHaveBeenCalledWith('sql');
  });

  it('selectExercise で選択問題が切り替わる', async () => {
    const { result } = renderHook(() => useExerciseEditor());
    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    act(() => {
      result.current.selectExercise(mockExercises[1]);
    });

    expect(result.current.selectedExercise?.id).toBe(2);
    expect(result.current.code).toBe(mockExercises[1].starterCode);
    expect(result.current.result).toBeNull();
    expect(result.current.showHint).toBe(false);
  });

  it('runCode でコードを実行し結果を受け取る', async () => {
    const mockResult: CodeExecutionResult = { stdout: 'Hello, World!\n', stderr: '', exitCode: 0 };
    vi.mocked(ExerciseRepository.execute).mockResolvedValue(mockResult);

    const { result } = renderHook(() => useExerciseEditor());
    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    await act(async () => {
      await result.current.runCode();
    });

    // 選択中の問題の language で実行される
    expect(ExerciseRepository.execute).toHaveBeenCalledWith(mockExercises[0].starterCode, 'php');
    expect(result.current.result).toEqual(mockResult);
    expect(result.current.running).toBe(false);
  });

  it('resetCode でスターターコードに戻る', async () => {
    const { result } = renderHook(() => useExerciseEditor());
    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    act(() => { result.current.setCode('変更後コード'); });
    act(() => { result.current.resetCode(); });

    expect(result.current.code).toBe(mockExercises[0].starterCode);
  });

  it('categories が重複なしで返る', async () => {
    const { result } = renderHook(() => useExerciseEditor());
    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    expect(result.current.categories).toEqual(['基礎']);
  });

  it('API エラー時に error がセットされる', async () => {
    vi.mocked(ExerciseRepository.listExercises).mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useExerciseEditor());
    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    expect(result.current.error).toBeTruthy();
  });
});
