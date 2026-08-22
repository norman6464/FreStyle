import { memo } from 'react';
import { TrashIcon } from '@heroicons/react/24/outline';
import { formatMonthDay } from '@/shared/lib/formatters';

interface DocumentListItemProps {
  id: string;
  title: string;
  /** backend は RFC3339 string を返す。 */
  updatedAt: string;
  isActive: boolean;
  onSelect: (id: string) => void;
  onDelete: (id: string) => void;
}

/**
 * DocumentListItem は /notes の一覧 1 件。リッチ文書はサマリ（doc 本体なし）なので、
 * 本文プレビューやピン留めは持たず、タイトルと更新日だけを見せる。
 */
export default memo(function DocumentListItem({
  id,
  title,
  updatedAt,
  isActive,
  onSelect,
  onDelete,
}: DocumentListItemProps) {
  const displayTitle = title || '無題';
  const dateStr = formatMonthDay(updatedAt);

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete(id);
  };

  return (
    <div
      role="button"
      tabIndex={0}
      aria-label={`ノート「${displayTitle}」を選択`}
      aria-pressed={isActive}
      onClick={() => onSelect(id)}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onSelect(id);
        }
      }}
      className={`w-full text-left px-3 py-2.5 rounded-lg transition-colors group cursor-pointer ${
        isActive ? 'bg-surface-2' : 'hover:bg-surface-2'
      }`}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium text-[var(--color-text-primary)]">{displayTitle}</p>
          <p className="mt-0.5 text-xs text-[var(--color-text-muted)]">{dateStr}</p>
        </div>
        <button
          type="button"
          onClick={handleDelete}
          aria-label={`ノート「${displayTitle}」を削除`}
          className="rounded p-1 text-[var(--color-text-muted)] opacity-0 transition-opacity hover:bg-surface-3 hover:text-[var(--color-text-primary)] focus:opacity-100 group-hover:opacity-100"
        >
          <TrashIcon className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
});
