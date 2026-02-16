/**
 * ソートオプション定数
 *
 * SortSelector と usePracticePage で共有
 */

export type SortOption = 'default' | 'difficulty-asc' | 'difficulty-desc' | 'name';

export interface SortOptionItem {
  value: SortOption;
  label: string;
}

export const SORT_OPTIONS: SortOptionItem[] = [
  { value: 'default', label: 'デフォルト' },
  { value: 'difficulty-asc', label: '難易度↑' },
  { value: 'difficulty-desc', label: '難易度↓' },
  { value: 'name', label: '名前順' },
];

export const DIFFICULTY_ORDER: Record<string, number> = {
  beginner: 1,
  intermediate: 2,
  advanced: 3,
};

export type NoteSortOption = 'default' | 'updated-asc' | 'title' | 'created-desc';

export interface NoteSortOptionItem {
  value: NoteSortOption;
  label: string;
}

export const NOTE_SORT_OPTIONS: NoteSortOptionItem[] = [
  { value: 'default', label: 'デフォルト' },
  { value: 'updated-asc', label: '古い順' },
  { value: 'title', label: 'タイトル順' },
  { value: 'created-desc', label: '作成日（新しい順）' },
];
