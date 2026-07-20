import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/shared/api/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

import apiClient from '@/shared/api/axios';
import TeachingMaterialRepository from '../TeachingMaterialRepository';

const mockedGet = vi.mocked(apiClient.get);
const mockedPost = vi.mocked(apiClient.post);
const mockedPut = vi.mocked(apiClient.put);
const mockedDelete = vi.mocked(apiClient.delete);

describe('TeachingMaterialRepository', () => {
  beforeEach(() => vi.clearAllMocks());

  it('get は ID 付きで GET する', async () => {
    mockedGet.mockResolvedValue({ data: { id: 5, title: 'X' } });
    const m = await TeachingMaterialRepository.get(5);
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/teaching-materials/5');
    expect(m.id).toBe(5);
  });

  it('create は courseId を含めて POST する', async () => {
    mockedPost.mockResolvedValue({ data: { id: 1 } });
    await TeachingMaterialRepository.create({
      courseId: 5,
      title: 'a',
      content: 'b',
      orderInCourse: 100,
      isPublished: true,
    });
    expect(mockedPost).toHaveBeenCalledWith('/api/v2/teaching-materials', {
      courseId: 5,
      title: 'a',
      content: 'b',
      orderInCourse: 100,
      isPublished: true,
    });
  });

  it('update は PUT する', async () => {
    mockedPut.mockResolvedValue({ data: { id: 1 } });
    await TeachingMaterialRepository.update(1, {
      title: 'x',
      content: 'y',
      orderInCourse: 110,
      isPublished: false,
    });
    expect(mockedPut).toHaveBeenCalledWith('/api/v2/teaching-materials/1', {
      title: 'x',
      content: 'y',
      orderInCourse: 110,
      isPublished: false,
    });
  });

  it('remove は DELETE する', async () => {
    mockedDelete.mockResolvedValue({});
    await TeachingMaterialRepository.remove(7);
    expect(mockedDelete).toHaveBeenCalledWith('/api/v2/teaching-materials/7');
  });
});
