import { findCourseCategory } from '@/entities/course';

/**
 * コース領域の URL slug と内部 key の相互変換（FRESTYLE-177）。
 *
 * 選択画面（CourseCategorySelectPage）と領域一覧（CoursesListPage）で共有する。
 */

/** URL で未分類を表す予約語。内部 key の ''（空）はパスに使えないため。 */
export const UNCATEGORIZED_SLUG = 'uncategorized';

/** 未分類('')や未知のカテゴリ値を未分類バケットの key('') に正規化する。 */
export function normalizeCategoryKey(category: string): string {
  return findCourseCategory(category) ? category : '';
}
