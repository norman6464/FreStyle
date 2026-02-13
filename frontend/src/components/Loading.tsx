/**
 * Loadingコンポーネント
 *
 * <p>役割:</p>
 * <ul>
 *   <li>統一されたローディング表示</li>
 *   <li>サイズバリエーション対応</li>
 *   <li>フルスクリーン/インラインモード対応</li>
 * </ul>
 *
 * <p>Presentation Layer:</p>
 * <ul>
 *   <li>UIのみに集中</li>
 *   <li>再利用可能なコンポーネント</li>
 * </ul>
 */

interface LoadingProps {
  /**
   * サイズ: small (16px), medium (32px), large (48px)
   */
  size?: 'small' | 'medium' | 'large';
  /**
   * フルスクリーンモード
   */
  fullscreen?: boolean;
  /**
   * 表示メッセージ
   */
  message?: string;
  /**
   * カスタムクラス
   */
  className?: string;
}

export default function Loading({
  size = 'medium',
  fullscreen = false,
  message,
  className = ''
}: LoadingProps) {
  const sizeClasses = {
    small: 'w-4 h-4 border-2',
    medium: 'w-8 h-8 border-2',
    large: 'w-12 h-12 border-3',
  };

  const spinner = (
    <div
      className={`${sizeClasses[size]} border-primary-200 border-t-primary-500 rounded-full animate-spin`}
      role="status"
      aria-label="読み込み中"
    />
  );

  if (fullscreen) {
    return (
      <div className="fixed inset-0 bg-white/80 backdrop-blur-sm flex items-center justify-center z-50">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 border-3 border-primary-200 border-t-primary-500 rounded-full animate-spin" />
          {message && (
            <p className="text-sm font-medium text-slate-700 animate-pulse">
              {message}
            </p>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className={`flex items-center justify-center ${className}`}>
      <div className="flex flex-col items-center gap-2">
        {spinner}
        {message && (
          <p className="text-xs text-slate-600">{message}</p>
        )}
      </div>
    </div>
  );
}
