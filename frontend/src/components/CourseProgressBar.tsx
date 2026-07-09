/**
 * CourseProgressBar はコース内の完了割合を示す進捗バー。
 * コース詳細の左パネルと、コース一覧のカード(FRESTYLE-98)で共用する。
 */
export default function CourseProgressBar({ completed, total }: { completed: number; total: number }) {
  // 呼び出し元のデータ差(完了行の残骸等)で completed > total が来ても 100% 超のバー幅や
  // 「3/2」表示にならないよう、コンポーネント側でクランプして全呼び出し元を安全にする。
  const safeCompleted = Math.max(0, Math.min(completed, total));
  const pct = total > 0 ? Math.round((safeCompleted / total) * 100) : 0;
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-xs text-[var(--color-text-muted)]">
        <span>学習の進捗</span>
        <span>
          {safeCompleted}/{total}（{pct}%）
        </span>
      </div>
      <div className="h-1.5 w-full rounded-full bg-surface-3 overflow-hidden">
        <div
          className="h-full rounded-full bg-green-400 transition-all"
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
