/** メニューカード読み込み中のプレースホルダ。実レイアウトと寸法を揃えてちらつきを防ぐ。 */
export default function MenuSkeleton() {
  return (
    <div className="space-y-8" aria-hidden="true">
      {[2, 3].map((cardCount, sectionIndex) => (
        <section key={sectionIndex} className="space-y-3">
          <div className="h-3 w-16 rounded bg-[var(--color-surface-3)] animate-pulse" />
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {Array.from({ length: cardCount }).map((_, cardIndex) => (
              <div
                key={cardIndex}
                className="h-[148px] rounded-xl border border-[var(--color-surface-3)] bg-[var(--color-surface-1)] p-5"
              >
                <div className="w-9 h-9 rounded-lg bg-[var(--color-surface-3)] animate-pulse" />
                <div className="mt-3 h-3.5 w-24 rounded bg-[var(--color-surface-3)] animate-pulse" />
                <div className="mt-3 h-3 w-full rounded bg-[var(--color-surface-3)] animate-pulse" />
              </div>
            ))}
          </div>
        </section>
      ))}
    </div>
  );
}
