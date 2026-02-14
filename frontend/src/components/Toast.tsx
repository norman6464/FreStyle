import { useEffect } from 'react';
import { CheckCircleIcon, ExclamationCircleIcon, InformationCircleIcon } from '@heroicons/react/24/outline';

export type ToastType = 'success' | 'error' | 'info';

interface ToastProps {
  type: ToastType;
  message: string;
  onClose: () => void;
}

const ICON_MAP = {
  success: CheckCircleIcon,
  error: ExclamationCircleIcon,
  info: InformationCircleIcon,
};

const COLOR_MAP = {
  success: 'border-emerald-500 bg-emerald-900/20 text-emerald-400',
  error: 'border-rose-500 bg-rose-900/20 text-rose-400',
  info: 'border-primary-500 bg-primary-900/20 text-primary-400',
};

export default function Toast({ type, message, onClose }: ToastProps) {
  useEffect(() => {
    const timer = setTimeout(onClose, 3000);
    return () => clearTimeout(timer);
  }, [onClose]);

  const Icon = ICON_MAP[type];

  return (
    <div
      role="alert"
      className={`flex items-center gap-2 px-4 py-3 rounded-lg border shadow-lg ${COLOR_MAP[type]} animate-fade-in`}
    >
      <Icon className="w-5 h-5 flex-shrink-0" />
      <p className="text-sm">{message}</p>
    </div>
  );
}
