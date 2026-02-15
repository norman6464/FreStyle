import { PencilIcon, TrashIcon } from '@heroicons/react/24/outline';

interface LinkBubbleMenuProps {
  url: string;
  onEdit: () => void;
  onRemove: () => void;
}

export default function LinkBubbleMenu({ url, onEdit, onRemove }: LinkBubbleMenuProps) {
  return (
    <div className="flex items-center gap-1 bg-[var(--color-surface-1)] border border-[var(--color-border)] rounded-lg shadow-lg px-2 py-1.5 text-sm">
      <a
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className="text-blue-400 hover:text-blue-300 truncate max-w-[200px]"
      >
        {url}
      </a>
      <button
        type="button"
        aria-label="リンクを編集"
        className="p-1 rounded hover:bg-[var(--color-surface-2)] text-[var(--color-text-secondary)]"
        onClick={onEdit}
      >
        <PencilIcon className="w-3.5 h-3.5" />
      </button>
      <button
        type="button"
        aria-label="リンクを削除"
        className="p-1 rounded hover:bg-[var(--color-surface-2)] text-[var(--color-text-secondary)]"
        onClick={onRemove}
      >
        <TrashIcon className="w-3.5 h-3.5" />
      </button>
    </div>
  );
}
