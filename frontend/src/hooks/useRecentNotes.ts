import { useEffect, useState } from 'react';
import { SessionNoteRepository } from '../repositories/SessionNoteRepository';
import type { SessionNote } from '../types';

export function useRecentNotes(limit = 3) {
  const [notes, setNotes] = useState<SessionNote[]>([]);
  const [totalCount, setTotalCount] = useState(0);

  useEffect(() => {
    const allNotes = SessionNoteRepository.getAll();
    const entries = Object.values(allNotes) as SessionNote[];

    const sortedNotes = [...entries].sort(
      (a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime(),
    );

    setNotes(sortedNotes.slice(0, limit));
    setTotalCount(entries.length);
  }, [limit]);

  return { notes, totalCount };
}
