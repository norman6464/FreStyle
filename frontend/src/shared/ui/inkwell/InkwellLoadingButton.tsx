import { ReactNode } from 'react';
import InkwellButton, { type InkwellButtonProps } from './InkwellButton';
import { useAsyncAction, type AsyncStatus } from './useAsyncAction';

export interface InkwellLoadingButtonProps
  extends Omit<InkwellButtonProps, 'onClick' | 'children'> {
  children: ReactNode;
  /** 押下で走らせる非同期処理。解決で成功、reject で失敗表示になる。 */
  onAction: () => Promise<void>;
  /** loading 中に読み上げる文言。 */
  loadingLabel?: string;
  /** 成功時に読み上げる文言。 */
  successLabel?: string;
  /** 失敗時に読み上げる文言。 */
  errorLabel?: string;
}

/**
 * 押下でラベルがスピナー→（成功でチェック / 失敗で×）に切り替わる送信ボタン。
 * 状態機械で二重送信を防ぎ、幅は固定してレイアウトが揺れない。動きは prefers-reduced-motion を尊重する。
 */
export default function InkwellLoadingButton({
  children,
  onAction,
  loadingLabel = '送信中',
  successLabel = '完了しました',
  errorLabel = '失敗しました',
  className = '',
  ...buttonProps
}: InkwellLoadingButtonProps) {
  const { status, run } = useAsyncAction(onAction);
  const busy = status !== 'idle';

  return (
    <InkwellButton
      {...buttonProps}
      // native disabled は使わずフォーカスを保持（状態変化を読み上げ可能に）。二重送信は状態機械で防ぐ。
      aria-disabled={busy || undefined}
      onClick={(e) => {
        if (busy) {
          e.preventDefault();
          return;
        }
        void run();
      }}
      className={`relative min-w-[8rem] ${className}`}
    >
      {/* ラベル層: loading 以降は透明化して幅だけ維持 */}
      <span className={status === 'idle' ? '' : 'invisible'}>{children}</span>

      {busy && (
        <span className="absolute inset-0 grid place-items-center">
          {status === 'loading' && (
            <span
              aria-hidden="true"
              className="h-5 w-5 rounded-full border-2 border-current border-t-transparent opacity-90 animate-spin motion-reduce:animate-none"
            />
          )}
          {status === 'success' && <CheckDraw />}
          {status === 'error' && <CrossMark />}
        </span>
      )}

      {/* 状態は見た目アイコンではなくテキストで支援技術へ伝える */}
      <span className="sr-only" role="status">
        {statusText(status, { loadingLabel, successLabel, errorLabel })}
      </span>
    </InkwellButton>
  );
}

function statusText(
  status: AsyncStatus,
  labels: { loadingLabel: string; successLabel: string; errorLabel: string },
): string {
  if (status === 'loading') return labels.loadingLabel;
  if (status === 'success') return labels.successLabel;
  if (status === 'error') return labels.errorLabel;
  return '';
}

/** 成功チェック。線を描く（reduced-motion では即表示）。 */
function CheckDraw() {
  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5" aria-hidden="true">
      <path
        d="M4 12.5l5 5 11-11"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        pathLength={1}
        className="[stroke-dasharray:1] [stroke-dashoffset:1] animate-inkwell-draw motion-reduce:animate-none motion-reduce:[stroke-dashoffset:0]"
      />
    </svg>
  );
}

/** 失敗マーク。 */
function CrossMark() {
  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5" aria-hidden="true">
      <path
        d="M6 6l12 12M18 6L6 18"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.5"
        strokeLinecap="round"
      />
    </svg>
  );
}
