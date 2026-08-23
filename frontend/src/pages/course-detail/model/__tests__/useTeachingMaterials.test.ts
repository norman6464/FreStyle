import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useTeachingMaterials } from '../useTeachingMaterials';
import { CourseRepository } from '@/entities/course';
import { TeachingMaterialRepository } from '@/entities/course';
import type { RichDocContent } from '@/shared/ui/RichTextEditor';

vi.mock('@/entities/course/api/courseRepository', () => ({
  default: {
    list: vi.fn(),
    get: vi.fn(),
    listMaterials: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
}));

vi.mock('@/entities/course/api/teachingMaterialRepository', () => ({
  default: {
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
}));

const courseMocks = vi.mocked(CourseRepository);
const materialMocks = vi.mocked(TeachingMaterialRepository);

const doc = (text: string): RichDocContent => ({
  type: 'doc',
  content: [{ type: 'paragraph', content: [{ type: 'text', text }] }],
});

const sample = (
  id: number,
  overrides: Partial<{ title: string; isPublished: boolean; orderInCourse: number }> = {},
) => ({
  id,
  companyId: 1,
  courseId: 5,
  createdByUserId: 1,
  title: overrides.title ?? `教材${id}`,
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

  it('選択した教材だけ本文(doc)を都度取得して selected に入る（全章は先読みしない）', async () => {
    courseMocks.listMaterials.mockResolvedValue([sample(1), sample(2)]);
    materialMocks.get.mockResolvedValue({ ...sample(1), doc: doc('本文1'), revision: 1 });
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));
    // 一覧取得だけでは本文取得(get)は呼ばれない。
    expect(materialMocks.get).not.toHaveBeenCalled();

    act(() => result.current.selectMaterial(1));
    await waitFor(() => expect(result.current.selected?.doc).toEqual(doc('本文1')));
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
      .mockResolvedValueOnce({ ...sample(2), doc: doc('本文2'), revision: 1 });
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.selectMaterial(1));
    await waitFor(() => expect(result.current.error).toBe('教材の取得に失敗しました'));

    // 別章へ切り替えた瞬間に error はクリアされ、新章の本文取得が走る。
    act(() => result.current.selectMaterial(2));
    expect(result.current.error).toBeNull();
    await waitFor(() => expect(result.current.selected?.doc).toEqual(doc('本文2')));
  });

  it('取得失敗後に同じ章を選び直すと再取得が走る（ローディングで固まらない）', async () => {
    courseMocks.listMaterials.mockResolvedValue([sample(1)]);
    materialMocks.get
      .mockRejectedValueOnce(new Error('boom'))
      .mockResolvedValueOnce({ ...sample(1), doc: doc('本文1'), revision: 1 });
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.selectMaterial(1));
    await waitFor(() => expect(result.current.error).toBe('教材の取得に失敗しました'));

    // 同じ章を選び直す → selectedId は不変だが selectionSeq が進むので effect が再実行される。
    act(() => result.current.selectMaterial(1));
    expect(result.current.error).toBeNull();
    await waitFor(() => expect(result.current.selected?.doc).toEqual(doc('本文1')));
    expect(materialMocks.get).toHaveBeenCalledTimes(2);
  });

  it('create 成功時にリストに追加され selected になる（一覧 state に doc は持たせない）', async () => {
    courseMocks.listMaterials.mockResolvedValue([]);
    materialMocks.create.mockResolvedValue({
      ...sample(99, { orderInCourse: 100 }),
      doc: doc('新規章の本文'),
      revision: 1,
    });
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));
    await act(async () => {
      await result.current.create({
        title: '新',
        orderInCourse: 100,
        isPublished: false,
      });
    });
    expect(materialMocks.create).toHaveBeenCalledWith(expect.objectContaining({ courseId: 5 }));
    expect(result.current.materials).toHaveLength(1);
    expect(result.current.materials[0].id).toBe(99);
    // 一覧はメタデータのみ（doc / revision は詳細キャッシュ側だけが持つ）。
    expect(result.current.materials[0].doc).toBeUndefined();
    expect(result.current.materials[0].revision).toBeUndefined();
    expect(result.current.selectedId).toBe(99);
    expect(result.current.selected?.doc).toEqual(doc('新規章の本文'));
  });

  it('syncDetail は詳細キャッシュへ doc を反映しつつ、一覧 state には doc を混ぜない', async () => {
    courseMocks.listMaterials.mockResolvedValue([sample(1)]);
    materialMocks.get.mockResolvedValue({ ...sample(1), doc: doc('本文1'), revision: 1 });
    const { result } = renderHook(() => useTeachingMaterials(5));
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.selectMaterial(1));
    await waitFor(() => expect(result.current.selected?.doc).toEqual(doc('本文1')));

    // doc 保存（PUT /doc）の応答で最新化された想定。
    act(() => result.current.syncDetail({ ...sample(1), doc: doc('保存後の本文'), revision: 2 }));
    expect(result.current.selected?.doc).toEqual(doc('保存後の本文'));
    expect(result.current.selected?.revision).toBe(2);
    expect(result.current.materials[0].doc).toBeUndefined();
    expect(result.current.materials[0].revision).toBeUndefined();
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
