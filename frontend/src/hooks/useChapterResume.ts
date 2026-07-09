import { useEffect, useRef } from 'react';
import CourseRepository from '../repositories/CourseRepository';
import type { TeachingMaterial } from '../types';

interface UseChapterResumeParams {
  /** false（管理ロール等）のときは何もしない。 */
  enabled: boolean;
  courseId: number | null;
  /** 章一覧（orderInCourse 順）。先頭がフォールバック先。 */
  materials: TeachingMaterial[];
  /** 章一覧の取得中フラグ。ロード完了後に一度だけ発火する。 */
  loading: boolean;
  selectedId: number | null;
  selectMaterial: (id: number | null) => void;
}

/**
 * useChapterResume — コース詳細を開いたとき、最後に閲覧した章（無ければ先頭の章）を
 * 一度だけ自動選択する（FRESTYLE-99「続きから表示」）。
 *
 * - 手動で章を選択済みの場合や、取得完了前に手動選択された場合は上書きしない
 * - 取得中に別コースへ切り替わった場合、遅れて届いた応答では選択しない（stale ガード）
 * - 履歴の取得に失敗しても先頭の章へフォールバックする（レジュームは補助機能）
 */
export function useChapterResume({
  enabled,
  courseId,
  materials,
  loading,
  selectedId,
  selectMaterial,
}: UseChapterResumeParams) {
  // コースごとに一度だけ発火させる。手動選択済みの場合も発火済み扱いにする。
  const resumedCourseRef = useRef<number | null>(null);
  // 非同期解決時に「まだ未選択か」「まだ同じコースか」を最新値で判定するための ref。
  const selectedIdRef = useRef<number | null>(selectedId);
  const courseIdRef = useRef<number | null>(courseId);

  useEffect(() => {
    selectedIdRef.current = selectedId;
  }, [selectedId]);

  useEffect(() => {
    courseIdRef.current = courseId;
  }, [courseId]);

  useEffect(() => {
    if (!enabled || courseId == null || loading) return;
    // コース切替直後は前コースの一覧が残っている commit があるため、
    // 一覧が「今の courseId のもの」になってから発火する。
    if (materials.length === 0 || materials[0].courseId !== courseId) return;
    if (resumedCourseRef.current === courseId) return;
    resumedCourseRef.current = courseId;
    if (selectedIdRef.current != null) return; // すでに（手動）選択済み

    const requestedCourseId = courseId;
    const fallbackId = materials[0].id;
    const selectIfStillRelevant = (id: number) => {
      // 応答待ちの間に別コースへ切り替わった / ユーザーが手動選択した場合は上書きしない。
      if (courseIdRef.current === requestedCourseId && selectedIdRef.current == null) {
        selectMaterial(id);
      }
    };
    CourseRepository.lastViewed(requestedCourseId)
      .then((view) => {
        // 履歴の章が非公開化等で一覧に無い場合も先頭へフォールバック。
        const target =
          view && materials.some((m) => m.id === view.teachingMaterialId)
            ? view.teachingMaterialId
            : fallbackId;
        selectIfStillRelevant(target);
      })
      .catch(() => {
        selectIfStillRelevant(fallbackId);
      });
  }, [enabled, courseId, loading, materials, selectMaterial]);
}
