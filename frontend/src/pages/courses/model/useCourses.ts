import { useCallback, useEffect, useMemo, useState } from 'react';
import { CourseRepository, type CoursePayload } from '@/entities/course';
import type { Course, CourseWithProgress } from '@/entities/course';

/**
 * useCourses — コース一覧 + 検索 + CRUD の状態管理。
 *
 * trainee は published のみ、 company_admin / super_admin は draft 含む全件を取得する
 * （フィルタは backend 側）。
 */
export function useCourses() {
  const [courses, setCourses] = useState<CourseWithProgress[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchAll = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const rows = await CourseRepository.list();
      // backend が 0 件時に null を返す事故に備えた防御(FRESTYLE-70)。
      setCourses(rows ?? []);
    } catch {
      setError('コースの取得に失敗しました');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const filtered = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return courses;
    return courses.filter(
      (c) => c.title.toLowerCase().includes(q) || c.description.toLowerCase().includes(q),
    );
  }, [courses, searchQuery]);

  const create = useCallback(async (initial: CoursePayload): Promise<Course | null> => {
    try {
      const created = await CourseRepository.create(initial);
      // 作成 API は進捗フィールドを返さないが、新規コースの章数・完了数は必ず 0。
      setCourses((prev) =>
        [...prev, { ...created, materialCount: 0, completedCount: 0 }].sort(byOrderThenId),
      );
      return created;
    } catch {
      setError('コースの作成に失敗しました');
      return null;
    }
  }, []);

  const update = useCallback(
    async (id: number, payload: CoursePayload): Promise<Course | null> => {
      try {
        const updated = await CourseRepository.update(id, payload);
        // 更新 API は進捗フィールドを返さないため、取得済みの値を引き継ぐ。
        setCourses((prev) =>
          prev
            .map((c) =>
              c.id === id
                ? { ...updated, materialCount: c.materialCount, completedCount: c.completedCount }
                : c,
            )
            .sort(byOrderThenId),
        );
        return updated;
      } catch {
        setError('コースの更新に失敗しました');
        return null;
      }
    },
    [],
  );

  const remove = useCallback(async (id: number): Promise<void> => {
    try {
      await CourseRepository.remove(id);
      setCourses((prev) => prev.filter((c) => c.id !== id));
    } catch {
      setError('コースの削除に失敗しました');
    }
  }, []);

  return {
    courses,
    filtered,
    loading,
    error,
    searchQuery,
    setSearchQuery,
    fetchAll,
    create,
    update,
    remove,
  };
}

// 並び順は backend 側と同じ: sort_order 昇順 → id 昇順
function byOrderThenId(a: Course, b: Course): number {
  if (a.sortOrder !== b.sortOrder) return a.sortOrder - b.sortOrder;
  return a.id - b.id;
}
