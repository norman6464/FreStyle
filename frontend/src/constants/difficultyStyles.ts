/**
 * 難易度スタイル定数
 *
 * DifficultyFilter と DailyChallengeCard で共有
 */

export interface DifficultyOption {
  value: string | null;
  label: string;
  style: string;
}

const DIFFICULTY_BADGE_STYLE = 'border border-[var(--color-text-muted)] text-[var(--color-text-primary)]';

export const DIFFICULTY_STYLES: Record<string, string> = {
  '初級': DIFFICULTY_BADGE_STYLE,
  '中級': DIFFICULTY_BADGE_STYLE,
  '上級': DIFFICULTY_BADGE_STYLE,
};

export const DIFFICULTY_OPTIONS: DifficultyOption[] = [
  { value: null, label: '全レベル', style: 'text-[var(--color-text-secondary)] bg-surface-2' },
  { value: '初級', label: '初級', style: 'text-emerald-400 bg-emerald-900/30' },
  { value: '中級', label: '中級', style: 'text-amber-400 bg-amber-900/30' },
  { value: '上級', label: '上級', style: 'text-rose-400 bg-rose-900/30' },
];
