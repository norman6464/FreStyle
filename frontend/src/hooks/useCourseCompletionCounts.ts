import { useEffect, useState } from 'react';
import LessonProgressRepository from '../repositories/LessonProgressRepository';

/**
 * useCourseCompletionCounts — current user の完了章数をコース別に集計する。
 * コース一覧のカードに進捗(完了章数/全章数)を出すための hook(FRESTYLE-98)。
 *
 * `enabled=false`(company_admin / super_admin など進捗記録を持たないロール)のときは
 * API を叩かず空 Map を保つ。取得失敗時も空 Map のまま黙って続行する(閲覧を阻害しない)。
 */
export function useCourseCompletionCounts(enabled: boolean) {
  const [countByCourse, setCountByCourse] = useState<Map<number, number>>(new Map());
  const [loading, setLoading] = useState(enabled);

  useEffect(() => {
    if (!enabled) {
      setCountByCourse(new Map());
      setLoading(false);
      return;
    }
    let active = true;
    setLoading(true);
    LessonProgressRepository.list()
      .then((rows) => {
        if (!active) return;
        const counts = new Map<number, number>();
        // backend が 0 件時に null を返す事故に備えた防御(FRESTYLE-70 と同じ理由)。
        for (const r of rows ?? []) {
          counts.set(r.courseId, (counts.get(r.courseId) ?? 0) + 1);
        }
        setCountByCourse(counts);
      })
      .catch(() => {
        // 進捗は補助情報なので、取得失敗時は空のまま黙って続行する。
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [enabled]);

  return { countByCourse, loading };
}
