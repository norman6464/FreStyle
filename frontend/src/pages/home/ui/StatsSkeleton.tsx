/** 右サイドバー（学習統計）読み込み中のプレースホルダ。 */
export default function StatsSkeleton() {
  return (
    <div
      className="space-y-4 p-4 rounded-xl border border-[var(--color-surface-3)] bg-[var(--color-surface-1)]"
      aria-hidden="true"
    >
      <div className="grid grid-cols-2 gap-3">
        {Array.from({ length: 4 }).map((_, index) => (
          <div key={index} className="h-[88px] rounded-lg bg-[var(--color-surface-2)] animate-pulse" />
        ))}
      </div>
      <div className="space-y-2">
        <div className="h-3 w-40 rounded bg-[var(--color-surface-3)] animate-pulse" />
        <div className="h-20 w-full rounded bg-[var(--color-surface-2)] animate-pulse" />
      </div>
    </div>
  );
}
