import { Suspense, useState } from 'react';
import { useParams } from 'react-router-dom';
import { ChevronDownIcon, ChevronUpIcon, ClipboardDocumentCheckIcon } from '@heroicons/react/24/outline';
import Loading from '../components/Loading';
import BackLink from '../components/exercise/BackLink';
import ExerciseHeader from '../components/exercise/ExerciseHeader';
import ExampleBlock from '../components/exercise/ExampleBlock';
import ExecutionResultTable from '../components/exercise/ExecutionResultTable';
import SubmitResultPanel from '../components/exercise/SubmitResultPanel';
import SubmissionRow from '../components/exercise/SubmissionRow';
import QaExerciseView from '../components/exercise/QaExerciseView';
import { useExerciseDetail } from '../hooks/useExerciseDetail';
import { lazyWithReload } from '../utils/lazyWithReload';
import { monacoLanguageOf } from '../utils/exerciseFormat';

const CodeEditor = lazyWithReload(() => import('../components/CodeEditor'), 'CodeEditor');

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
 * mode='qa' (docker / kubernetes など サンドボックス実行が困難な題材) は QaExerciseView に
 * 描画を委譲する。
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

  if (ex.mode === 'qa') {
    return (
      <QaExerciseView
        exercise={ex}
        starterCode={code}
        onCodeChange={setCode}
        submitting={submitting}
        submitResult={submitResult}
        submitError={submitError}
        onSubmit={submitCode}
        onReset={resetCode}
      />
    );
  }

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-4xl mx-auto space-y-6">
      <ExerciseHeader exercise={ex} submitResult={submitResult} />

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
          {detail.examples.length > 0
            ? detail.examples.map((example, idx) => (
                <ExampleBlock
                  key={example.id}
                  index={idx + 1}
                  total={detail.examples.length}
                  example={example}
                />
              ))
            : ex.expectedOutput && (
                // examples が登録されていない演習 (seed.py 経由の go / linux / git 等)
                // は exercise 自身の expectedOutput を 単一の入力例として表示する。
                <ExampleBlock
                  index={1}
                  total={1}
                  example={{
                    id: 0,
                    exerciseId: ex.id,
                    orderIndex: 1,
                    inputText: '',
                    expectedOutput: ex.expectedOutput,
                    createdAt: '',
                    updatedAt: '',
                  }}
                />
              )}
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
        <div className="bg-[#1e1e1e]">
          <Suspense fallback={<div style={{ height: 360 }} className="bg-[#1e1e1e]" />}>
            {/* autoGrow=true (default) で エディタの高さが行数に合わせて伸びる。 */}
            {/* ページ側でスクロールするため、 エディタ内部の縦スクロールが発生しない。 */}
            <CodeEditor
              value={code}
              onChange={setCode}
              language={monacoLanguageOf(ex.language)}
              minHeight={260}
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
          // examples が空の演習 (seed.py 経由で投入された linux/git/go 等) は
          // exercise 自身の expectedOutput を fallback で使う。
          // これがないと preview 比較が常に空文字 vs 空文字 で 「◎ 一致」 と出てしまい、
          // 提出時のバックエンド側 fallback と齟齬が出る。
          expected={detail.examples[0]?.expectedOutput ?? ex.expectedOutput ?? ''}
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
