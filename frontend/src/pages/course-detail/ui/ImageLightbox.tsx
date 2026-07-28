import { useEffect, useRef } from 'react';
import { XMarkIcon } from '@heroicons/react/24/outline';

/**
 * 本文内の画像をモーダルで拡大表示する(FRESTYLE-191)。
 * 背景クリック / 閉じるボタン / Esc キーで閉じる。開いたら閉じるボタンへフォーカスを移す。
 */
export default function ImageLightbox({
  src,
  alt,
  onClose,
}: {
  src: string;
  alt: string;
  onClose: () => void;
}) {
  const closeButtonRef = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    closeButtonRef.current?.focus();
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);
  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={alt || '画像の拡大表示'}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4 sm:p-10"
      onClick={onClose}
    >
      <button
        ref={closeButtonRef}
        type="button"
        onClick={onClose}
        aria-label="閉じる"
        className="absolute top-4 right-4 rounded-full bg-white/90 p-2 text-gray-700 shadow hover:bg-white transition-colors"
      >
        <XMarkIcon className="w-5 h-5" />
      </button>
      {/* 画像自体のクリックでは閉じない(誤タップで閉じるのを防ぐ)。背景クリックのみで閉じる。 */}
      <img
        src={src}
        alt={alt}
        className="max-h-full max-w-full rounded-lg bg-white shadow-xl"
        onClick={(event) => event.stopPropagation()}
      />
    </div>
  );
}
