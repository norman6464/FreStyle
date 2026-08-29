import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTeachingMaterialEditor } from '../useTeachingMaterialEditor';
import type { TeachingMaterial } from '@/entities/course';
import TeachingMaterialRepository from '@/entities/course/api/teachingMaterialRepository';

vi.mock('@/entities/course/api/teachingMaterialRepository', () => ({
  default: { get: vi.fn(), updateDoc: vi.fn() },
}));

const sample = (id: number, title = `教材${id}`): TeachingMaterial => ({
  id,
  companyId: 1,
  courseId: 5,
  createdByUserId: 1,
  title,
  orderInCourse: id * 10,
  isPublished: true,
  createdAt: '',
  updatedAt: '',
});

describe('useTeachingMaterialEditor', () => {
  it('selectedId が変わったときだけ editor 状態を読み直す (autosave 後の re-fetch では上書きしない)', () => {
    const update = vi.fn().mockResolvedValue(undefined);

    // 初回: 教材5 を読み込む
    const { result, rerender } = renderHook(
      ({ selectedId, selected }: { selectedId: number | null; selected: TeachingMaterial | null }) =>
        useTeachingMaterialEditor({ selectedId, selected, update }),
      {
        initialProps: { selectedId: 5 as number | null, selected: sample(5, '初期タイトル') as TeachingMaterial | null },
      },
    );

    expect(result.current.editTitle).toBe('初期タイトル');

    // ユーザがタイトルを入力した想定
    act(() => result.current.handleTitleChange('入力中の差分'));
    expect(result.current.editTitle).toBe('入力中の差分');

    // autosave 後の materials 再 fetch で selected が新しい参照に置き換わる。
    // ただし selectedId は同じ 5 なので editor 状態を上書きしてはいけない。
    rerender({
      selectedId: 5,
      selected: sample(5, '初期タイトル'), // 新しい object reference
    });
    expect(result.current.editTitle).toBe('入力中の差分'); // 上書きされない

    // 別の教材 7 を選択 → 今度は state を入れ替える
    rerender({
      selectedId: 7,
      selected: sample(7, '別教材のタイトル'),
    });
    expect(result.current.editTitle).toBe('別教材のタイトル');
  });

  it('selectedId が null になると editor 状態をクリアする', () => {
    const update = vi.fn().mockResolvedValue(undefined);
    const { result, rerender } = renderHook(
      ({ selectedId, selected }: { selectedId: number | null; selected: TeachingMaterial | null }) =>
        useTeachingMaterialEditor({ selectedId, selected, update }),
      {
        initialProps: { selectedId: 5 as number | null, selected: sample(5, 'X') as TeachingMaterial | null },
      },
    );
    expect(result.current.editTitle).toBe('X');

    rerender({ selectedId: null, selected: null });
    expect(result.current.editTitle).toBe('');
    expect(result.current.editDoc).toEqual({ type: 'doc', content: [{ type: 'paragraph' }] });
  });
});

describe('flushSave（画面を離れる前に送り切る）', () => {
  const doc = { type: 'doc', content: [] } as const;

  it('待っている本文の保存を、遅延を待たずに送る', async () => {
    // 保存は打鍵から 800ms 待って送る。待っている間に画面を離れると
    // タイマーごと捨てられて直近の入力が消えるので、離れる前に送り切る。
    vi.mocked(TeachingMaterialRepository.updateDoc).mockResolvedValue(sample(5));
    const { result } = renderHook(() =>
      useTeachingMaterialEditor({ selectedId: 5, selected: sample(5), update: vi.fn() }),
    );

    act(() => result.current.handleDocChange({ ...doc }));
    expect(TeachingMaterialRepository.updateDoc).not.toHaveBeenCalled();

    await act(async () => {
      result.current.flushSave();
    });

    expect(TeachingMaterialRepository.updateDoc).toHaveBeenCalledTimes(1);
  });

  it('待っている題名の保存も送る', async () => {
    const update = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useTeachingMaterialEditor({ selectedId: 5, selected: sample(5), update }),
    );

    act(() => result.current.handleTitleChange('打ちかけの題名'));
    expect(update).not.toHaveBeenCalled();

    await act(async () => {
      result.current.flushSave();
    });

    expect(update).toHaveBeenCalledWith(5, expect.objectContaining({ title: '打ちかけの題名' }));
  });

  it('待っている保存が無ければ何も送らない（離れるたびに書き込まない）', async () => {
    const update = vi.fn().mockResolvedValue(undefined);
    vi.mocked(TeachingMaterialRepository.updateDoc).mockClear();
    const { result } = renderHook(() =>
      useTeachingMaterialEditor({ selectedId: 5, selected: sample(5), update }),
    );

    await act(async () => {
      result.current.flushSave();
    });

    expect(update).not.toHaveBeenCalled();
    expect(TeachingMaterialRepository.updateDoc).not.toHaveBeenCalled();
  });
});
