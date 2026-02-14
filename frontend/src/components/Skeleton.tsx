/**
 * Skeletonコンポーネント
 *
 * <p>役割:</p>
 * <ul>
 *   <li>コンテンツの形に合わせたスケルトンUI</li>
 *   <li>より良いローディング体験の提供</li>
 * </ul>
 *
 * <p>Presentation Layer:</p>
 * <ul>
 *   <li>UIのみに集中</li>
 *   <li>再利用可能なコンポーネント</li>
 * </ul>
 */

interface SkeletonProps {
  /**
   * 幅（Tailwindクラスまたはピクセル値）
   */
  width?: string;
  /**
   * 高さ（Tailwindクラスまたはピクセル値）
   */
  height?: string;
  /**
   * 円形スケルトン
   */
  circle?: boolean;
  /**
   * カスタムクラス
   */
  className?: string;
  /**
   * 繰り返し回数（複数のスケルトンを表示）
   */
  count?: number;
}

export default function Skeleton({
  width = 'w-full',
  height = 'h-4',
  circle = false,
  className = '',
  count = 1
}: SkeletonProps) {
  const baseClasses = `bg-gradient-to-r from-slate-100 via-slate-200 to-slate-100 bg-[length:200%_100%] animate-skeleton ${circle ? 'rounded-full' : 'rounded'}`;

  const skeletonElement = (
    <div
      className={`${baseClasses} ${width} ${height} ${className}`}
      role="status"
      aria-label="読み込み中"
    />
  );

  if (count === 1) {
    return skeletonElement;
  }

  return (
    <>
      {Array.from({ length: count }).map((_, index) => (
        <div key={index} className="mb-2 last:mb-0">
          {skeletonElement}
        </div>
      ))}
    </>
  );
}

/**
 * カード型スケルトン
 */
export function SkeletonCard({ className = '' }: { className?: string }) {
  return (
    <div className={`bg-surface-1 rounded-xl border border-surface-3 p-4 ${className}`}>
      <div className="flex items-center gap-3 mb-3">
        <Skeleton circle width="w-10" height="h-10" />
        <div className="flex-1">
          <Skeleton width="w-24" height="h-4" className="mb-2" />
          <Skeleton width="w-16" height="h-3" />
        </div>
      </div>
      <Skeleton width="w-full" height="h-3" className="mb-2" />
      <Skeleton width="w-3/4" height="h-3" />
    </div>
  );
}

/**
 * リスト型スケルトン
 */
export function SkeletonList({ count = 3 }: { count?: number }) {
  return (
    <div className="space-y-2">
      {Array.from({ length: count }).map((_, index) => (
        <div key={index} className="flex items-center gap-3 p-3 bg-surface-1 rounded-lg border border-surface-3">
          <Skeleton circle width="w-8" height="h-8" />
          <div className="flex-1">
            <Skeleton width="w-32" height="h-4" className="mb-1.5" />
            <Skeleton width="w-20" height="h-3" />
          </div>
        </div>
      ))}
    </div>
  );
}
