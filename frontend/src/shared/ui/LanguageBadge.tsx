import { languageBadgeClass } from '@/shared/config/languageBadgeClasses';

interface LanguageBadgeProps {
  /** 言語・技術の値（例: 'docker' / 'go' / 'terraform'）。 */
  language: string;
  /** 等幅フォントで表示する（詳細ヘッダ用）。 */
  mono?: boolean;
  className?: string;
}

// 未知の言語は従来どおり無彩色にフォールバックする。
const FALLBACK = 'bg-surface-3 text-[var(--color-text-muted)] border-transparent';

/** 言語・技術を識別色付きのバッジで表示する（演習・コース共用）。 */
export default function LanguageBadge({ language, mono = false, className = '' }: LanguageBadgeProps) {
  // 全大文字(TYPESCRIPT)は圧が強いというユーザー要望で、先頭のみ大文字の表記にする(FRESTYLE-121)。
  const label = language ? language.charAt(0).toUpperCase() + language.slice(1).toLowerCase() : language;
  return (
    <span
      className={`inline-flex items-center rounded border px-1.5 py-0.5 text-xs font-medium tracking-wide ${
        mono ? 'font-mono' : ''
      } ${languageBadgeClass(language) ?? FALLBACK} ${className}`}
    >
      {label}
    </span>
  );
}
