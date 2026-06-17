import { useCallback, useEffect, useMemo, useState } from 'react';
import CourseRepository from '../repositories/CourseRepository';
import TeachingMaterialRepository, {
  type TeachingMaterialCreatePayload,
  type TeachingMaterialUpdatePayload,
} from '../repositories/TeachingMaterialRepository';
import type { TeachingMaterial } from '../types';

/**
 * useTeachingMaterials — 指定コース配下の教材の「一覧(メタデータ) + 選択 + CRUD」の状態管理。
 *
 * 効率化のため一覧は本文(content)を含まない軽量メタデータで取得し、 選択された教材の本文だけ
 * `GetByID` で都度取得してキャッシュする（全章を先読みしない）。autosave は useTeachingMaterialEditor。
 */
export function useTeachingMaterials(courseId: number | null) {
  // materials は content を持たないメタデータ一覧。 detailCache は本文込みの取得済み教材。
  const [materials, setMaterials] = useState<TeachingMaterial[]>([]);
  const [detailCache, setDetailCache] = useState<Record<number, TeachingMaterial>>({});
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [detailLoading, setDetailLoading] = useState(false);
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

  // コースが変わったら選択 / 本文キャッシュをリセットする。
  useEffect(() => {
    setSelectedId(null);
    setDetailCache({});
  }, [courseId]);

  // 選択された教材の本文を都度取得してキャッシュする（未取得時のみ）。
  useEffect(() => {
    if (selectedId == null || detailCache[selectedId]) return;
    let active = true;
    setDetailLoading(true);
    TeachingMaterialRepository.get(selectedId)
      .then((m) => {
        if (active) setDetailCache((prev) => ({ ...prev, [m.id]: m }));
      })
      .catch(() => {
        if (active) setError('教材の取得に失敗しました');
      })
      .finally(() => {
        if (active) setDetailLoading(false);
      });
    return () => {
      active = false;
    };
  }, [selectedId, detailCache]);

  // selected は本文込みのキャッシュから返す（未取得なら null = 取得中）。
  const selected = selectedId != null ? (detailCache[selectedId] ?? null) : null;

  // 章を選ぶときは前章で出た取得エラーを先にクリアする。 こうしないと別章へ切り替えても
  // 古い error が残り、「取得中ローディング」ではなくエラー扱いのまま表示されてしまう。
  const selectMaterial = useCallback((id: number | null) => {
    setError(null);
    setSelectedId(id);
  }, []);

  const filtered = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return materials;
    // 一覧は本文を持たないためタイトル検索のみ（本文全文検索は別途 backend 検索で対応）。
    return materials.filter((m) => m.title.toLowerCase().includes(q));
  }, [materials, searchQuery]);

  const create = useCallback(
    async (initial: Omit<TeachingMaterialCreatePayload, 'courseId'>): Promise<TeachingMaterial | null> => {
      if (!courseId) {
        setError('コースが選択されていません');
        return null;
      }
      try {
        const created = await TeachingMaterialRepository.create({ ...initial, courseId });
        setMaterials((prev) => [...prev, stripContent(created)].sort(byOrderThenId));
        setDetailCache((prev) => ({ ...prev, [created.id]: created }));
        setSelectedId(created.id);
        return created;
      } catch {
        setError('教材の作成に失敗しました');
        return null;
      }
    },
    [courseId],
  );

  const update = useCallback(async (id: number, payload: TeachingMaterialUpdatePayload): Promise<void> => {
    try {
      const updated = await TeachingMaterialRepository.update(id, payload);
      setMaterials((prev) => prev.map((m) => (m.id === id ? stripContent(updated) : m)).sort(byOrderThenId));
      setDetailCache((prev) => ({ ...prev, [id]: updated }));
    } catch {
      setError('教材の更新に失敗しました');
    }
  }, []);

  const remove = useCallback(
    async (id: number): Promise<void> => {
      try {
        await TeachingMaterialRepository.remove(id);
        setMaterials((prev) => prev.filter((m) => m.id !== id));
        setDetailCache((prev) => {
          const next = { ...prev };
          delete next[id];
          return next;
        });
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
    detailLoading,
    error,
    searchQuery,
    setSearchQuery,
    selectMaterial,
    fetchAll,
    create,
    update,
    remove,
  };
}

// 一覧(メタデータ)に本文を持たせない（先読み抑制 + 一貫性のため content を空にする）。
function stripContent(m: TeachingMaterial): TeachingMaterial {
  return { ...m, content: '' };
}

function byOrderThenId(a: TeachingMaterial, b: TeachingMaterial): number {
  if (a.orderInCourse !== b.orderInCourse) return a.orderInCourse - b.orderInCourse;
  return a.id - b.id;
}
