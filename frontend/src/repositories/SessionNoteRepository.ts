import apiClient from '../lib/axios';
import type { SessionNote } from '../types';

const STORAGE_KEY = 'freestyle_session_notes';

function getAllLocalNotes(): Record<string, SessionNote> {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return {};
  try {
    return JSON.parse(raw);
  } catch {
    localStorage.removeItem(STORAGE_KEY);
    return {};
  }
}

function saveLocalNote(sessionId: number, note: string): void {
  const notes = getAllLocalNotes();
  notes[String(sessionId)] = {
    sessionId,
    note,
    updatedAt: new Date().toISOString(),
  };
  localStorage.setItem(STORAGE_KEY, JSON.stringify(notes));
}

export const SessionNoteRepository = {
  getAll(): Record<string, SessionNote> {
    return getAllLocalNotes();
  },

  async get(sessionId: number): Promise<SessionNote | null> {
    try {
      const response = await apiClient.get<SessionNote>(`/api/session-notes/${sessionId}`);
      return response.data;
    } catch {
      const notes = getAllLocalNotes();
      return notes[String(sessionId)] || null;
    }
  },

  async save(sessionId: number, note: string): Promise<void> {
    try {
      await apiClient.put(`/api/session-notes/${sessionId}`, { note });
    } catch {
      saveLocalNote(sessionId, note);
    }
  },
};
