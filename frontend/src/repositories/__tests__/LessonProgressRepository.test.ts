import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/shared/api/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

import apiClient from '@/shared/api/axios';
import LessonProgressRepository from '../LessonProgressRepository';

const mockedGet = vi.mocked(apiClient.get);
const mockedPost = vi.mocked(apiClient.post);
const mockedDelete = vi.mocked(apiClient.delete);

describe('LessonProgressRepository', () => {
  beforeEach(() => vi.clearAllMocks());

  it('list は完了一覧を GET する', async () => {
    mockedGet.mockResolvedValue({ data: [{ id: 1, teachingMaterialId: 10 }] });
    const rows = await LessonProgressRepository.list();
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/lesson-progress');
    expect(rows).toHaveLength(1);
  });

  it('complete は teachingMaterialId を body に POST する（userId は渡さない）', async () => {
    mockedPost.mockResolvedValue({});
    await LessonProgressRepository.complete(10);
    expect(mockedPost).toHaveBeenCalledWith('/api/v2/lesson-progress', {
      teachingMaterialId: 10,
    });
  });

  it('incomplete は ID 付きで DELETE する', async () => {
    mockedDelete.mockResolvedValue({});
    await LessonProgressRepository.incomplete(10);
    expect(mockedDelete).toHaveBeenCalledWith('/api/v2/lesson-progress/10');
  });
});
