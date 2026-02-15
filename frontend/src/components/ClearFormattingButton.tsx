import { XMarkIcon } from '@heroicons/react/24/outline';

interface ClearFormattingButtonProps {
  onClearFormatting: () => void;
}

export default function ClearFormattingButton({ onClearFormatting }: ClearFormattingButtonProps) {
  return (
    <button
      type="button"
      aria-label="書式クリア"
      className="w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-faint)] transition-colors"
      onClick={onClearFormatting}
    >
      <XMarkIcon className="w-4 h-4" />
    </button>
  );
}
