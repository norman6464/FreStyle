import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../lib/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

import apiClient from '../../lib/axios';
import ExerciseRepository from '../ExerciseRepository';

const mockedGet = vi.mocked(apiClient.get);
const mockedPost = vi.mocked(apiClient.post);

describe('ExerciseRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('listExercises: 言語フィルタ付きで一覧を取得する', async () => {
    mockedGet.mockResolvedValue({ data: [] });
    await ExerciseRepository.listExercises('php');
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/exercises?language=php');
  });

  it('listExercises: 言語未指定なら全言語を取得する', async () => {
    mockedGet.mockResolvedValue({ data: [] });
    await ExerciseRepository.listExercises();
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/exercises');
  });

  it('getDetail: slug で詳細を取得する', async () => {
    mockedGet.mockResolvedValue({ data: { exercise: {}, examples: [] } });
    await ExerciseRepository.getDetail('php-1');
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/exercises/php-1');
  });

  it('execute: コードを実行する', async () => {
    mockedPost.mockResolvedValue({ data: { stdout: 'ok', stderr: '', exitCode: 0 } });
    const out = await ExerciseRepository.execute('<?php echo "ok";', 'php');
    expect(out.stdout).toBe('ok');
    expect(mockedPost).toHaveBeenCalledWith('/api/v2/code/execute', { code: '<?php echo "ok";', language: 'php' });
  });

  it('submit: コードを提出する', async () => {
    mockedPost.mockResolvedValue({ data: { submissionId: 1, isCorrect: true, results: [] } });
    const out = await ExerciseRepository.submit('php-1', '<?php');
    expect(out.isCorrect).toBe(true);
    expect(mockedPost).toHaveBeenCalledWith('/api/v2/exercises/php-1/submit', { code: '<?php' });
  });

  it('listSubmissions: 履歴を取得する', async () => {
    mockedGet.mockResolvedValue({ data: [] });
    await ExerciseRepository.listSubmissions('php-1');
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/exercises/php-1/submissions');
  });

  it('listExercises: slug に特殊文字が含まれてもエンコードされる', async () => {
    mockedGet.mockResolvedValue({ data: { exercise: {}, examples: [] } });
    await ExerciseRepository.getDetail('php/1');
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/exercises/php%2F1');
  });
});
