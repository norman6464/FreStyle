import { useState, useCallback } from 'react';
import type { Note } from '../types';
import NoteRepository from '../repositories/NoteRepository';

export function useNotes() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [selectedNoteId, setSelectedNoteId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchNotes = useCallback(async () => {
    setLoading(true);
    try {
      const data = await NoteRepository.fetchNotes();
      setNotes(data);
    } catch {
      // エラーハンドリング
    } finally {
      setLoading(false);
    }
  }, []);

  const createNote = useCallback(async (title: string) => {
    try {
      const note = await NoteRepository.createNote(title);
      await fetchNotes();
      setSelectedNoteId(note.noteId);
      return note;
    } catch {
      return null;
    }
  }, [fetchNotes]);

  const updateNote = useCallback(async (noteId: string, data: { title: string; content: string; isPinned: boolean }) => {
    try {
      await NoteRepository.updateNote(noteId, data);
      setNotes((prev) =>
        prev.map((n) => (n.noteId === noteId ? { ...n, ...data, updatedAt: Date.now() } : n))
      );
    } catch {
      // エラーハンドリング
    }
  }, []);

  const deleteNote = useCallback(async (noteId: string) => {
    try {
      await NoteRepository.deleteNote(noteId);
      await fetchNotes();
      if (selectedNoteId === noteId) {
        setSelectedNoteId(null);
      }
    } catch {
      // エラーハンドリング
    }
  }, [fetchNotes, selectedNoteId]);

  const selectNote = useCallback((noteId: string | null) => {
    setSelectedNoteId(noteId);
  }, []);

  return {
    notes,
    selectedNoteId,
    loading,
    fetchNotes,
    createNote,
    updateNote,
    deleteNote,
    selectNote,
  };
}
