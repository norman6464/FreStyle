import { lazy, Suspense } from 'react';
import { usePhpEditor } from '../hooks/usePhpEditor';
import { PhpExercise } from '../types';

const CodeEditor = lazy(() => import('../components/CodeEditor'));

/** CodeEditorPage — PHP コード実行環境ページ。 */
export default function CodeEditorPage() {
  const {
    exercises,
    categories,
    selectedExercise,
    code,
    setCode,
    result,
    running,
    showHint,
    setShowHint,
    loadingExercises,
    error,
    selectExercise,
    runCode,
    resetCode,
  } = usePhpEditor();

  if (loadingExercises) {
    return (
      <div className="flex items-center justify-center h-full">
        <span className="text-[var(--color-text-secondary)]">読み込み中...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <span className="text-red-500">{error}</span>
      </div>
    );
  }

  return (
    <div className="flex h-full overflow-hidden">
      {/* 左ペイン: 演習リスト */}
      <aside className="w-64 flex-shrink-0 border-r border-[var(--color-border)] overflow-y-auto bg-[var(--color-surface-1)]">
        <div className="p-4 border-b border-[var(--color-border)]">
          <h2 className="font-semibold text-[var(--color-text-primary)]">演習問題</h2>
          <p className="text-xs text-[var(--color-text-secondary)] mt-1">
            {exercises.length} 問
          </p>
        </div>
        {categories.map((category) => (
          <div key={category}>
            <div className="px-4 py-2 text-xs font-semibold text-[var(--color-text-secondary)] uppercase tracking-wider bg-[var(--color-surface-2)]">
              {category}
            </div>
            {exercises
              .filter((ex) => ex.category === category)
              .map((ex) => (
                <ExerciseItem
                  key={ex.id}
                  exercise={ex}
                  isSelected={selectedExercise?.id === ex.id}
                  onClick={() => selectExercise(ex)}
                />
              ))}
          </div>
        ))}
      </aside>

      {/* 中央 + 右ペイン */}
      <div className="flex flex-col flex-1 overflow-hidden">
        {/* ヘッダー */}
        {selectedExercise && (
          <div className="flex-shrink-0 px-6 py-4 border-b border-[var(--color-border)] bg-[var(--color-surface-1)]">
            <div className="flex items-start justify-between gap-4">
              <div>
                <span className="text-xs text-[var(--color-text-secondary)] bg-[var(--color-surface-3)] px-2 py-0.5 rounded mr-2">
                  {selectedExercise.category}
                </span>
                <h1 className="inline font-semibold text-lg text-[var(--color-text-primary)]">
                  {selectedExercise.title}
                </h1>
                <p className="mt-2 text-sm text-[var(--color-text-secondary)] whitespace-pre-wrap">
                  {selectedExercise.description}
                </p>
              </div>
              <button
                onClick={() => setShowHint(!showHint)}
                className="flex-shrink-0 text-xs px-3 py-1.5 rounded border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-surface-3)] transition-colors"
              >
                {showHint ? 'ヒントを隠す' : 'ヒントを見る'}
              </button>
            </div>
            {showHint && selectedExercise.hintText && (
              <div className="mt-3 p-3 bg-blue-500/10 border border-blue-500/20 rounded-lg text-sm text-[var(--color-text-primary)]">
                💡 {selectedExercise.hintText}
              </div>
            )}
          </div>
        )}

        {/* エディタ + 実行結果 */}
        <div className="flex flex-1 overflow-hidden">
          {/* エディタ */}
          <div className="flex flex-col flex-1 overflow-hidden">
            <div className="flex items-center justify-between px-4 py-2 bg-[var(--color-surface-2)] border-b border-[var(--color-border)]">
              <span className="text-xs text-[var(--color-text-secondary)] font-mono">PHP</span>
              <div className="flex gap-2">
                <button
                  onClick={resetCode}
                  className="text-xs px-3 py-1 rounded text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-3)] transition-colors"
                >
                  リセット
                </button>
                <button
                  onClick={runCode}
                  disabled={running}
                  className="text-xs px-4 py-1 rounded bg-green-600 hover:bg-green-500 disabled:opacity-50 text-white font-medium transition-colors flex items-center gap-1.5"
                >
                  {running ? (
                    <>
                      <span className="inline-block w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                      実行中...
                    </>
                  ) : (
                    '▶ 実行'
                  )}
                </button>
              </div>
            </div>
            <div className="flex-1 overflow-hidden">
              <Suspense fallback={<div className="h-full bg-[var(--color-surface-1)]" />}>
                <CodeEditor
                  value={code}
                  onChange={setCode}
                  language="php"
                  height="100%"
                />
              </Suspense>
            </div>
          </div>

          {/* 出力パネル */}
          <div className="w-80 flex-shrink-0 flex flex-col border-l border-[var(--color-border)] bg-[var(--color-surface-1)]">
            <div className="flex items-center justify-between px-4 py-2 bg-[var(--color-surface-2)] border-b border-[var(--color-border)]">
              <span className="text-xs text-[var(--color-text-secondary)]">出力</span>
              {result && (
                <span
                  className={`text-xs px-2 py-0.5 rounded ${
                    result.exitCode === 0
                      ? 'bg-green-500/20 text-green-400'
                      : 'bg-red-500/20 text-red-400'
                  }`}
                >
                  {result.exitCode === 0 ? '✓ 成功' : `✗ エラー (${result.exitCode})`}
                </span>
              )}
            </div>
            <div className="flex-1 overflow-y-auto p-4 font-mono text-sm">
              {!result && !running && (
                <p className="text-[var(--color-text-secondary)] text-xs">
                  「▶ 実行」ボタンを押すと結果が表示されます
                </p>
              )}
              {result?.stdout && (
                <pre className="whitespace-pre-wrap text-[var(--color-text-primary)] break-words">
                  {result.stdout}
                </pre>
              )}
              {result?.stderr && (
                <pre className="whitespace-pre-wrap text-red-400 break-words mt-2">
                  {result.stderr}
                </pre>
              )}
            </div>

            {/* 期待される出力 */}
            {selectedExercise?.expectedOutput && (
              <div className="border-t border-[var(--color-border)] p-4">
                <p className="text-xs text-[var(--color-text-secondary)] mb-2">期待される出力</p>
                <pre className="text-xs font-mono text-[var(--color-text-secondary)] whitespace-pre-wrap break-words bg-[var(--color-surface-2)] p-2 rounded">
                  {selectedExercise.expectedOutput}
                </pre>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function ExerciseItem({
  exercise,
  isSelected,
  onClick,
}: {
  exercise: PhpExercise;
  isSelected: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={`w-full text-left px-4 py-2.5 text-sm transition-colors border-l-2 ${
        isSelected
          ? 'bg-[var(--color-surface-3)] border-l-blue-500 text-[var(--color-text-primary)]'
          : 'border-l-transparent text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-2)] hover:text-[var(--color-text-primary)]'
      }`}
    >
      <span className="text-xs text-[var(--color-text-secondary)] mr-1">
        #{exercise.orderIndex}
      </span>
      {exercise.title}
    </button>
  );
}
