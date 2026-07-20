import { CheckCircleIcon, XCircleIcon, ClipboardDocumentCheckIcon } from '@heroicons/react/24/outline';
import ExerciseHeader from './ExerciseHeader';
import MarkdownView from '@/components/message/MarkdownView';
import { MasterExercise, ExerciseSubmitResult } from '../model/types';

interface Props {
  exercise: MasterExercise;
  starterCode: string;
  onCodeChange: (code: string) => void;
  submitting: boolean;
  submitResult: ExerciseSubmitResult | null;
  submitError: string | null;
  onSubmit: () => void;
  onReset: () => void;
}

/**
 * QaExerciseView — `mode='qa'` の演習向け、 単行 input + 提出 + 解説 表示。
 *
 * docker / kubernetes / 多くのネットワーク系コマンドのように、 サンドボックス
 * 実行が困難な題材を 「コマンドを書き取る」 形式で扱う。
 *
 * UX:
 *   1. 質問文 (description) を表示
 *   2. `$` プロンプト付きの 単行 input
 *   3. 提出ボタン (Enter でも提出可)
 *   4. 採点後:
 *      - 正解: 緑バナー + 期待コマンドの再表示 + explanation (markdown)
 *      - 不正解: 赤バナー + 「もう一度入力してください」 (input は維持)
 */
export default function QaExerciseView({
  exercise: ex,
  starterCode,
  onCodeChange,
  submitting,
  submitResult,
  submitError,
  onSubmit,
  onReset,
}: Props) {
  const isCorrect = submitResult?.isCorrect === true;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (submitting || !starterCode.trim()) return;
    onSubmit();
  };

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-3xl mx-auto space-y-6">
      <ExerciseHeader exercise={ex} submitResult={submitResult} />

      {isCorrect && (
        <div
          role="status"
          aria-live="polite"
          className="rounded-lg border border-green-500/30 bg-green-500/10 px-4 py-3 flex items-center gap-2 text-sm text-green-400"
        >
          <CheckCircleIcon className="w-5 h-5 flex-shrink-0" />
          正解です。 詳細については下の解説を確認してください。
        </div>
      )}
      {submitResult && !isCorrect && (
        <div
          role="status"
          aria-live="polite"
          className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 flex items-center gap-2 text-sm text-red-400"
        >
          <XCircleIcon className="w-5 h-5 flex-shrink-0" />
          不正解です。 もう一度入力してください。
        </div>
      )}

      <section className="rounded-lg border border-surface-3 bg-surface-1 p-5 space-y-4">
        <p className="text-xs text-[var(--color-text-muted)]">
          以下の問題に対応するコマンドを入力してください。
        </p>
        <div className="prose prose-sm max-w-none text-sm text-[var(--color-text-primary)] leading-relaxed">
          <MarkdownView content={ex.description} />
        </div>

        <form onSubmit={handleSubmit} className="space-y-3">
          <div className="flex items-center gap-2 px-3 py-2 rounded-md bg-[#1e1e1e] border border-surface-3 font-mono text-sm">
            <span className="text-emerald-400 select-none">$</span>
            <input
              type="text"
              value={starterCode}
              onChange={(e) => onCodeChange(e.target.value)}
              disabled={submitting}
              autoFocus
              spellCheck={false}
              autoComplete="off"
              autoCapitalize="off"
              className="flex-1 bg-transparent outline-none text-white placeholder:text-[var(--color-text-muted)]"
              placeholder="ここにコマンドを入力..."
              aria-label="コマンドを入力"
            />
          </div>
          <div className="flex gap-2">
            <button
              type="submit"
              disabled={submitting || !starterCode.trim()}
              className="flex-1 flex items-center justify-center gap-2 px-4 py-2 rounded-md bg-amber-700/70 hover:bg-amber-700/90 disabled:opacity-50 text-white text-sm font-semibold transition-colors"
            >
              <ClipboardDocumentCheckIcon className="w-4 h-4" />
              {submitting ? '採点中...' : '解答する'}
            </button>
            <button
              type="button"
              onClick={onReset}
              disabled={submitting}
              className="px-4 py-2 rounded-md text-sm text-[var(--color-text-muted)] hover:bg-surface-2 hover:text-[var(--color-text-primary)] disabled:opacity-50 transition-colors"
            >
              リセット
            </button>
          </div>
          {submitError && (
            <p role="alert" className="text-xs text-red-400">{submitError}</p>
          )}
        </form>
      </section>

      {isCorrect && (
        <section className="rounded-lg border border-surface-3 bg-surface-1 overflow-hidden">
          <div className="px-4 py-3 bg-green-500/10 border-b border-green-500/20">
            <p className="text-xs font-semibold text-green-400 uppercase tracking-wider mb-1">
              回答は正解です
            </p>
            <code className="text-sm text-[var(--color-text-primary)] font-mono">
              {ex.expectedOutput}
            </code>
          </div>
          {ex.explanation && (
            <div className="px-4 py-3">
              <p className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wider mb-2">
                説明
              </p>
              <div className="prose prose-sm max-w-none text-sm text-[var(--color-text-primary)]">
                <MarkdownView content={ex.explanation} />
              </div>
            </div>
          )}
        </section>
      )}
    </div>
  );
}
