import { SessionNoteRepository } from '../repositories/SessionNoteRepository';
import Card from './Card';

export default function RecentNotesCard() {
  const allNotes = SessionNoteRepository.getAll();
  const noteEntries = Object.values(allNotes);

  if (noteEntries.length === 0) return null;

  const sortedNotes = [...noteEntries]
    .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
    .slice(0, 3);

  return (
    <Card>
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-[var(--color-text-secondary)]">最近のメモ</p>
        <span className="text-[10px] text-[var(--color-text-faint)]">{noteEntries.length}件</span>
      </div>

      <div className="space-y-2">
        {sortedNotes.map((entry) => (
          <div
            key={entry.sessionId}
            className="bg-surface-2 rounded p-2"
          >
            <p className="text-xs text-[var(--color-text-secondary)] line-clamp-2">{entry.note}</p>
            <p className="text-[10px] text-[var(--color-text-faint)] mt-1">
              {new Date(entry.updatedAt).toLocaleDateString('ja-JP')}
            </p>
          </div>
        ))}
      </div>
    </Card>
  );
}
