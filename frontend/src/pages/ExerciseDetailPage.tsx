import { Suspense, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  XCircleIcon,
  ChevronDownIcon,
  ChevronUpIcon,
} from '@heroicons/react/24/outline';
import Loading from '../components/Loading';
import { useExerciseDetail } from '../hooks/useExerciseDetail';
import { lazyWithReload } from '../utils/lazyWithReload';
import { ExerciseSubmission, ExerciseTestCaseResult } from '../types';

const CodeEditor = lazyWithReload(() => import('../components/CodeEditor'), 'CodeEditor');

/**
 * ExerciseDetailPage — `/code-editor/:slug` の詳細画面。
 *
 * 上段: 問題タイトル + 説明 + 入出力例 (テストケース表示)
 * 下段: コードエディタ + 実行 / 提出ボタン + 結果パネル + 履歴
 */
export default function ExerciseDetailPage() {
  const { slug } = useParams<{ slug: string }>();
  const {
    detail,
    code,
    setCode,
    loading,
    error,
    running,
    executionResult,
    submitting,
    submitResult,
    submitError,
    submissions,
    runCode,
    submitCode,
    resetCode,
  } = useExerciseDetail(slug);

  const [showHint, setShowHint] = useState(false);

  if (loading) {
    return <Loading message="読み込み中..." className="min-h-[50vh]" />;
  }
  if (error || !detail) {
    return (
      <div className="px-6 pt-6 max-w-3xl mx-auto">
        <BackLink />
        <p className="mt-6 text-red-500">{error ?? '演習問題が見つかりません'}</p>
      </div>
    );
  }

  const ex = detail.exercise;

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* ヘッダー */}
      <header className="flex-shrink-0 px-6 py-3 border-b border-surface-3 bg-surface-1 space-y-2">
        <BackLink />
        <div className="flex items-center gap-3 flex-wrap">
          <h1 className="text-lg font-bold text-[var(--color-text-primary)]">{ex.title}</h1>
          <span className="text-xs px-1.5 py-0.5 rounded bg-surface-3 text-[var(--color-text-muted)] uppercase">
            {ex.language}
          </span>
          {submitResult && (
            submitResult.isCorrect
              ? <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-green-500/15 text-green-400 border border-green-500/30">
                  <CheckCircleIcon className="w-4 h-4" /> 全テストケース合格
                </span>
              : <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-red-500/15 text-red-400 border border-red-500/30">
                  <XCircleIcon className="w-4 h-4" /> 不合格
                </span>
          )}
        </div>
      </header>

      <div className="flex-1 flex flex-col lg:flex-row min-h-0 overflow-hidden">
        {/* 左: 問題説明 + テストケース + 履歴 */}
        <aside className="lg:w-2/5 lg:flex-shrink-0 overflow-y-auto p-6 space-y-5 border-b lg:border-b-0 lg:border-r border-surface-3 bg-surface-1">
          <section className="space-y-2">
            <h2 className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wider">
              問題
            </h2>
            <p className="text-sm text-[var(--color-text-primary)] whitespace-pre-wrap leading-relaxed">
              {ex.description}
            </p>
          </section>

          {ex.hintText && (
            <section className="space-y-2">
              <button
                onClick={() => setShowHint((v) => !v)}
                className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] flex items-center gap-1"
              >
                {showHint ? <ChevronUpIcon className="w-3.5 h-3.5" /> : <ChevronDownIcon className="w-3.5 h-3.5" />}
                ヒントを{showHint ? '隠す' : '見る'}
              </button>
              {showHint && (
                <div className="p-3 bg-blue-500/10 border border-blue-500/20 rounded-md text-xs text-[var(--color-text-primary)] whitespace-pre-wrap">
                  {ex.hintText}
                </div>
              )}
            </section>
          )}

          <section className="space-y-2">
            <h2 className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wider">
              入出力例 ({detail.examples.length})
            </h2>
            <div className="space-y-3">
              {detail.examples.map((ex, idx) => (
                <ExampleBlock
                  key={ex.id}
                  index={idx + 1}
                  input={ex.inputText}
                  expected={ex.expectedOutput}
                />
              ))}
            </div>
          </section>

          {submissions.length > 0 && (
            <section className="space-y-2">
              <h2 className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wider">
                提出履歴 ({submissions.length})
              </h2>
              <ul className="space-y-1">
                {submissions.slice(0, 10).map((s) => (
                  <SubmissionRow key={s.id} submission={s} />
                ))}
              </ul>
            </section>
          )}
        </aside>

        {/* 右: エディタ + 結果 */}
        <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
          <div className="flex items-center justify-between px-4 py-2 bg-surface-2 border-b border-surface-3 flex-shrink-0">
            <span className="text-xs text-[var(--color-text-muted)] font-mono uppercase">{ex.language}</span>
            <div className="flex gap-2">
              <button
                onClick={resetCode}
                className="text-xs px-3 py-1 rounded text-[var(--color-text-muted)] hover:bg-surface-3 hover:text-[var(--color-text-primary)] transition-colors"
              >
                リセット
              </button>
              <button
                onClick={runCode}
                disabled={running || submitting}
                className="text-xs px-3 py-1 rounded border border-surface-3 text-[var(--color-text-secondary)] hover:bg-surface-3 hover:text-[var(--color-text-primary)] disabled:opacity-50 transition-colors"
              >
                {running ? '実行中...' : '▶ 実行'}
              </button>
              <button
                onClick={submitCode}
                disabled={running || submitting}
                className="text-xs px-3 py-1 rounded bg-primary-500 hover:bg-primary-600 disabled:opacity-50 text-white font-medium transition-colors"
              >
                {submitting ? '採点中...' : '提出'}
              </button>
            </div>
          </div>

          <div className="flex-1 overflow-hidden bg-[#1e1e1e]">
            <Suspense fallback={<div className="h-full bg-[#1e1e1e]" />}>
              <CodeEditor
                value={code}
                onChange={setCode}
                language={ex.language === 'php' ? 'php' : 'plaintext'}
                height="100%"
              />
            </Suspense>
          </div>

          <div className="border-t border-surface-3 bg-surface-1 max-h-[40%] min-h-[140px] overflow-y-auto p-4 text-xs space-y-3">
            {!executionResult && !submitResult && !submitError && !running && !submitting && (
              <p className="text-[var(--color-text-muted)]">
                「▶ 実行」で出力プレビュー、「提出」で全テストケース採点。
              </p>
            )}

            {executionResult && !submitResult && (
              <ExecutionPreview result={executionResult} />
            )}

            {submitError && (
              <p className="text-red-400">{submitError}</p>
            )}

            {submitResult && (
              <SubmitResultPanel results={submitResult.results} />
            )}
          </div>
        </main>
      </div>
    </div>
  );
}

