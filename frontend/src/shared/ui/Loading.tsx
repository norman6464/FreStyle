interface LoadingProps {
  size?: 'small' | 'medium' | 'large';
  fullscreen?: boolean;
  message?: string;
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
    medium: 'w-8 h-8 border-3',
    large: 'w-10 h-10 border-4',
  };

  const spinner = (
    <div
      className={`${sizeClasses[size]} border-gray-200 border-t-blue-500 rounded-full animate-spin`}
      role="status"
      aria-label="読み込み中"
    />
  );

  if (fullscreen) {
    return (
      <div className="fixed inset-0 bg-white flex flex-col items-center justify-center z-50">
        <div className="w-10 h-10 border-4 border-gray-200 border-t-blue-500 rounded-full animate-spin" role="status" aria-label="読み込み中" />
        {message && (
          <p className="mt-4 text-sm text-gray-500">{message}</p>
        )}
      </div>
    );
  }

  return (
    <div className={`flex items-center justify-center ${className}`}>
      <div className="flex flex-col items-center gap-3">
        {spinner}
        {message && (
          <p className="text-sm text-gray-500">{message}</p>
        )}
      </div>
    </div>
  );
}
