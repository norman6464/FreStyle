import { useEffect } from 'react';
import { CheckCircleIcon, ExclamationCircleIcon, XMarkIcon } from '@heroicons/react/24/solid';

interface FormMessageProps {
  message: { type: 'error' | 'success'; text: string } | null;
  onDismiss?: () => void;
}

export default function FormMessage({ message, onDismiss }: FormMessageProps) {
  useEffect(() => {
    if (!message || !onDismiss) return;
    const timer = setTimeout(onDismiss, 5000);
    return () => clearTimeout(timer);
  }, [message, onDismiss]);

  if (!message) return null;

  const isError = message.type === 'error';

  return (
    <div
      role="alert"
      className={`mb-4 p-3 rounded-lg text-sm font-medium flex items-start gap-2 ${
        isError
          ? 'bg-rose-900/30 text-rose-400 border border-rose-800'
          : 'bg-emerald-900/30 text-emerald-400 border border-emerald-800'
      }`}
    >
      {isError ? (
        <ExclamationCircleIcon className="w-5 h-5 flex-shrink-0 mt-0.5" />
      ) : (
        <CheckCircleIcon className="w-5 h-5 flex-shrink-0 mt-0.5" />
      )}
      <span className="flex-1">{message.text}</span>
      {onDismiss && (
        <button
          type="button"
          onClick={onDismiss}
          aria-label="閉じる"
          className="flex-shrink-0 mt-0.5 hover:opacity-70 transition-opacity"
        >
          <XMarkIcon className="w-4 h-4" />
        </button>
      )}
    </div>
  );
}
