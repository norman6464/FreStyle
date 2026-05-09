import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../lib/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

import apiClient from '../../lib/axios';
import TeachingMaterialRepository from '../TeachingMaterialRepository';

const mockedGet = vi.mocked(apiClient.get);
const mockedPost = vi.mocked(apiClient.post);
const mockedPut = vi.mocked(apiClient.put);
const mockedDelete = vi.mocked(apiClient.delete);

describe('TeachingMaterialRepository', () => {
  beforeEach(() => vi.clearAllMocks());

  it('list は /api/v2/teaching-materials を GET する', async () => {
    mockedGet.mockResolvedValue({ data: [] });
    await TeachingMaterialRepository.list();
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/teaching-materials');
  });

  it('get は ID 付きで GET する', async () => {
    mockedGet.mockResolvedValue({ data: { id: 5, title: 'X' } });
    const m = await TeachingMaterialRepository.get(5);
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/teaching-materials/5');
    expect(m.id).toBe(5);
  });

  it('create は POST する', async () => {
    mockedPost.mockResolvedValue({ data: { id: 1 } });
    await TeachingMaterialRepository.create({ title: 'a', content: 'b', isPublished: true });
    expect(mockedPost).toHaveBeenCalledWith('/api/v2/teaching-materials', {
      title: 'a',
      content: 'b',
      isPublished: true,
    });
  });

  it('update は PUT する', async () => {
    mockedPut.mockResolvedValue({ data: { id: 1 } });
    await TeachingMaterialRepository.update(1, { title: 'x', content: 'y', isPublished: false });
    expect(mockedPut).toHaveBeenCalledWith('/api/v2/teaching-materials/1', {
      title: 'x',
      content: 'y',
      isPublished: false,
    });
  });

  it('remove は DELETE する', async () => {
    mockedDelete.mockResolvedValue({});
    await TeachingMaterialRepository.remove(7);
    expect(mockedDelete).toHaveBeenCalledWith('/api/v2/teaching-materials/7');
  });
});
