import { useState } from 'react';
import { SparklesIcon, XMarkIcon } from '@heroicons/react/24/outline';

type LearningStep = {
  /** ステップのタイトル */
  title: string;
  /** 1 文程度の補足 */
  description: string;
};

type FirstTimeWelcomeProps = {
  /** 挨拶のタイトル */
  title?: string;
  /** 学習ステップ（3〜5 件を推奨） */
  steps: LearningStep[];
  /** 最初のアクションへのラベル */
  primaryActionLabel?: string;
  /** 最初のアクションをクリックしたときのコールバック */
  onPrimaryAction?: () => void;
  /** 閉じた状態を永続化するキー */
  storageKey?: string;
};

/**
 * 初回訪問者向けのウェルカムカード。
 * 「このアプリで何ができるのか」「まず何をすればよいのか」を 3〜5 ステップで示す。
 */
export default function FirstTimeWelcome({
  title = 'ようこそ FreStyle へ',
  steps,
  primaryActionLabel = 'はじめて練習する',
  onPrimaryAction,
  storageKey,
}: FirstTimeWelcomeProps) {
  const [dismissed, setDismissed] = useState<boolean>(() => {
    if (!storageKey || typeof window === 'undefined') return false;
    try {
      return window.localStorage.getItem(storageKey) === 'dismissed';
    } catch {
      return false;
    }
  });

  if (dismissed) return null;

  const handleDismiss = () => {
    setDismissed(true);
    if (storageKey && typeof window !== 'undefined') {
      try {
        window.localStorage.setItem(storageKey, 'dismissed');
      } catch {
        // localStorage 不可の環境では無視
      }
    }
  };

  return (
    <section
      aria-labelledby="first-time-welcome-title"
      className="relative rounded-2xl border border-primary-400/30 bg-gradient-to-br from-primary-500/10 via-surface-1 to-surface-1 p-6 shadow-sm animate-scale-in"
    >
      <button
        type="button"
        onClick={handleDismiss}
        aria-label="このカードを閉じる"
        className="absolute right-3 top-3 rounded-full p-1 text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] hover:bg-surface-3 focus:outline-none focus:ring-2 focus:ring-primary-400"
      >
        <XMarkIcon className="h-4 w-4" aria-hidden="true" />
      </button>

      <div className="flex items-center gap-2">
        <SparklesIcon className="h-6 w-6 text-primary-300" aria-hidden="true" />
        <h2
          id="first-time-welcome-title"
          className="text-lg font-bold text-[var(--color-text-primary)] sm:text-xl"
        >
          {title}
        </h2>
      </div>
      <p className="mt-2 text-sm text-[var(--color-text-secondary)] leading-relaxed">
        AI と会話しながら、新卒エンジニア向けのビジネスコミュニケーションを練習できます。
        まずは以下のステップで使ってみましょう。
      </p>

      <ol className="mt-4 space-y-3">
        {steps.map((step, index) => (
          <li key={step.title} className="flex items-start gap-3">
            <span
              className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary-500/20 text-xs font-semibold text-primary-300"
              aria-hidden="true"
            >
              {index + 1}
            </span>
            <div className="min-w-0">
              <p className="text-sm font-semibold text-[var(--color-text-primary)]">
                {step.title}
              </p>
              <p className="text-xs text-[var(--color-text-muted)] leading-relaxed">
                {step.description}
              </p>
            </div>
          </li>
        ))}
      </ol>

      {onPrimaryAction && (
        <button
          type="button"
          onClick={onPrimaryAction}
          className="mt-5 inline-flex items-center justify-center rounded-lg bg-primary-500 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-600 active:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:ring-offset-2 focus:ring-offset-[var(--color-surface-1)]"
        >
          {primaryActionLabel}
        </button>
      )}
    </section>
  );
}
