import type { SessionNote } from '../types';

const STORAGE_KEY = 'freestyle_session_notes';

function getAllNotes(): Record<string, SessionNote> {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return {};
  return JSON.parse(raw);
}

export const SessionNoteRepository = {
  get(sessionId: number): SessionNote | null {
    const notes = getAllNotes();
    return notes[String(sessionId)] || null;
  },

  save(sessionId: number, note: string): void {
    const notes = getAllNotes();
    notes[String(sessionId)] = {
      sessionId,
      note,
      updatedAt: new Date().toISOString(),
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(notes));
  },
};
