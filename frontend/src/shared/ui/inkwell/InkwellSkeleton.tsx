export interface InkwellSkeletonProps {
  /** 形状。text=丸角の行 / rect=矩形 / circle=円。 */
  variant?: 'text' | 'rect' | 'circle';
  width?: number | string;
  height?: number | string;
  className?: string;
}

/**
 * 読み込み中のプレースホルダ。左右に光沢が流れる（背景 position のみ＝layout に触れない）。
 * reduced-motion では流れを止め、静的な淡色ブロックにする。
 */
export default function InkwellSkeleton({
  variant = 'text',
  width,
  height,
  className = '',
}: InkwellSkeletonProps) {
  const shape =
    variant === 'circle' ? 'rounded-full' : variant === 'rect' ? 'rounded' : 'rounded';
  const defaultSize =
    variant === 'text' ? 'h-4 w-full' : variant === 'circle' ? 'h-10 w-10' : 'h-24 w-full';

  return (
    <span
      role="status"
      aria-label="読み込み中"
      style={{ width, height }}
      className={`block ${defaultSize} ${shape} bg-[linear-gradient(90deg,rgba(0,0,0,0.06)_25%,rgba(0,0,0,0.12)_37%,rgba(0,0,0,0.06)_63%)] bg-[length:400%_100%] animate-skeleton motion-reduce:animate-none ${className}`}
    />
  );
}
