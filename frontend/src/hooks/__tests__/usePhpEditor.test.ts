import { renderHook, act, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { usePhpEditor } from '../usePhpEditor';
import PhpRepository from '../../repositories/PhpRepository';
import type { PhpExercise, CodeExecutionResult } from '../../types';

vi.mock('../../repositories/PhpRepository');

const mockExercises: PhpExercise[] = [
  {
    id: 1,
    orderIndex: 1,
    category: '基礎',
    title: 'こんにちは世界',
    description: 'Hello World を表示しましょう',
    starterCode: '<?php\necho "Hello, World!\\n";',
    hintText: 'echo を使います',
    expectedOutput: 'Hello, World!',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 2,
    orderIndex: 2,
    category: '基礎',
    title: '変数と文字列',
    description: '変数を使いましょう',
    starterCode: '<?php\n$name = "test";',
    hintText: '$ で変数を宣言します',
    expectedOutput: 'test',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  },
];

describe('usePhpEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(PhpRepository.listExercises).mockResolvedValue(mockExercises);
  });

  it('初期ロード時に演習一覧を取得し最初の問題を選択する', async () => {
    const { result } = renderHook(() => usePhpEditor());

    expect(result.current.loadingExercises).toBe(true);

    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    expect(result.current.exercises).toHaveLength(2);
    expect(result.current.selectedExercise?.id).toBe(1);
    expect(result.current.code).toBe(mockExercises[0].starterCode);
  });

  it('selectExercise で選択問題が切り替わる', async () => {
    const { result } = renderHook(() => usePhpEditor());
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
    vi.mocked(PhpRepository.execute).mockResolvedValue(mockResult);

    const { result } = renderHook(() => usePhpEditor());
    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    await act(async () => {
      await result.current.runCode();
    });

    expect(PhpRepository.execute).toHaveBeenCalledWith(mockExercises[0].starterCode);
    expect(result.current.result).toEqual(mockResult);
    expect(result.current.running).toBe(false);
  });

  it('resetCode でスターターコードに戻る', async () => {
    const { result } = renderHook(() => usePhpEditor());
    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    act(() => { result.current.setCode('変更後コード'); });
    act(() => { result.current.resetCode(); });

    expect(result.current.code).toBe(mockExercises[0].starterCode);
  });

  it('categories が重複なしで返る', async () => {
    const { result } = renderHook(() => usePhpEditor());
    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    expect(result.current.categories).toEqual(['基礎']);
  });

  it('API エラー時に error がセットされる', async () => {
    vi.mocked(PhpRepository.listExercises).mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => usePhpEditor());
    await waitFor(() => expect(result.current.loadingExercises).toBe(false));

    expect(result.current.error).toBeTruthy();
  });
});
