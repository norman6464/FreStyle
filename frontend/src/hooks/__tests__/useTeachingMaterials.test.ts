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

  it('searchQuery でタイトルがフィルタされる', async () => {
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

  it('選択した教材だけ本文を都度取得して selected に入る（全章は先読みしない）', async () => {
    courseMocks.listMaterials.mockResolvedValue([sample(1), sample(2)]);
    materialMocks.get.mockResolvedValue({ ...sample(1), content: '# 本文1' });
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));
    // 一覧取得だけでは本文取得(get)は呼ばれない。
    expect(materialMocks.get).not.toHaveBeenCalled();

    act(() => result.current.selectMaterial(1));
    await waitFor(() => expect(result.current.selected?.content).toBe('# 本文1'));
    expect(materialMocks.get).toHaveBeenCalledWith(1);
    expect(materialMocks.get).toHaveBeenCalledTimes(1);

    // 再選択しても キャッシュ済みなら get は増えない。
    act(() => result.current.selectMaterial(2));
    act(() => result.current.selectMaterial(1));
    await waitFor(() => expect(result.current.selected?.id).toBe(1));
    expect(materialMocks.get).toHaveBeenCalledTimes(2); // 1 と 2 の計2回（1 の再取得は無し）
  });

  it('別の章を選択すると前章の取得エラーはクリアされる（取得中ローディングに戻せる）', async () => {
    courseMocks.listMaterials.mockResolvedValue([sample(1), sample(2)]);
    materialMocks.get
      .mockRejectedValueOnce(new Error('boom'))
      .mockResolvedValueOnce({ ...sample(2), content: '# 本文2' });
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.selectMaterial(1));
    await waitFor(() => expect(result.current.error).toBe('教材の取得に失敗しました'));

    // 別章へ切り替えた瞬間に error はクリアされ、新章の本文取得が走る。
    act(() => result.current.selectMaterial(2));
    expect(result.current.error).toBeNull();
    await waitFor(() => expect(result.current.selected?.content).toBe('# 本文2'));
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
    materialMocks.get.mockResolvedValue(sample(1));
    materialMocks.remove.mockResolvedValue(undefined);
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.selectMaterial(1));
    await waitFor(() => expect(result.current.selected?.id).toBe(1));
    await act(async () => {
      await result.current.remove(1);
    });
    expect(result.current.materials).toHaveLength(0);
    expect(result.current.selectedId).toBeNull();
  });
});
