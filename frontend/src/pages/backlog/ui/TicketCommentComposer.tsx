import { useState } from 'react';
import type { TicketCommentSegment } from '@/entities/ticket';

export interface TicketCommentComposerProps {
  /** 失敗は投げてくる前提（投げられたら入力を保つ）。 */
  onSubmit: (body: TicketCommentSegment[]) => Promise<void>;
  placeholder?: string;
  submitLabel?: string;
  autoFocus?: boolean;
}

/**
 * 発言の入力欄。当面はふつうの複数行入力（@ の候補選択による名指しは、backend の
 * 人の一覧 API が入ってから別の入力欄に差し替える。それまでは打っても名指しにならない
 * ただの文字として送られる — 通知は飛ばないが、送信自体は妨げない）。
 *
 * 空判定は「trim 後が空」。backend は「配列が空」または「text ノードだけで trim 後が
 * 全部空」を本文全体ごと 400 で拒む境界を持つので、ここで先に止める
 * （サーバー応答は invalid_request に潰れて理由が分からないため）。
 */
export default function TicketCommentComposer({
  onSubmit,
  placeholder = 'コメントを書く',
  submitLabel = '送信',
  autoFocus = false,
}: TicketCommentComposerProps) {
  const [value, setValue] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const trimmed = value.trim();
  const canSubmit = trimmed !== '' && !submitting;

  const handleSubmit = async () => {
    if (!canSubmit) return;
    setSubmitting(true);
    setError(null);
    try {
      await onSubmit([{ kind: 'text', text: value }]);
      setValue('');
    } catch {
      setError('送信できませんでした。もう一度お試しください。');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="rounded-xl border border-surface-3 p-2">
      <textarea
        value={value}
        onChange={(e) => setValue(e.target.value)}
        placeholder={placeholder}
        aria-label={placeholder}
        rows={2}
        disabled={submitting}
        autoFocus={autoFocus}
        className="w-full resize-none bg-transparent text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none disabled:opacity-60"
      />
      {error && (
        <p role="alert" className="mt-1 text-xs leading-relaxed text-red-700">
          {error}
        </p>
      )}
      <div className="mt-1.5 flex items-center">
        <span className="text-[11px] text-[var(--color-text-muted)]">@ で名前を挙げると通知が届きます</span>
        <button
          type="button"
          onClick={() => void handleSubmit()}
          disabled={!canSubmit}
          className="ml-auto rounded border-0 bg-brand-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-brand-700 disabled:opacity-50"
        >
          {submitting ? '送信中…' : submitLabel}
        </button>
      </div>
    </div>
  );
}
