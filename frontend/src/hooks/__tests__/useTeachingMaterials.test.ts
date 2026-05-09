import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useTeachingMaterials } from '../useTeachingMaterials';
import TeachingMaterialRepository from '../../repositories/TeachingMaterialRepository';

vi.mock('../../repositories/TeachingMaterialRepository', () => ({
  default: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
}));

const mocks = vi.mocked(TeachingMaterialRepository);

const sample = (id: number, overrides: Partial<{ title: string; isPublished: boolean }> = {}) => ({
  id,
  companyId: 1,
  createdByUserId: 1,
  title: overrides.title ?? `教材${id}`,
  content: '',
  isPublished: overrides.isPublished ?? true,
  createdAt: '',
  updatedAt: '',
});

describe('useTeachingMaterials', () => {
  beforeEach(() => vi.clearAllMocks());

  it('初期化で list が呼ばれて結果が state に入る', async () => {
    mocks.list.mockResolvedValue([sample(1), sample(2)]);
    const { result } = renderHook(() => useTeachingMaterials());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.materials).toHaveLength(2);
  });

  it('list 失敗時に error をセット', async () => {
    mocks.list.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useTeachingMaterials());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe('教材の取得に失敗しました');
  });

  it('searchQuery でタイトル / 本文がフィルタされる', async () => {
    mocks.list.mockResolvedValue([
      sample(1, { title: 'PHP 入門' }),
      sample(2, { title: 'Go 入門' }),
    ]);
    const { result } = renderHook(() => useTeachingMaterials());
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.setSearchQuery('php'));
    expect(result.current.filtered).toHaveLength(1);
    expect(result.current.filtered[0].title).toBe('PHP 入門');
  });

  it('create 成功時にリスト先頭に追加され selected になる', async () => {
    mocks.list.mockResolvedValue([]);
    mocks.create.mockResolvedValue(sample(99));
    const { result } = renderHook(() => useTeachingMaterials());
    await waitFor(() => expect(result.current.loading).toBe(false));
    await act(async () => {
      await result.current.create({ title: '新', content: '', isPublished: false });
    });
    expect(result.current.materials[0].id).toBe(99);
    expect(result.current.selectedId).toBe(99);
  });

  it('remove 成功時にリストから消え selected もクリア', async () => {
    mocks.list.mockResolvedValue([sample(1)]);
    mocks.remove.mockResolvedValue(undefined);
    const { result } = renderHook(() => useTeachingMaterials());
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.selectMaterial(1));
    await act(async () => {
      await result.current.remove(1);
    });
    expect(result.current.materials).toHaveLength(0);
    expect(result.current.selectedId).toBeNull();
  });
});
