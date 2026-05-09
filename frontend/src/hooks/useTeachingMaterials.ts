import { useCallback, useEffect, useMemo, useState } from 'react';
import TeachingMaterialRepository, {
  type TeachingMaterialPayload,
} from '../repositories/TeachingMaterialRepository';
import type { TeachingMaterial } from '../types';

/**
 * useTeachingMaterials — 教材一覧 + 選択 + CRUD の状態管理。
 *
 * NotesPage の useNotes と同じ「左パネル一覧 + 右パネル詳細」構成を想定。
 * autosave は別 hook (useTeachingMaterialEditor) に分離する。
 */
export function useTeachingMaterials() {
  const [materials, setMaterials] = useState<TeachingMaterial[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchAll = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const rows = await TeachingMaterialRepository.list();
      setMaterials(rows);
    } catch {
      setError('教材の取得に失敗しました');
    } finally {
      setLoading(false);
    }
  }, []);

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

  const create = useCallback(async (initial: TeachingMaterialPayload): Promise<TeachingMaterial | null> => {
    try {
      const created = await TeachingMaterialRepository.create(initial);
      setMaterials((prev) => [created, ...prev]);
      setSelectedId(created.id);
      return created;
    } catch {
      setError('教材の作成に失敗しました');
      return null;
    }
  }, []);

  const update = useCallback(
    async (id: number, payload: TeachingMaterialPayload): Promise<void> => {
      try {
        const updated = await TeachingMaterialRepository.update(id, payload);
        setMaterials((prev) => prev.map((m) => (m.id === id ? updated : m)));
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
