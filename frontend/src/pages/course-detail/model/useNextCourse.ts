import { useEffect, useState } from 'react';
import { CourseRepository } from '@/entities/course';
import type { CourseWithProgress } from '@/entities/course';

/**
 * useNextCourse — 現在のコースの「次のコース」を返す(FRESTYLE-102)。
 *
 * 並び順は backend のコース一覧 API と同じ(sort_order 昇順 → id 昇順)。
 * 最終章の末尾から一覧に戻らず次のコースへ直行する導線に使う。
 * 補助導線なので、取得失敗・自分が並び順で最後・一覧に見つからない場合は null(導線を出さないだけ)。
 * `enabled=false`(管理ロール等)のときは API を叩かない。
 */
export function useNextCourse(courseId: number | null, enabled: boolean) {
  const [nextCourse, setNextCourse] = useState<CourseWithProgress | null>(null);

  useEffect(() => {
    if (!enabled || courseId == null) {
      setNextCourse(null);
      return;
    }
    let active = true;
    // コース切替・権限切替の直後に前回の「次のコース」が一瞬残らないよう、再取得開始時にクリアする。
    setNextCourse(null);
    CourseRepository.list()
      .then((rows) => {
        if (!active) return;
        // backend が 0 件時に null を返す事故に備えた防御(FRESTYLE-70 と同じ理由)。
        const list = rows ?? [];
        const idx = list.findIndex((c) => c.id === courseId);
        setNextCourse(idx >= 0 && idx < list.length - 1 ? list[idx + 1] : null);
      })
      .catch(() => {
        // 補助導線なので、取得失敗時は導線を出さないまま黙って続行する。
        if (active) setNextCourse(null);
      });
    return () => {
      active = false;
    };
  }, [courseId, enabled]);

  return { nextCourse };
}
