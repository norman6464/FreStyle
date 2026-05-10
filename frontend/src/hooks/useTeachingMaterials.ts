import { useCallback, useEffect, useMemo, useState } from 'react';
import CourseRepository from '../repositories/CourseRepository';
import TeachingMaterialRepository, {
  type TeachingMaterialCreatePayload,
  type TeachingMaterialUpdatePayload,
} from '../repositories/TeachingMaterialRepository';
import type { TeachingMaterial } from '../types';

/**
 * useTeachingMaterials — 指定コース配下の教材一覧 + 選択 + CRUD の状態管理。
 *
 * `courseId` が 0 / null なら何もせず空配列を保持する（routing で `/courses/:id` に
 * 入った瞬間にコース ID が確定する想定）。 autosave は別 hook (useTeachingMaterialEditor) に分離。
 */
export function useTeachingMaterials(courseId: number | null) {
  const [materials, setMaterials] = useState<TeachingMaterial[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchAll = useCallback(async () => {
    if (!courseId) {
      setMaterials([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const rows = await CourseRepository.listMaterials(courseId);
      setMaterials(rows);
    } catch {
      setError('教材の取得に失敗しました');
    } finally {
      setLoading(false);
    }
  }, [courseId]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const selected = useMemo(
    () => materials.find((m) => m.id === selectedId) ?? null,
    [materials, selectedId],
  );

  const filtered = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return materials;
    return materials.filter(
      (m) => m.title.toLowerCase().includes(q) || m.content.toLowerCase().includes(q),
    );
  }, [materials, searchQuery]);

  const create = useCallback(
    async (initial: Omit<TeachingMaterialCreatePayload, 'courseId'>): Promise<TeachingMaterial | null> => {
      if (!courseId) {
        setError('コースが選択されていません');
        return null;
      }
      try {
        const created = await TeachingMaterialRepository.create({ ...initial, courseId });
        setMaterials((prev) => [...prev, created].sort(byOrderThenId));
        setSelectedId(created.id);
        return created;
      } catch {
        setError('教材の作成に失敗しました');
        return null;
      }
    },
    [courseId],
  );

  const update = useCallback(
    async (id: number, payload: TeachingMaterialUpdatePayload): Promise<void> => {
      try {
        const updated = await TeachingMaterialRepository.update(id, payload);
        setMaterials((prev) => prev.map((m) => (m.id === id ? updated : m)).sort(byOrderThenId));
      } catch {
        setError('教材の更新に失敗しました');
      }
    },
    [],
  );

  const remove = useCallback(
    async (id: number): Promise<void> => {
      try {
        await TeachingMaterialRepository.remove(id);
        setMaterials((prev) => prev.filter((m) => m.id !== id));
        if (selectedId === id) setSelectedId(null);
      } catch {
        setError('教材の削除に失敗しました');
      }
    },
    [selectedId],
  );

  return {
    materials,
    filtered,
    selectedId,
    selected,
    loading,
    error,
    searchQuery,
    setSearchQuery,
    selectMaterial: setSelectedId,
    fetchAll,
    create,
    update,
    remove,
  };
}

function byOrderThenId(a: TeachingMaterial, b: TeachingMaterial): number {
  if (a.orderInCourse !== b.orderInCourse) return a.orderInCourse - b.orderInCourse;
  return a.id - b.id;
}
