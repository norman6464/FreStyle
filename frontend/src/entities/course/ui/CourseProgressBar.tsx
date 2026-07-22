/**
 * CourseProgressBar はコース内の完了割合を示す進捗バー。
 * コース詳細の左パネルと、コース一覧のカード(FRESTYLE-98)で共用する。
 * 「どのくらい残っているか」を数えなくても分かるよう、進行中は残り章数を明記し、
 * 全章完了は「すべて完了」+ 濃い緑バーで示す(FRESTYLE-114)。
 */
export default function CourseProgressBar({ completed, total }: { completed: number; total: number }) {
  // 呼び出し元のデータ差(完了行の残骸等)で completed > total が来ても 100% 超のバー幅や
  // 「3/2」表示にならないよう、コンポーネント側でクランプして全呼び出し元を安全にする。
  const safeCompleted = Math.max(0, Math.min(completed, total));
  const pct = total > 0 ? Math.round((safeCompleted / total) * 100) : 0;
  const remaining = total - safeCompleted;
  const isComplete = total > 0 && remaining === 0;
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-xs text-[var(--color-text-muted)]">
        <span>学習の進捗</span>
        {isComplete ? (
          <span className="font-semibold text-emerald-600">すべて完了（{total} 章）</span>
        ) : (
          <span>
            {safeCompleted}/{total}（{pct}%・残り {remaining} 章）
          </span>
        )}
      </div>
      <div className="h-1.5 w-full rounded-full bg-surface-3 overflow-hidden">
        <div
          className={`h-full rounded-full transition-all ${isComplete ? 'bg-emerald-500' : 'bg-green-400'}`}
          style={{ width: `${pct}%` }}
          role="progressbar"
          aria-label="学習の進捗"
          aria-valuenow={pct}
          aria-valuemin={0}
          aria-valuemax={100}
        />
      </div>
    </div>
  );
}
