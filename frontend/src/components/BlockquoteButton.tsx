import { ChatBubbleBottomCenterTextIcon } from '@heroicons/react/24/outline';

interface BlockquoteButtonProps {
  onBlockquote: () => void;
}

export default function BlockquoteButton({ onBlockquote }: BlockquoteButtonProps) {
  return (
    <button
      type="button"
      aria-label="引用"
      className="w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-faint)] transition-colors"
      onClick={onBlockquote}
    >
      <ChatBubbleBottomCenterTextIcon className="w-4 h-4" />
    </button>
  );
}
