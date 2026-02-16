import { memo } from 'react';
import { PencilSquareIcon, TrashIcon, CheckIcon, XMarkIcon } from '@heroicons/react/24/outline';
import { formatDate } from '../utils/formatters';

interface AiSessionListItemProps {
  id: number;
  title?: string;
  createdAt?: string;
  isActive: boolean;
  isEditing: boolean;
  editingTitle: string;
  onSelect: (id: number) => void;
  onStartEdit: (session: { id: number; title: string }) => void;
  onDelete: (id: number) => void;
  onSaveTitle: (id: number) => void;
  onCancelEdit: () => void;
  onEditingTitleChange: (title: string) => void;
}

export default memo(function AiSessionListItem({
  id,
  title,
  createdAt,
  isActive,
  isEditing,
  editingTitle,
  onSelect,
  onStartEdit,
  onDelete,
  onSaveTitle,
  onCancelEdit,
  onEditingTitleChange,
}: AiSessionListItemProps) {
  return (
    <div
      role="button"
      tabIndex={0}
      aria-label={title || '新しいチャット'}
      className={`group flex items-center justify-between px-3 py-2.5 rounded-lg cursor-pointer transition-colors ${
        isActive
          ? 'bg-surface-2 text-primary-300'
          : 'hover:bg-surface-2'
      }`}
      onClick={() => !isEditing && onSelect(id)}
      onKeyDown={(e) => { if ((e.key === 'Enter' || e.key === ' ') && !isEditing) { e.preventDefault(); onSelect(id); } }}
    >
      <div className="flex-1 min-w-0">
        {isEditing ? (
          <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
            <input
              type="text"
              value={editingTitle}
              onChange={(e) => onEditingTitleChange(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') onSaveTitle(id);
                if (e.key === 'Escape') onCancelEdit();
              }}
              className="flex-1 text-xs px-2 py-1 border border-[var(--color-border-hover)] rounded focus:outline-none focus:ring-1 focus:ring-primary-400"
              autoFocus
            />
            <button onClick={() => onSaveTitle(id)} className="p-0.5 hover:bg-green-900/30 rounded" aria-label="保存">
              <CheckIcon className="w-3.5 h-3.5 text-green-400" />
            </button>
            <button onClick={onCancelEdit} className="p-0.5 hover:bg-surface-3 rounded" aria-label="キャンセル">
              <XMarkIcon className="w-3.5 h-3.5 text-[var(--color-text-muted)]" />
            </button>
          </div>
        ) : (
          <>
            <p className="text-sm font-medium truncate">{title || '新しいチャット'}</p>
            <p className="text-[11px] text-[var(--color-text-muted)]">
              {createdAt ? formatDate(createdAt) : ''}
            </p>
          </>
        )}
      </div>
      {!isEditing && (
        <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={(e) => { e.stopPropagation(); onStartEdit({ id, title }); }}
            className="p-1 hover:bg-blue-900/30 rounded"
            title="タイトルを編集"
            aria-label="タイトルを編集"
          >
            <PencilSquareIcon className="w-3.5 h-3.5 text-blue-500" />
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onDelete(id);
            }}
            className="p-1 hover:bg-rose-900/30 rounded"
            title="削除"
            aria-label="削除"
          >
            <TrashIcon className="w-3.5 h-3.5 text-rose-500" />
          </button>
        </div>
      )}
    </div>
  );
});
