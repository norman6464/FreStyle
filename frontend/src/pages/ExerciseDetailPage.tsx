import { Suspense, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  XCircleIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  DocumentTextIcon,
  InboxIcon,
  ClipboardDocumentCheckIcon,
} from '@heroicons/react/24/outline';
import Loading from '../components/Loading';
import { useExerciseDetail } from '../hooks/useExerciseDetail';
import { lazyWithReload } from '../utils/lazyWithReload';
import {
  CodeExecutionResult,
  ExerciseSubmission,
  ExerciseTestCaseResult,
  MasterExerciseExample,
} from '../types';

const CodeEditor = lazyWithReload(() => import('../components/CodeEditor'), 'CodeEditor');

// master_exercises.language → Monaco のシンタックスハイライト言語ID へのマッピング。
// Monaco は `php` / `go` / `shell` を組み込みでサポート。 該当しない場合は plaintext に fallback する。
function monacoLanguageOf(lang: string): string {
  switch (lang) {
    case 'php':
      return 'php';
    case 'go':
      return 'go';
    case 'bash':
      return 'shell';
    default:
      return 'plaintext';
  }
}

/**
 * ExerciseDetailPage — `/code-editor/:slug` の詳細画面。
 *
 * 縦 1 カラムのスクロールレイアウト:
 *   1. ヘッダー (戻るリンク + タイトル + 採点結果バッジ)
 *   2. 問題カード（説明 + 入出力例ブロック）
 *   3. ヒント開閉
 *   4. コードエディタ + 「提出前動作確認」(単発実行)
 *   5. 実行結果テーブル（status / 出力 / 期待値）
 *   6. 「コードを提出する」フルワイドボタン
 *   7. 提出履歴
 *
 * 単一テストケースしか表示しない既存 LP よりも、 全テストケースを縦方向に並べる方が
 * 視認性が高い。 採点後はテーブルで status / 期待出力 / 実出力が並び、 不一致の
 * 特定が容易になるよう設計。
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
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-4xl mx-auto space-y-6">
      {/* ヘッダー */}
      <header className="space-y-3">
        <BackLink />
        <div className="flex items-start gap-3">
          <div className="flex-shrink-0 w-10 h-10 rounded-md bg-primary-500/10 border border-primary-500/30 flex items-center justify-center">
            <DocumentTextIcon className="w-5 h-5 text-primary-400" />
          </div>
          <div className="min-w-0 flex-1">
            <h1 className="text-xl font-bold text-[var(--color-text-primary)] leading-tight">
              {ex.title}
            </h1>
            <div className="mt-1 flex items-center gap-2 text-xs text-[var(--color-text-muted)]">
              <span className="px-1.5 py-0.5 rounded bg-surface-3 uppercase">{ex.language}</span>
              <span>難易度 {'★'.repeat(Math.max(1, Math.min(5, ex.difficulty)))}</span>
              <span>#{ex.orderIndex}</span>
            </div>
          </div>
          {submitResult && <ResultBadge isCorrect={submitResult.isCorrect} />}
        </div>
      </header>

      {/* 問題カード */}
      <section className="rounded-lg border border-surface-3 bg-surface-1 p-5 space-y-5">
        <div className="space-y-2">
          <h2 className="text-base font-semibold text-[var(--color-text-primary)] flex items-center gap-2">
            <span aria-hidden>📃</span>
            下記の問題をプログラミングしてみよう！
          </h2>
          <p className="text-sm text-[var(--color-text-primary)] whitespace-pre-wrap leading-relaxed">
            {ex.description}
          </p>
        </div>

        <div className="text-xs text-primary-400">
          ▼ 下記解答欄にコードを記入してみよう
        </div>

        <div className="space-y-3">
          {detail.examples.map((example, idx) => (
            <ExampleBlock
              key={example.id}
              index={idx + 1}
              total={detail.examples.length}
              example={example}
            />
          ))}
        </div>

        {ex.hintText && (
          <div>
            <button
              onClick={() => setShowHint((v) => !v)}
              className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] flex items-center gap-1"
            >
              {showHint ? <ChevronUpIcon className="w-3.5 h-3.5" /> : <ChevronDownIcon className="w-3.5 h-3.5" />}
              ヒントを{showHint ? '隠す' : '見る'}
            </button>
            {showHint && (
              <div className="mt-2 p-3 bg-blue-500/10 border border-blue-500/20 rounded-md text-xs text-[var(--color-text-primary)] whitespace-pre-wrap">
                {ex.hintText}
              </div>
            )}
          </div>
        )}
      </section>

      {/* テストケース注意書き */}
      <p className="text-xs text-orange-400 bg-orange-500/10 border border-orange-500/20 rounded-md px-3 py-2">
        ❓ 複数のテストケースで採点しますので、 動作確認用の入力例だけでなく入力値を変えてのデバッグもおすすめします。
      </p>

      {/* エディタ */}
      <section className="rounded-lg border border-surface-3 bg-surface-1 overflow-hidden">
        <div className="flex items-center justify-between px-4 py-2 bg-surface-2 border-b border-surface-3">
          <span className="text-sm font-semibold text-[var(--color-text-primary)]">解答コード入力欄</span>
          <div className="flex items-center gap-2">
            <button
              onClick={resetCode}
              className="text-xs px-3 py-1 rounded text-[var(--color-text-muted)] hover:bg-surface-3 hover:text-[var(--color-text-primary)] transition-colors"
            >
              リセット
            </button>
            <span className="text-xs px-2 py-0.5 rounded bg-surface-3 text-[var(--color-text-muted)] font-mono uppercase">
              {ex.language}
            </span>
          </div>
        </div>
        <div className="h-[360px] bg-[#1e1e1e]">
          <Suspense fallback={<div className="h-full bg-[#1e1e1e]" />}>
            <CodeEditor
              value={code}
              onChange={setCode}
              language={monacoLanguageOf(ex.language)}
              height="100%"
            />
          </Suspense>
        </div>
        <div className="px-4 py-3 bg-surface-2 border-t border-surface-3">
          <button
            onClick={runCode}
            disabled={running || submitting}
            className="w-full px-4 py-2 rounded-md bg-blue-500/80 hover:bg-blue-500 disabled:opacity-50 text-white text-sm font-medium transition-colors"
          >
            {running ? '実行中...' : 'コード実行'}
          </button>
        </div>
      </section>

      {/* 実行結果テーブル */}
      {(executionResult || submitError) && (
        <ExecutionResultTable
          result={executionResult}
          expected={detail.examples[0]?.expectedOutput ?? ''}
          submitError={submitError}
        />
      )}

      {/* 採点結果（提出後） */}
      {submitResult && <SubmitResultPanel results={submitResult.results} />}

      {/* 提出ボタン: 中央寄せ・控えめなトーン・横幅は max-w-sm */}
      <div className="flex justify-center">
        <button
          onClick={submitCode}
          disabled={running || submitting}
          className="w-full max-w-sm flex items-center justify-center gap-2 px-6 py-2.5 rounded-md bg-amber-700/70 hover:bg-amber-700/90 disabled:opacity-50 text-white text-sm font-semibold transition-colors"
        >
          <ClipboardDocumentCheckIcon className="w-4 h-4" />
          {submitting ? '採点中...' : 'コードを提出する'}
        </button>
      </div>

      {/* 提出履歴 */}
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

