/**
 * シナリオラベル定数
 *
 * ScenarioCard・usePracticePage で使用するDB値⇔日本語の変換マッピング
 */

export const DIFFICULTY_LABEL: Record<string, string> = {
  beginner: '初級',
  intermediate: '中級',
  advanced: '上級',
};

export const CATEGORY_LABEL: Record<string, string> = {
  customer: '顧客折衝',
  senior: 'シニア・上司',
  team: 'チーム内',
};

/** 日本語カテゴリ→DB値の逆引き（usePracticePageのフィルタリングで使用） */
export const CATEGORY_LABEL_TO_DB: Record<string, string> = Object.fromEntries(
  Object.entries(CATEGORY_LABEL).map(([k, v]) => [v, k])
);

export const DIFFICULTY_DESCRIPTION: Record<string, string> = {
  beginner: '基本的な報連相',
  intermediate: '利害関係の調整',
  advanced: '複雑な交渉・説得',
};
