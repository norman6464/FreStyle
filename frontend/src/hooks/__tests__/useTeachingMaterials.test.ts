import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useTeachingMaterials } from '../useTeachingMaterials';
import CourseRepository from '../../repositories/CourseRepository';
import TeachingMaterialRepository from '../../repositories/TeachingMaterialRepository';

vi.mock('../../repositories/CourseRepository', () => ({
  default: {
    list: vi.fn(),
    get: vi.fn(),
    listMaterials: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
}));

vi.mock('../../repositories/TeachingMaterialRepository', () => ({
  default: {
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
}));

const courseMocks = vi.mocked(CourseRepository);
const materialMocks = vi.mocked(TeachingMaterialRepository);

const sample = (
  id: number,
  overrides: Partial<{ title: string; isPublished: boolean; orderInCourse: number }> = {},
) => ({
  id,
  companyId: 1,
  courseId: 5,
  createdByUserId: 1,
  title: overrides.title ?? `教材${id}`,
  content: '',
  orderInCourse: overrides.orderInCourse ?? id * 10,
  isPublished: overrides.isPublished ?? true,
  createdAt: '',
  updatedAt: '',
});

describe('useTeachingMaterials', () => {
  beforeEach(() => vi.clearAllMocks());

  it('courseId が null なら API は呼ばれず空配列のまま', async () => {
    const { result } = renderHook(() => useTeachingMaterials(null));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(courseMocks.listMaterials).not.toHaveBeenCalled();
    expect(result.current.materials).toEqual([]);
  });

  it('courseId 指定で listMaterials が呼ばれて結果が state に入る', async () => {
    courseMocks.listMaterials.mockResolvedValue([sample(1), sample(2)]);
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(courseMocks.listMaterials).toHaveBeenCalledWith(5);
    expect(result.current.materials).toHaveLength(2);
  });

  it('listMaterials 失敗時に error をセット', async () => {
    courseMocks.listMaterials.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe('教材の取得に失敗しました');
  });

  it('searchQuery でタイトル / 本文がフィルタされる', async () => {
    courseMocks.listMaterials.mockResolvedValue([
      sample(1, { title: 'PHP 入門' }),
      sample(2, { title: 'Go 入門' }),
    ]);
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.setSearchQuery('php'));
    expect(result.current.filtered).toHaveLength(1);
    expect(result.current.filtered[0].title).toBe('PHP 入門');
  });

  it('create 成功時にリストに追加され selected になる', async () => {
    courseMocks.listMaterials.mockResolvedValue([]);
    materialMocks.create.mockResolvedValue(sample(99, { orderInCourse: 100 }));
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));
    await act(async () => {
      await result.current.create({
        title: '新',
        content: '',
        orderInCourse: 100,
        isPublished: false,
      });
    });
    expect(materialMocks.create).toHaveBeenCalledWith(expect.objectContaining({ courseId: 5 }));
    expect(result.current.materials).toHaveLength(1);
    expect(result.current.materials[0].id).toBe(99);
    expect(result.current.selectedId).toBe(99);
  });

  it('remove 成功時にリストから消え selected もクリア', async () => {
    courseMocks.listMaterials.mockResolvedValue([sample(1)]);
    materialMocks.remove.mockResolvedValue(undefined);
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.selectMaterial(1));
    await act(async () => {
      await result.current.remove(1);
    });
    expect(result.current.materials).toHaveLength(0);
    expect(result.current.selectedId).toBeNull();
  });
});