function ResultBadge({ isCorrect }: { isCorrect: boolean }) {
  if (isCorrect) {
    return (
      <span className="flex-shrink-0 flex items-center gap-1 text-xs px-2 py-1 rounded-full bg-green-500/15 text-green-400 border border-green-500/30">
        <CheckCircleIcon className="w-4 h-4" /> 全テストケース合格
      </span>
    );
  }
  return (
    <span className="flex-shrink-0 flex items-center gap-1 text-xs px-2 py-1 rounded-full bg-red-500/15 text-red-400 border border-red-500/30">
      <XCircleIcon className="w-4 h-4" /> 不合格
    </span>
  );
}

function ExampleBlock({
  index,
  total,
  example,
}: {
  index: number;
  total: number;
  example: MasterExerciseExample;
}) {
  // 例が複数ある場合は番号を表示し、 1 件しかないときはシンプルに「入力される値 / 期待する出力」だけにする。
  const suffix = total > 1 ? ` ${index}` : '';
  const inputDisplay = example.inputText.length > 0 ? example.inputText : 'ありません。';
  return (
    <div className="space-y-3">
      <div className="space-y-1.5 pb-4 border-b border-surface-3 last:border-b-0 last:pb-0">
        <div className="flex items-center gap-2 text-sm text-[var(--color-text-primary)] font-semibold">
          <InboxIcon className="w-4 h-4 text-[var(--color-text-muted)]" />
          入力される値{suffix}
        </div>
        <pre className="whitespace-pre-wrap break-words text-sm text-[var(--color-text-primary)]">
          {inputDisplay}
        </pre>
        {example.inputText.length === 0 && (
          <p className="text-xs text-[var(--color-text-muted)] leading-relaxed pt-1">
            入力値最終行の末尾に改行が 1 つ入ります。<br />
            文字列は標準入力から渡されます。
          </p>
        )}
      </div>

      <div className="space-y-1.5">
        <div className="flex items-center gap-2 text-sm text-[var(--color-text-primary)] font-semibold">
          <ClipboardDocumentCheckIcon className="w-4 h-4 text-[var(--color-text-muted)]" />
          期待する出力{suffix}
        </div>
        <pre className="whitespace-pre-wrap break-words text-sm text-[var(--color-text-primary)] bg-surface-2 border border-surface-3 rounded px-3 py-2 font-mono">
          {example.expectedOutput}
        </pre>
      </div>
    </div>
  );
}