function BackLink() {
  return (
    <Link
      to="/code-editor"
      className="inline-flex items-center gap-1 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
    >
      <ArrowLeftIcon className="w-3.5 h-3.5" />
      問題一覧に戻る
    </Link>
  );
}

function ExampleBlock({ index, input, expected }: { index: number; input: string; expected: string }) {
  return (
    <div className="rounded-md border border-surface-3 bg-surface-2 p-2 space-y-2 text-xs font-mono">
      <p className="text-[var(--color-text-muted)] not-italic font-sans text-[10px] uppercase tracking-wider">
        例 {index}
      </p>
      {input && (
        <div>
          <p className="text-[10px] text-[var(--color-text-muted)] font-sans mb-0.5">入力</p>
          <pre className="whitespace-pre-wrap break-words bg-[var(--color-surface)] p-1.5 rounded text-[var(--color-text-primary)]">{input}</pre>
        </div>
      )}
      <div>
        <p className="text-[10px] text-[var(--color-text-muted)] font-sans mb-0.5">期待出力</p>
        <pre className="whitespace-pre-wrap break-words bg-[var(--color-surface)] p-1.5 rounded text-[var(--color-text-primary)]">{expected}</pre>
      </div>
    </div>
  );
}

function ExecutionPreview({ result }: { result: { stdout: string; stderr: string; exitCode: number } }) {
  return (
    <div className="space-y-2">
      <p className="text-[var(--color-text-muted)]">
        実行結果（exit code: {result.exitCode}）
      </p>
      {result.stdout && (
        <pre className="whitespace-pre-wrap break-words font-mono text-[var(--color-text-primary)]">{result.stdout}</pre>
      )}
      {result.stderr && (
        <pre className="whitespace-pre-wrap break-words font-mono text-red-400">{result.stderr}</pre>
      )}
    </div>
  );
}

