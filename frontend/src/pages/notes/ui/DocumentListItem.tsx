import { memo } from 'react';
import { TrashIcon } from '@heroicons/react/24/outline';
import { formatDate } from '@/shared/lib/formatters';

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
 * DocumentListItem は /notes 一覧の 1 件。リッチ文書はサマリ（doc 本体なし）なので、
 * 本文プレビューやピン留めは持たず、タイトルと更新日だけを見せる。
 *
 * 選択用と削除用はネイティブ button を並列に置く（対話要素を入れ子にしない a11y 規則）。
 * 選択状態は aria-current で表す。<ul> の子として使う。
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
  const dateStr = formatDate(updatedAt);

  return (
    <li>
      <div
        className={`group flex items-stretch rounded-lg transition-colors ${
          isActive ? 'bg-surface-2' : 'hover:bg-surface-2'
        }`}
      >
        <button
          type="button"
          onClick={() => onSelect(id)}
          aria-current={isActive ? 'true' : undefined}
          aria-label={`ノート「${displayTitle}」を選択`}
          className="min-w-0 flex-1 px-3 py-2.5 text-left"
        >
          <p className="truncate text-sm font-medium text-[var(--color-text-primary)]">{displayTitle}</p>
          <time dateTime={updatedAt} className="mt-0.5 block text-xs text-[var(--color-text-muted)]">
            {dateStr}
          </time>
        </button>
        <button
          type="button"
          onClick={() => onDelete(id)}
          aria-label={`ノート「${displayTitle}」を削除`}
          className="flex items-center px-2 text-[var(--color-text-muted)] opacity-0 transition-opacity hover:text-[var(--color-text-primary)] focus:opacity-100 group-hover:opacity-100"
        >
          <TrashIcon className="h-4 w-4" />
        </button>
      </div>
    </li>
  );
});
