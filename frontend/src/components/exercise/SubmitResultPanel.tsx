import { CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline';
import { ExerciseTestCaseResult } from '../../types';

/**
 * 提出 後 の 採点 結果 を 「テストケース 単位」 で 折りたたみ 表示。
 * 各 行 (details) を 開く と 入力 / 期待 / 実出力 / stderr を 確認 できる。
 */
export default function SubmitResultPanel({ results }: { results: ExerciseTestCaseResult[] }) {
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
