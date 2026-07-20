import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/shared/api/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

import apiClient from '@/shared/api/axios';
import ExerciseRepository from '../exerciseRepository';

const mockedGet = vi.mocked(apiClient.get);
const mockedPost = vi.mocked(apiClient.post);

describe('ExerciseRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('listExercises: 言語フィルタ付きで一覧を取得する', async () => {
    mockedGet.mockResolvedValue({ data: { items: [], hasNext: false, offset: 0, limit: 20 } });
    await ExerciseRepository.listExercises('php');
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/exercises?language=php&offset=0&limit=20');
  });

  it('listExercises: 言語未指定なら全言語を取得する', async () => {
    mockedGet.mockResolvedValue({ data: { items: [], hasNext: false, offset: 0, limit: 20 } });
    await ExerciseRepository.listExercises();
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/exercises?offset=0&limit=20');
  });

  it('listExercises: offset と limit を指定できる', async () => {
    mockedGet.mockResolvedValue({ data: { items: [], hasNext: true, offset: 20, limit: 20 } });
    await ExerciseRepository.listExercises('go', 20, 20);
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/exercises?language=go&offset=20&limit=20');
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

  it('warmup: 言語を指定して実行環境を温める', async () => {
    mockedPost.mockResolvedValue({ data: { ready: true } });
    await ExerciseRepository.warmup('go');
    expect(mockedPost).toHaveBeenCalledWith('/api/v2/code/warmup', { language: 'go' });
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
