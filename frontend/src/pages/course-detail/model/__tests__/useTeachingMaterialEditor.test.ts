import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTeachingMaterialEditor } from '../useTeachingMaterialEditor';
import type { TeachingMaterial } from '@/entities/course';

const sample = (id: number, content = ''): TeachingMaterial => ({
  id,
  companyId: 1,
  courseId: 5,
  createdByUserId: 1,
  title: `教材${id}`,
  content,
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
        initialProps: { selectedId: 5 as number | null, selected: sample(5, '初期コンテンツ') as TeachingMaterial | null },
      },
    );

    expect(result.current.editContent).toBe('初期コンテンツ');

    // ユーザがエディタで入力した想定
    act(() => result.current.handleContentChange('入力中の差分'));
    expect(result.current.editContent).toBe('入力中の差分');

    // autosave 後の materials 再 fetch で selected が新しい参照に置き換わる。
    // ただし selectedId は同じ 5 なので editor 状態を上書きしてはいけない。
    rerender({
      selectedId: 5,
      selected: sample(5, '初期コンテンツ'), // 新しい object reference
    });
    expect(result.current.editContent).toBe('入力中の差分'); // 上書きされない

    // 別の教材 7 を選択 → 今度は state を入れ替える
    rerender({
      selectedId: 7,
      selected: sample(7, '別教材のコンテンツ'),
    });
    expect(result.current.editContent).toBe('別教材のコンテンツ');
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
    expect(result.current.editContent).toBe('X');

    rerender({ selectedId: null, selected: null });
    expect(result.current.editContent).toBe('');
    expect(result.current.editTitle).toBe('');
  });
});
