import { ReactNode, useState } from 'react';
import { LightBulbIcon, XMarkIcon } from '@heroicons/react/24/outline';

type GuidedHintProps = {
  /** ヒントのタイトル */
  title: string;
  /** ヒントの本文 */
  children: ReactNode;
  /** 閉じた状態を localStorage に保存するキー（指定すると永続化） */
  storageKey?: string;
  /** 表示トーン */
  tone?: 'info' | 'success' | 'warning';
  /** 閉じるボタンを表示するか */
  dismissible?: boolean;
  /** 閉じたときに呼ばれるコールバック */
  onDismiss?: () => void;
};

const TONE_CLASSES: Record<NonNullable<GuidedHintProps['tone']>, string> = {
  info: 'border-l-primary-400 bg-primary-500/10',
  success: 'border-l-emerald-400 bg-emerald-500/10',
  warning: 'border-l-amber-400 bg-amber-500/10',
};

const ICON_TONE_CLASSES: Record<NonNullable<GuidedHintProps['tone']>, string> = {
  info: 'text-primary-300',
  success: 'text-emerald-300',
  warning: 'text-amber-300',
};

/**
 * 初心者向けのガイドヒント。画面上部や機能説明の近くに表示し、
 * 一度閉じたら localStorage に記録して再表示しない。
 */
export default function GuidedHint({
  title,
  children,
  storageKey,
  tone = 'info',
  dismissible = true,
  onDismiss,
}: GuidedHintProps) {
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
        // localStorage が使えない環境では単にスキップ
      }
    }
    onDismiss?.();
  };

  return (
    <aside
      role="note"
      className={`flex items-start gap-3 rounded-lg border border-surface-3 border-l-4 ${TONE_CLASSES[tone]} p-4`}
    >
      <LightBulbIcon
        className={`h-5 w-5 shrink-0 ${ICON_TONE_CLASSES[tone]}`}
        aria-hidden="true"
      />
      <div className="flex-1 min-w-0">
        <p className="mb-1 text-sm font-semibold text-[var(--color-text-primary)]">{title}</p>
        <div className="text-sm text-[var(--color-text-secondary)] leading-relaxed">{children}</div>
      </div>
      {dismissible && (
        <button
          type="button"
          onClick={handleDismiss}
          aria-label="ヒントを閉じる"
          className="shrink-0 rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] hover:bg-surface-3 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-400"
        >
          <XMarkIcon className="h-4 w-4" aria-hidden="true" />
        </button>
      )}
    </aside>
  );
}
