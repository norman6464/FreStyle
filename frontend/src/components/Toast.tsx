import { useEffect } from 'react';
import {
  CheckCircleIcon,
  ExclamationCircleIcon,
  InformationCircleIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';

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
  // 既存の色設計（emerald / rose / primary）を維持しつつ、 ライトテーマでも
  // 読めるように背景は淡色 + 文字は濃色のコントラスト。
  success: 'border-emerald-400 bg-emerald-50 text-emerald-700',
  error: 'border-rose-400 bg-rose-50 text-rose-700',
  info: 'border-taupe-400 bg-taupe-50 text-taupe-700',
};

const ICON_COLOR_MAP = {
  success: 'text-emerald-500',
  error: 'text-rose-500',
  info: 'text-taupe-500',
};

/**
 * Toast — 画面上部から落ちてバウンドする通知。
 *
 * 配置は ToastContainer 側（fixed top center）。 本コンポーネントは見た目と
 * 4 秒オートクローズ + アニメーションのみ責任を持つ。
 */
export default function Toast({ type, message, onClose }: ToastProps) {
  useEffect(() => {
    const timer = setTimeout(onClose, 4000);
    return () => clearTimeout(timer);
  }, [onClose]);

  const Icon = ICON_MAP[type];

  return (
    <div
      role="alert"
      className={`pointer-events-auto flex items-start gap-3 px-5 py-3.5 rounded-lg border shadow-xl min-w-[280px] max-w-md ${COLOR_MAP[type]} animate-toast-drop`}
    >
      <Icon className={`w-6 h-6 flex-shrink-0 ${ICON_COLOR_MAP[type]}`} />
      <p className="text-sm leading-relaxed flex-1">{message}</p>
      <button
        type="button"
        onClick={onClose}
        aria-label="閉じる"
        className="-mr-1 -mt-0.5 p-1 rounded hover:bg-black/5 transition-colors"
      >
        <XMarkIcon className="w-4 h-4" />
      </button>
    </div>
  );
}
