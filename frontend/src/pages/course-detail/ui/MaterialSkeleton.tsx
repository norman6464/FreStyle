/**
 * 章本文の取得中に出す骨組み。
 *
 * バーのスピナーではなく記事レイアウト（タイトル / メタ / 本文行 / 図ブロック）を
 * 模した pulse プレースホルダにすることで、 章切り替え時のちらつきを抑え体感速度を上げる。
 */
export default function MaterialSkeleton() {
  return (
    // 実表示(灰青背景 + 白カード)と同じ配色にして、取得完了時の切り替わりで背景が変わらないようにする。
    <div className="flex-1 bg-[var(--color-reading-surface)]">
      <div
        className="mx-auto w-full max-w-[860px] px-6 sm:px-10 py-8 sm:py-10 animate-pulse"
        aria-hidden="true"
      >
        <div className="bg-white border border-surface-3 rounded-xl shadow-sm px-6 sm:px-10 py-8 sm:py-10">
          {/* タイトル + メタはカードの先頭(FRESTYLE-178 の実表示に合わせて跳ねを防ぐ)。 */}
          <div className="mb-6 pb-6 border-b border-surface-3 space-y-3">
            <div className="h-8 w-3/4 rounded bg-surface-3" />
            <div className="h-3 w-40 rounded bg-surface-2" />
          </div>
          <div className="space-y-3">
            <div className="h-4 w-full rounded bg-surface-2" />
            <div className="h-4 w-11/12 rounded bg-surface-2" />
            <div className="h-4 w-4/5 rounded bg-surface-2" />
          </div>
          <div className="mt-6 h-40 w-full rounded-lg bg-surface-2" />
          <div className="mt-6 space-y-3">
            <div className="h-4 w-full rounded bg-surface-2" />
            <div className="h-4 w-10/12 rounded bg-surface-2" />
            <div className="h-4 w-3/4 rounded bg-surface-2" />
          </div>
        </div>
      </div>
      <span className="sr-only" role="status">
        読み込み中
      </span>
    </div>
  );
}
