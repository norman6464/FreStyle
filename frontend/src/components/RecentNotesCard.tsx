import { SessionNoteRepository } from '../repositories/SessionNoteRepository';

export default function RecentNotesCard() {
  const allNotes = SessionNoteRepository.getAll();
  const noteEntries = Object.values(allNotes);

  if (noteEntries.length === 0) return null;

  const sortedNotes = [...noteEntries]
    .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
    .slice(0, 3);

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-[#D0D0D0]">最近のメモ</p>
        <span className="text-[10px] text-[#666666]">{noteEntries.length}件</span>
      </div>

      <div className="space-y-2">
        {sortedNotes.map((entry) => (
          <div
            key={entry.sessionId}
            className="bg-surface-2 rounded p-2"
          >
            <p className="text-xs text-[#D0D0D0] line-clamp-2">{entry.note}</p>
            <p className="text-[10px] text-[#666666] mt-1">
              {new Date(entry.updatedAt).toLocaleDateString('ja-JP')}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}
