import Card from './Card';
import { formatDate } from '../utils/formatters';
import { useRecentNotes } from '../hooks/useRecentNotes';

export default function RecentNotesCard() {
  const { notes: sortedNotes, totalCount } = useRecentNotes(3);

  if (totalCount === 0) return null;

  return (
    <Card>
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-[var(--color-text-secondary)]">最近のメモ</p>
        <span className="text-[10px] text-[var(--color-text-faint)]">{totalCount}件</span>
      </div>

      <div className="space-y-2">
        {sortedNotes.map((entry) => (
          <div
            key={entry.sessionId}
            className="bg-surface-2 rounded p-2"
          >
            <p className="text-xs text-[var(--color-text-secondary)] line-clamp-2">{entry.note}</p>
            <p className="text-[10px] text-[var(--color-text-faint)] mt-1">
              {formatDate(entry.updatedAt)}
            </p>
          </div>
        ))}
      </div>
    </Card>
  );
}
