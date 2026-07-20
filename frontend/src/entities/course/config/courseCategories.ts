/**
 * コースの学習領域カテゴリと表示色の対応（FRESTYLE-67）。
 *
 * 「色＝学習領域」の連想を一貫させるため、カテゴリは backend の
 * domain.ValidCourseCategories と 1:1 の固定リストで、色もここで固定する。
 * アクセシビリティ配慮のため色だけに依存せず、必ず label をバッジ文字列として
 * 併記して使うこと（色覚特性・スクリーンリーダー対応）。
 */
export interface CourseCategoryDef {
  /** backend に保存される値（domain.CourseCategory* と一致） */
  key: string;
  /** バッジに表示する日本語名 */
  label: string;
  /** カテゴリバッジの配色 */
  badgeClass: string;
  /** カード左端の色帯（border-l-4 と併用） */
  barClass: string;
}

export const COURSE_CATEGORIES: CourseCategoryDef[] = [
  {
    key: 'dev-basics',
    label: '開発基礎',
    badgeClass: 'bg-amber-500/15 text-amber-700 border border-amber-500/30',
    barClass: 'border-l-amber-500',
  },
  {
    key: 'backend',
    label: 'バックエンド開発',
    badgeClass: 'bg-blue-500/15 text-blue-700 border border-blue-500/30',
    barClass: 'border-l-blue-500',
  },
  {
    key: 'architecture',
    label: '設計・アーキテクチャ',
    badgeClass: 'bg-violet-500/15 text-violet-700 border border-violet-500/30',
    barClass: 'border-l-violet-500',
  },
  {
    key: 'database',
    label: 'データベース',
    badgeClass: 'bg-emerald-500/15 text-emerald-700 border border-emerald-500/30',
    barClass: 'border-l-emerald-500',
  },
  {
    key: 'infra',
    label: 'インフラ・クラウド',
    badgeClass: 'bg-orange-500/15 text-orange-700 border border-orange-500/30',
    barClass: 'border-l-orange-500',
  },
  {
    key: 'security',
    label: 'セキュリティ',
    badgeClass: 'bg-rose-500/15 text-rose-700 border border-rose-500/30',
    barClass: 'border-l-rose-500',
  },
  {
    key: 'product',
    label: 'プロダクト・仕様',
    badgeClass: 'bg-cyan-500/15 text-cyan-700 border border-cyan-500/30',
    barClass: 'border-l-cyan-500',
  },
  {
    key: 'design',
    label: 'デザインパターン',
    badgeClass: 'bg-pink-500/15 text-pink-700 border border-pink-500/30',
    barClass: 'border-l-pink-500',
  },
];

/** key からカテゴリ定義を引く。未分類('')や未知の値は undefined（無色表示）。 */
export function findCourseCategory(key: string | undefined): CourseCategoryDef | undefined {
  if (!key) return undefined;
  return COURSE_CATEGORIES.find((c) => c.key === key);
}
