import { ArrowRightIcon, ArrowLeftIcon } from '@heroicons/react/24/outline';

interface IndentButtonsProps {
  onIndent: () => void;
  onOutdent: () => void;
}

export default function IndentButtons({ onIndent, onOutdent }: IndentButtonsProps) {
  return (
    <div className="flex items-center gap-0.5">
      <button
        type="button"
        aria-label="インデント減少"
        className="w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-faint)] transition-colors"
        onClick={onOutdent}
      >
        <ArrowLeftIcon className="w-4 h-4" />
      </button>
      <button
        type="button"
        aria-label="インデント増加"
        className="w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-faint)] transition-colors"
        onClick={onIndent}
      >
        <ArrowRightIcon className="w-4 h-4" />
      </button>
    </div>
  );
}