function SubmitResultPanel({ results }: { results: ExerciseTestCaseResult[] }) {
  return (
    <div className="space-y-2">
      {results.map((r) => (
        <TestCaseResultRow key={r.orderIndex} r={r} />
      ))}
    </div>
  );
}

function TestCaseResultRow({ r }: { r: ExerciseTestCaseResult }) {
  return (
    <details
      className={`rounded border ${r.passed ? 'border-green-500/30 bg-green-500/5' : 'border-red-500/30 bg-red-500/5'} p-2`}
    >
      <summary className="cursor-pointer flex items-center gap-2 text-xs">
        {r.passed
          ? <CheckCircleIcon className="w-4 h-4 text-green-400" />
          : <XCircleIcon className="w-4 h-4 text-red-400" />}
        <span className="font-mono text-[var(--color-text-primary)]">テストケース {r.orderIndex}</span>
        <span className={r.passed ? 'text-green-400' : 'text-red-400'}>
          {r.passed ? '合格' : '不合格'}
        </span>
      </summary>
      <div className="mt-2 grid grid-cols-1 md:grid-cols-2 gap-2 text-[11px] font-mono">
        {r.input && (
          <div>
            <p className="text-[var(--color-text-muted)] mb-0.5">入力</p>
            <pre className="whitespace-pre-wrap break-words bg-[var(--color-surface)] p-1.5 rounded">{r.input}</pre>
          </div>
        )}
        <div>
          <p className="text-[var(--color-text-muted)] mb-0.5">期待出力</p>
          <pre className="whitespace-pre-wrap break-words bg-[var(--color-surface)] p-1.5 rounded">{r.expectedOutput}</pre>
        </div>
        <div className="md:col-span-2">
          <p className="text-[var(--color-text-muted)] mb-0.5">実際の出力</p>
          <pre className="whitespace-pre-wrap break-words bg-[var(--color-surface)] p-1.5 rounded">{r.actualOutput || '(なし)'}</pre>
        </div>
        {r.stderr && (
          <div className="md:col-span-2">
            <p className="text-[var(--color-text-muted)] mb-0.5">stderr</p>
            <pre className="whitespace-pre-wrap break-words bg-[var(--color-surface)] p-1.5 rounded text-red-400">{r.stderr}</pre>
          </div>
        )}
      </div>
    </details>
  );
}

function SubmissionRow({ submission }: { submission: ExerciseSubmission }) {
  const date = new Date(submission.submittedAt);
  const stamp = `${date.getFullYear()}/${String(date.getMonth() + 1).padStart(2, '0')}/${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
  return (
    <li className="flex items-center gap-2 text-xs px-2 py-1 rounded border border-surface-3 bg-surface-2">
      {submission.isCorrect
        ? <CheckCircleIcon className="w-4 h-4 text-green-400 flex-shrink-0" />
        : <XCircleIcon className="w-4 h-4 text-red-400 flex-shrink-0" />}
      <span className="font-mono text-[var(--color-text-primary)]">{stamp}</span>
      <span className={submission.isCorrect ? 'text-green-400' : 'text-red-400'}>
        {submission.isCorrect ? '合格' : '不合格'}
      </span>
    </li>
  );
}
