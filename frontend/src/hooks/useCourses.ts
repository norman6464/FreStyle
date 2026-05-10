import { useCallback, useEffect, useMemo, useState } from 'react';
import CourseRepository, {
  type CoursePayload,
} from '../repositories/CourseRepository';
import type { Course } from '../types';

/**
 * useCourses — コース一覧 + 検索 + CRUD の状態管理。
 *
 * trainee は published のみ、 company_admin / super_admin は draft 含む全件を取得する
 * （フィルタは backend 側）。
 */
export function useCourses() {
  const [courses, setCourses] = useState<Course[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchAll = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const rows = await CourseRepository.list();
      setCourses(rows);
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
      setCourses((prev) => [...prev, created].sort(byOrderThenId));
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
        setCourses((prev) => prev.map((c) => (c.id === id ? updated : c)).sort(byOrderThenId));
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
