import { memo } from 'react';
import { TrashIcon } from '@heroicons/react/24/outline';
import { MapPinIcon as MapPinOutline } from '@heroicons/react/24/outline';
import { MapPinIcon as MapPinSolid } from '@heroicons/react/24/solid';
import { tiptapToPlainText } from '../utils/tiptapToPlainText';
import { formatMonthDay } from '../utils/formatters';

interface NoteListItemProps {
  noteId: string;
  title: string;
  content: string;
  updatedAt: number;
  isPinned: boolean;
  isActive: boolean;
  onSelect: (noteId: string) => void;
  onDelete: (noteId: string) => void;
  onTogglePin: (noteId: string) => void;
}

export default memo(function NoteListItem({
  noteId,
  title,
  content,
  updatedAt,
  isPinned,
  isActive,
  onSelect,
  onDelete,
  onTogglePin,
}: NoteListItemProps) {
  const displayTitle = title || '無題';
  const preview = tiptapToPlainText(content).replace(/\n/g, ' ').slice(0, 60);
  const dateStr = formatMonthDay(updatedAt);

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete(noteId);
  };

  const handleTogglePin = (e: React.MouseEvent) => {
    e.stopPropagation();
    onTogglePin(noteId);
  };

  const PinIcon = isPinned ? MapPinSolid : MapPinOutline;

  return (
    <div
      role="button"
      tabIndex={0}
      aria-label={`ノート「${displayTitle}」を選択`}
      aria-pressed={isActive}
      onClick={() => onSelect(noteId)}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(noteId); } }}
      className={`w-full text-left px-3 py-2.5 rounded-lg transition-colors group cursor-pointer ${
        isActive ? 'bg-surface-2' : 'hover:bg-surface-2'
      }`}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-[var(--color-text-primary)] truncate">
            {isPinned && <PinIcon className="w-3.5 h-3.5 inline mr-1 text-primary-500" />}
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
        <div className="flex items-center gap-0.5">
          <button
            onClick={handleTogglePin}
            aria-label={isPinned ? 'ピン留め解除' : 'ピン留め'}
            className={`p-1 rounded transition-all ${
              isPinned ? 'opacity-100 text-primary-500' : 'opacity-0 group-hover:opacity-100 hover:bg-surface-3'
            }`}
          >
            <PinIcon className="w-3.5 h-3.5" />
          </button>
          <button
            onClick={handleDelete}
            aria-label="ノートを削除"
            className="p-1 opacity-0 group-hover:opacity-100 hover:bg-surface-3 rounded transition-all"
          >
            <TrashIcon className="w-3.5 h-3.5 text-[var(--color-text-muted)]" />
          </button>
        </div>
      </div>
    </div>
  );
});
