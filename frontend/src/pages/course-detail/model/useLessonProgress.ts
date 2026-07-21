import { useCallback, useEffect, useRef, useState } from 'react';
import { LessonProgressRepository } from '@/entities/course';

/**
 * useLessonProgress — current user の教材（レッスン）完了状態を管理する。
 *
 * `enabled=false`（教材を管理する company_admin / super_admin など、 学習者ではないロール）の
 * ときは API を叩かず空集合を保つ。 完了トグルは楽観的更新し、 失敗時は元の状態へロールバックする。
 */
export function useLessonProgress(enabled: boolean) {
  const [completedIds, setCompletedIds] = useState<Set<number>>(new Set());
  const [loading, setLoading] = useState(enabled);
  // 同一教材への連打（冒頭/末尾の 2 ボタン）で toggle が並行し、
  // 失敗ロールバックが完了順依存でズレるのを防ぐ in-flight ガード。
  const inFlightRef = useRef<Set<number>>(new Set());

  useEffect(() => {
    if (!enabled) {
      setCompletedIds(new Set());
      setLoading(false);
      return;
    }
    let active = true;
    setLoading(true);
    LessonProgressRepository.list()
      .then((rows) => {
        if (!active) return;
        setCompletedIds(new Set(rows.map((r) => r.teachingMaterialId)));
      })
      .catch(() => {
        // 進捗は補助情報なので、 取得失敗時は空のまま黙って続行する（教材閲覧は阻害しない）。
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [enabled]);

  const isCompleted = useCallback(
    (materialId: number) => completedIds.has(materialId),
    [completedIds],
  );

  const setLocal = useCallback((materialId: number, done: boolean) => {
    setCompletedIds((prev) => {
      const next = new Set(prev);
      if (done) next.add(materialId);
      else next.delete(materialId);
      return next;
    });
  }, []);

  /** toggle は materialId を done の状態へ更新する。 楽観的更新し、 失敗時は false を返してロールバック。 */
  const toggle = useCallback(
    async (materialId: number, done: boolean): Promise<boolean> => {
      if (inFlightRef.current.has(materialId)) return true; // 同一教材の二重送信を無視
      inFlightRef.current.add(materialId);
      setLocal(materialId, done);
      try {
        if (done) await LessonProgressRepository.complete(materialId);
        else await LessonProgressRepository.incomplete(materialId);
        return true;
      } catch {
        setLocal(materialId, !done);
        return false;
      } finally {
        inFlightRef.current.delete(materialId);
      }
    },
    [setLocal],
  );

  return { completedIds, isCompleted, toggle, loading };
}
