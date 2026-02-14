import { TrashIcon } from '@heroicons/react/24/outline';

interface NoteListItemProps {
  noteId: string;
  title: string;
  content: string;
  updatedAt: number;
  isPinned: boolean;
  isActive: boolean;
  onSelect: (noteId: string) => void;
  onDelete: (noteId: string) => void;
}

export default function NoteListItem({
  noteId,
  title,
  content,
  updatedAt,
  isActive,
  onSelect,
  onDelete,
}: NoteListItemProps) {
  const displayTitle = title || '無題';
  const preview = content.replace(/\n/g, ' ').slice(0, 60);
  const date = new Date(updatedAt);
  const dateStr = `${date.getMonth() + 1}/${date.getDate()}`;

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete(noteId);
  };

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => onSelect(noteId)}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(noteId); } }}
      className={`w-full text-left px-3 py-2.5 rounded-lg transition-colors group cursor-pointer ${
        isActive ? 'bg-surface-2' : 'hover:bg-surface-2'
      }`}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-[var(--color-text-primary)] truncate">
            {displayTitle}
          </p>
          {preview && (
            <p className="text-xs text-[var(--color-text-muted)] truncate mt-0.5">
              {preview}
            </p>
          )}
          <p className="text-[11px] text-[var(--color-text-faint)] mt-1">
            {dateStr}
          </p>
        </div>
        <button
          onClick={handleDelete}
          aria-label="ノートを削除"
          className="p-1 opacity-0 group-hover:opacity-100 hover:bg-surface-3 rounded transition-all"
        >
          <TrashIcon className="w-3.5 h-3.5 text-[var(--color-text-muted)]" />
        </button>
      </div>
    </div>
  );
}
