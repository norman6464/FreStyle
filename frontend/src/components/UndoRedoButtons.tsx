import { ArrowUturnLeftIcon, ArrowUturnRightIcon } from '@heroicons/react/24/outline';

interface UndoRedoButtonsProps {
  onUndo: () => void;
  onRedo: () => void;
}

export default function UndoRedoButtons({ onUndo, onRedo }: UndoRedoButtonsProps) {
  return (
    <div className="flex items-center gap-0.5">
      <button
        type="button"
        aria-label="元に戻す"
        className="w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-faint)] transition-colors"
        onClick={onUndo}
      >
        <ArrowUturnLeftIcon className="w-4 h-4" />
      </button>
      <button
        type="button"
        aria-label="やり直す"
        className="w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-faint)] transition-colors"
        onClick={onRedo}
      >
        <ArrowUturnRightIcon className="w-4 h-4" />
      </button>
    </div>
  );
}
