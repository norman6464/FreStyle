import { useMemo } from 'react';
import { SessionNoteRepository } from '../repositories/SessionNoteRepository';
import type { SessionNote } from '../types';

export function useRecentNotes(limit = 3) {
  const notes = useMemo(() => {
    const allNotes = SessionNoteRepository.getAll();
    const entries = Object.values(allNotes) as SessionNote[];
    return [...entries]
      .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
      .slice(0, limit);
  }, [limit]);

  const totalCount = useMemo(() => {
    return Object.keys(SessionNoteRepository.getAll()).length;
  }, []);

  return { notes, totalCount };
}
