import { findExerciseLanguage } from '../constants/exerciseLanguages';

interface LanguageBadgeProps {
  /** 演習の言語値（例: 'docker' / 'go'）。 */
  language: string;
  /** 等幅フォントで表示する（詳細ヘッダ用）。 */
  mono?: boolean;
  className?: string;
}

// 未知の言語は従来どおり無彩色にフォールバックする。
const FALLBACK = 'bg-surface-3 text-[var(--color-text-muted)] border-transparent';

/** 演習の言語を、言語ごとの識別色を付けたバッジで表示する。 */
export default function LanguageBadge({ language, mono = false, className = '' }: LanguageBadgeProps) {
  const def = findExerciseLanguage(language);
  return (
    <span
      className={`inline-flex items-center rounded border px-1.5 py-0.5 text-xs font-medium uppercase tracking-wide ${
        mono ? 'font-mono' : ''
      } ${def?.badgeClass ?? FALLBACK} ${className}`}
    >
      {language}
    </span>
  );
}
