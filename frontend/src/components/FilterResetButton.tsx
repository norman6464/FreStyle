import { ArrowPathIcon } from '@heroicons/react/24/outline';

interface FilterResetButtonProps {
  isActive: boolean;
  onReset: () => void;
}

export default function FilterResetButton({ isActive, onReset }: FilterResetButtonProps) {
  if (!isActive) return null;

  return (
    <button
      onClick={onReset}
      className="flex items-center gap-1 text-[10px] font-medium text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] transition-colors"
    >
      <ArrowPathIcon className="w-3 h-3" />
      リセット
    </button>
  );
}
