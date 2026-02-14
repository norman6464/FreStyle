import { useCallback, useEffect, useRef } from 'react';
import { TrashIcon, QuestionMarkCircleIcon } from '@heroicons/react/24/outline';

interface ConfirmModalProps {
  isOpen: boolean;
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  onConfirm: () => void;
  onCancel: () => void;
  isDanger?: boolean;
}

export default function ConfirmModal({
  isOpen,
  title = '確認',
  message,
  confirmText = '削除',
  cancelText = 'キャンセル',
  onConfirm,
  onCancel,
  isDanger = true,
}: ConfirmModalProps) {
  const cancelRef = useRef<HTMLButtonElement>(null);
  const confirmRef = useRef<HTMLButtonElement>(null);

  // ESCキーで閉じる
  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onCancel();
      }
    };
    window.addEventListener('keydown', handleEsc);
    return () => window.removeEventListener('keydown', handleEsc);
  }, [isOpen, onCancel]);

  // 自動フォーカス
  useEffect(() => {
    if (isOpen) {
      cancelRef.current?.focus();
    }
  }, [isOpen]);

  // フォーカストラップ
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key !== 'Tab') return;

    const focusable = [cancelRef.current, confirmRef.current].filter(Boolean) as HTMLElement[];
    if (focusable.length === 0) return;

    const currentIndex = focusable.indexOf(document.activeElement as HTMLElement);

    if (e.shiftKey) {
      e.preventDefault();
      const prevIndex = currentIndex <= 0 ? focusable.length - 1 : currentIndex - 1;
      focusable[prevIndex].focus();
    } else {
      e.preventDefault();
      const nextIndex = currentIndex >= focusable.length - 1 ? 0 : currentIndex + 1;
      focusable[nextIndex].focus();
    }
  }, []);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* オーバーレイ */}
      <div
        className="absolute inset-0 bg-black/50 animate-fade-in"
        onClick={onCancel}
      />

      {/* モーダル */}
      <div role="dialog" aria-modal="true" aria-labelledby="confirm-modal-title" onKeyDown={handleKeyDown} className="relative bg-surface-1 rounded-2xl shadow-md p-6 mx-4 max-w-sm w-full animate-fade-in">
        {/* アイコン */}
        <div className="flex justify-center mb-4">
          <div
            className={`w-14 h-14 rounded-full flex items-center justify-center ${
              isDanger ? 'bg-red-900/30' : 'bg-surface-3'
            }`}
          >
            {isDanger ? (
              <TrashIcon className="w-7 h-7 text-red-600" />
            ) : (
              <QuestionMarkCircleIcon className="w-7 h-7 text-primary-400" />
            )}
          </div>
        </div>

        {/* タイトル */}
        <h3 id="confirm-modal-title" className="text-xl font-bold text-[var(--color-text-primary)] text-center mb-2">
          {title}
        </h3>

        {/* メッセージ */}
        <p className="text-[var(--color-text-tertiary)] text-center mb-6">{message}</p>

        {/* ボタン */}
        <div className="flex gap-3">
          <button
            ref={cancelRef}
            onClick={onCancel}
            className="flex-1 px-4 py-2.5 bg-surface-3 hover:bg-surface-3 text-[var(--color-text-secondary)] font-medium rounded-xl transition-colors duration-150"
          >
            {cancelText}
          </button>
          <button
            ref={confirmRef}
            onClick={onConfirm}
            className={`flex-1 px-4 py-2.5 font-medium rounded-xl transition-colors duration-150 ${
              isDanger
                ? 'bg-red-500 hover:bg-red-600 text-white'
                : 'bg-primary-500 hover:bg-primary-600 text-white'
            }`}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
}
