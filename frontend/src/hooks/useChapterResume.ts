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
  // 非同期解決時に「まだ未選択か」を最新値で判定するための ref。
  const selectedIdRef = useRef<number | null>(selectedId);
  // 発火判定時に最新の一覧を参照するための ref（materials の identity 変化で再発火させない）。
  const materialsRef = useRef<TeachingMaterial[]>(materials);

  useEffect(() => {
    selectedIdRef.current = selectedId;
  }, [selectedId]);

  useEffect(() => {
    materialsRef.current = materials;
  }, [materials]);

  useEffect(() => {
    if (!enabled || courseId == null || loading) return;
    const mats = materialsRef.current;
    // コース切替直後は前コースの一覧が残っている commit があるため、
    // 一覧が「今の courseId のもの」になってから発火する。
    if (mats.length === 0 || mats[0].courseId !== courseId) return;
    if (resumedCourseRef.current === courseId) return;
    resumedCourseRef.current = courseId;
    if (selectedIdRef.current != null) return; // すでに（手動）選択済み

    const fallbackId = mats[0].id;
    const selectIfStillUnselected = (id: number) => {
      // 履歴取得中にユーザーが手動選択していたら上書きしない。
      if (selectedIdRef.current == null) selectMaterial(id);
    };
    CourseRepository.lastViewed(courseId)
      .then((view) => {
        // 履歴の章が非公開化等で一覧に無い場合も先頭へフォールバック。
        const target =
          view && mats.some((m) => m.id === view.teachingMaterialId)
            ? view.teachingMaterialId
            : fallbackId;
        selectIfStillUnselected(target);
      })
      .catch(() => {
        selectIfStillUnselected(fallbackId);
      });
  }, [enabled, courseId, loading, materials, selectMaterial]);
}
