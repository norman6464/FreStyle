import { Suspense, useEffect, useState } from 'react';
import { CheckCircleIcon, ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/24/outline';
import { ExerciseHeader } from '@/entities/exercise';
import { monacoLanguageOf } from '@/entities/exercise';
import type { MasterExercise, ExerciseSubmitResult } from '@/entities/exercise';
import MarkdownView from '@/shared/ui/MarkdownView';
import LanguageBadge from '@/shared/ui/LanguageBadge';
import { lazyWithReload } from '@/shared/lib/lazyWithReload';

const CodeEditor = lazyWithReload(() => import('@/shared/ui/CodeEditor'), 'CodeEditor');

/** 入力のたびに iframe 全体が再パースされるのを避けるための、プレビュー反映の遅延 (ms)。 */
const PREVIEW_DEBOUNCE_MS = 300;

/** value の変化を delayMs 落ち着いてから反映する（タイプ中の連続反映を抑える）。 */
function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(timer);
  }, [value, delayMs]);
  return debounced;
}

interface Props {
  exercise: MasterExercise;
  code: string;
  onCodeChange: (code: string) => void;
  submitting: boolean;
  submitResult: ExerciseSubmitResult | null;
  submitError: string | null;
  /** 今回の提出 or 過去履歴に正解があるか（親ページで算出）。 */
  solved: boolean;
  onSubmit: () => void;
  onReset: () => void;
}

/**
 * PreviewExerciseView — `mode='preview'` (HTML/CSS ライブプレビュー演習) の描画。
 *
 * サンドボックス実行や stdout 比較をせず、エディタの内容を iframe で即時描画し、
 * 見本（expectedOutput の模範 HTML）と見比べながら作る。採点は学習者の自己申告
 * （「できた！」ボタン → 既存の提出 API。backend は isCorrect=true / results=[] を返す）。
 *
 * セキュリティ: iframe は `sandbox=""`（スクリプト実行・同一オリジンとも不許可）で描画する。
 * 学習者の入力 HTML がアプリの origin で動くことは無い。この属性を緩めないこと。
 */
export default function PreviewExerciseView({
  exercise: ex,
  code,
  onCodeChange,
  submitting,
  submitResult,
  submitError,
  solved,
  onSubmit,
  onReset,
}: Props) {
  const [showHint, setShowHint] = useState(false);
  // タイプごとの iframe 再パースを避けるため、プレビューは 300ms デバウンスで反映する。
  const previewHtml = useDebouncedValue(code, PREVIEW_DEBOUNCE_MS);

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-4xl mx-auto space-y-6">
      <ExerciseHeader exercise={ex} submitResult={submitResult} />

      {/* 問題カード */}
      <section className="rounded-lg border border-surface-3 bg-surface-1 p-5 space-y-5">
        <div className="space-y-2">
          <h2 className="text-base font-semibold text-[var(--color-text-primary)] flex items-center gap-2">
            <span aria-hidden>📃</span>
            下記の見本と同じ見た目になるようにコードを書いてみよう！
          </h2>
          <div className="prose prose-sm max-w-none text-sm text-[var(--color-text-primary)] leading-relaxed">
            <MarkdownView content={ex.description} />
          </div>
        </div>

        <div className="text-xs text-taupe-400">
          ▼ 下記解答欄にコードを記入すると、プレビューへ即時に反映されます
        </div>

        {ex.hintText && (
          <div>
            <button
              onClick={() => setShowHint((v) => !v)}
              aria-expanded={showHint}
              aria-controls="exercise-hint-panel"
              className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] flex items-center gap-1"
            >
              {showHint ? <ChevronUpIcon className="w-3.5 h-3.5" /> : <ChevronDownIcon className="w-3.5 h-3.5" />}
              ヒントを{showHint ? '隠す' : '見る'}
            </button>
            {showHint && (
              <div
                id="exercise-hint-panel"
                className="mt-2 p-3 bg-blue-500/10 border border-blue-500/20 rounded-md text-xs text-[var(--color-text-primary)] whitespace-pre-wrap"
              >
                {ex.hintText}
              </div>
            )}
          </div>
        )}
      </section>

      {/* エディタ */}
      <section className="rounded-lg border border-surface-3 bg-surface-1 overflow-hidden">
        <div className="flex items-center justify-between px-4 py-2 bg-surface-2 border-b border-surface-3">
          <span className="text-sm font-semibold text-[var(--color-text-primary)]">解答コード入力欄</span>
          <div className="flex items-center gap-2">
            <button
              onClick={onReset}
              className="text-xs px-3 py-1 rounded text-[var(--color-text-muted)] hover:bg-surface-3 hover:text-[var(--color-text-primary)] transition-colors"
            >
              リセット
            </button>
            <LanguageBadge language={ex.language} mono />
          </div>
        </div>
        <div className="bg-[#1e1e1e]">
          <Suspense fallback={<div style={{ height: 360 }} className="bg-[#1e1e1e]" />}>
            <CodeEditor
              value={code}
              onChange={onCodeChange}
              language={monacoLanguageOf(ex.language)}
              minHeight={260}
            />
          </Suspense>
        </div>
      </section>

      {/* プレビューと見本: 見比べられるよう横並び（狭い画面では縦積み） */}
      <section className="grid gap-4 sm:grid-cols-2">
        <div className="rounded-lg border border-surface-3 bg-surface-1 overflow-hidden">
          <div className="px-4 py-2 bg-surface-2 border-b border-surface-3">
            <span className="text-sm font-semibold text-[var(--color-text-primary)]">プレビュー</span>
            <span className="ml-2 text-xs text-[var(--color-text-muted)]">あなたのコードの表示結果</span>
          </div>
          {/* sandbox="" はセキュリティ要件（スクリプト実行・同一オリジンとも不許可）。緩めない。 */}
          <iframe
            srcDoc={previewHtml}
            sandbox=""
            title="プレビュー"
            className="w-full h-72 bg-white"
          />
        </div>
        <div className="rounded-lg border border-surface-3 bg-surface-1 overflow-hidden">
          <div className="px-4 py-2 bg-surface-2 border-b border-surface-3">
            <span className="text-sm font-semibold text-[var(--color-text-primary)]">見本</span>
            <span className="ml-2 text-xs text-[var(--color-text-muted)]">この見た目を目指そう</span>
          </div>
          <iframe
            srcDoc={ex.expectedOutput}
            sandbox=""
            title="見本"
            className="w-full h-72 bg-white"
          />
        </div>
      </section>

      {/* できた！ボタン: 見本と同じ見た目にできたら自己申告で提出する */}
      {solved ? (
        <div className="flex justify-center">
          <p
            role="status"
            className="w-full max-w-sm flex items-center justify-center gap-2 px-6 py-2.5 rounded-md bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-sm font-semibold"
          >
            <CheckCircleIcon className="w-4 h-4" />
            この演習はクリア済みです
          </p>
        </div>
      ) : (
        <div className="flex justify-center">
          <button
            onClick={onSubmit}
            disabled={submitting || !code.trim()}
            className="w-full max-w-sm flex items-center justify-center gap-2 px-6 py-2.5 rounded-md bg-amber-700/70 hover:bg-amber-700/90 disabled:opacity-50 text-white text-sm font-semibold transition-colors"
          >
            <CheckCircleIcon className="w-4 h-4" />
            {submitting ? '記録中...' : 'できた！'}
          </button>
        </div>
      )}
      {submitError && (
        <p role="alert" className="text-center text-xs text-red-400">{submitError}</p>
      )}
    </div>
  );
}