function ExecutionResultTable({
  result,
  expected,
  submitError,
}: {
  result: CodeExecutionResult | null;
  expected: string;
  submitError: string | null;
}) {
  if (submitError) {
    return (
      <div className="rounded-md border border-red-500/30 bg-red-500/5 p-3 text-xs text-red-400">
        {submitError}
      </div>
    );
  }
  if (!result) return null;

  const status = result.exitCode === 0 ? 'Success' : `Failure (exit ${result.exitCode})`;
  const matched = result.exitCode === 0 && normalizeOutput(result.stdout) === normalizeOutput(expected);

  return (
    <div className="rounded-lg border border-surface-3 bg-surface-1 overflow-hidden">
      <table className="w-full text-xs">
        <tbody>
          <tr className="border-b border-surface-3">
            <th className="text-left px-4 py-2 font-medium text-[var(--color-text-muted)] bg-surface-2 w-1/3">
              実行結果ステータス
            </th>
            <td className="px-4 py-2 text-[var(--color-text-primary)]">{status}</td>
          </tr>
          <tr className="border-b border-surface-3">
            <th className="text-left px-4 py-2 font-medium text-[var(--color-text-muted)] bg-surface-2 align-top">
              提出コードのアウトプット
            </th>
            <td className="px-4 py-2">
              <pre className="whitespace-pre-wrap break-words font-mono text-[var(--color-text-primary)]">
                {result.stdout || '(なし)'}
              </pre>
              {result.stderr && (
                <pre className="mt-2 whitespace-pre-wrap break-words font-mono text-red-400">
                  {result.stderr}
                </pre>
              )}
            </td>
          </tr>
          <tr className="border-b border-surface-3">
            <th className="text-left px-4 py-2 font-medium text-[var(--color-text-muted)] bg-surface-2 align-top">
              期待する出力
            </th>
            <td className="px-4 py-2">
              <pre className="whitespace-pre-wrap break-words font-mono text-[var(--color-text-primary)]">{expected}</pre>
            </td>
          </tr>
        </tbody>
      </table>
      <div
        className={`px-4 py-2 text-xs font-semibold ${
          matched
            ? 'bg-green-500/15 text-green-400 border-t border-green-500/30'
            : 'bg-red-500/15 text-red-400 border-t border-red-500/30'
        }`}
      >
        コード実行結果: {matched ? '◎ 期待出力と一致' : '✗ 期待出力と不一致'}
      </div>
    </div>
  );
}

function SubmitResultPanel({ results }: { results: ExerciseTestCaseResult[] }) {
  return (
    <div className="rounded-lg border border-surface-3 bg-surface-1 p-4 space-y-2">
      <h3 className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wider">
        テストケース採点結果 ({results.filter((r) => r.passed).length}/{results.length} 合格)
      </h3>
      <div className="space-y-1.5">
        {results.map((r) => (
          <TestCaseResultRow key={r.orderIndex} r={r} />
        ))}
      </div>
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
        <span className={r.passed ? 'text-green-400' : 'text-red-400'}>{r.passed ? '合格' : '不合格'}</span>
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
  const stamp = `${date.getFullYear()}/${pad(date.getMonth() + 1)}/${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
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

function pad(n: number) {
  return String(n).padStart(2, '0');
}

// バックエンドの normalizeOutput と同等の正規化（提出前確認の「期待値と一致」表示用）。
// 末尾の改行 / 空白を吸収して厳密一致判定する。
function normalizeOutput(s: string): string {
  return s.replace(/\r\n/g, '\n').replace(/\r/g, '\n').replace(/[\s]+$/, '');
}
