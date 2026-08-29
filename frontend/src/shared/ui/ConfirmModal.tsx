import { useCallback, useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
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

  // document.body へ出す。呼び出し元の DOM の中に置くと、全面のオーバーレイが
  // その要素の子孫になり、モーダルの上での右クリック・ドラッグ・クリックが
  // 呼び出し元のハンドラへ伝わってしまう（行の右クリックメニューが背後で開く等）。
  // 親の opacity / overflow / transform の影響も受けなくなる。
  return createPortal(
    // React のポータルは DOM を移すだけで、**イベントは React のツリーを辿って
    // 呼び出し元へ伝わる**。モーダルの上での操作が背後の行や一覧に届くと、
    // 右クリックでメニューが背後に開く・行のドラッグが始まる、といった取り違えになる。
    // ここで止めて、モーダルが開いている間の入力はモーダルだけのものにする。
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      onClick={(event) => event.stopPropagation()}
      onContextMenu={(event) => event.stopPropagation()}
      onDragStart={(event) => event.stopPropagation()}
      onMouseDown={(event) => event.stopPropagation()}
    >
      {/* オーバーレイ */}
      <div
        className="absolute inset-0 bg-black/50 animate-fade-in"
        onClick={onCancel}
      />

      {/* モーダル */}
      <div role="dialog" aria-modal="true" aria-labelledby="confirm-modal-title" onKeyDown={handleKeyDown} className="relative bg-surface-1 rounded-2xl shadow-md p-6 mx-4 max-w-sm w-full animate-scale-in">
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
              <QuestionMarkCircleIcon className="w-7 h-7 text-taupe-400" />
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
            className="flex-1 px-4 py-2.5 bg-surface-3 hover:bg-surface-2 text-[var(--color-text-secondary)] font-medium rounded-xl transition-colors duration-150"
          >
            {cancelText}
          </button>
          <button
            ref={confirmRef}
            onClick={onConfirm}
            // 白文字に対して 500 番は 3.7:1 で、読みやすさの基準（4.5:1）に届かない。
            // 600 番（赤 4.8:1 / 青 5.2:1）から始める。
            className={`flex-1 px-4 py-2.5 font-medium rounded-xl transition-colors duration-150 ${
              isDanger
                ? 'bg-red-600 hover:bg-red-700 text-white'
                : 'bg-brand-600 hover:bg-brand-700 text-white'
            }`}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>,
    document.body,
  );
}
