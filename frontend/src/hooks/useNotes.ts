import { useState, useCallback, useMemo, useRef } from 'react';
import type { Note } from '../types';
import NoteRepository from '../repositories/NoteRepository';

export function useNotes() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [selectedNoteId, setSelectedNoteId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const notesRef = useRef<Note[]>(notes);
  notesRef.current = notes;

  const fetchNotes = useCallback(async () => {
    setLoading(true);
    try {
      const data = await NoteRepository.fetchNotes();
      setNotes(Array.isArray(data) ? data : []);
    } catch {
      // エラーハンドリング
    } finally {
      setLoading(false);
    }
  }, []);

  const createNote = useCallback(async (title: string) => {
    try {
      const note = await NoteRepository.createNote(title);
      setNotes((prev) => [note, ...prev]);
      setSelectedNoteId(note.noteId);
      return note;
    } catch {
      return null;
    }
  }, []);

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
      setNotes((prev) => prev.filter((n) => n.noteId !== noteId));
      setSelectedNoteId((prev) => (prev === noteId ? null : prev));
    } catch {
      // エラーハンドリング
    }
  }, []);

  const togglePin = useCallback(async (noteId: string) => {
    const note = notesRef.current.find((n) => n.noteId === noteId);
    if (!note) return;
    const newPinned = !note.isPinned;
    try {
      await NoteRepository.updateNote(noteId, {
        title: note.title,
        content: note.content,
        isPinned: newPinned,
      });
      setNotes((prev) =>
        prev.map((n) => (n.noteId === noteId ? { ...n, isPinned: newPinned } : n))
      );
    } catch {
      // エラーハンドリング
    }
  }, []);

  const selectNote = useCallback((noteId: string | null) => {
    setSelectedNoteId(noteId);
  }, []);

  const filteredNotes = useMemo(() => {
    const query = searchQuery.toLowerCase();
    const filtered = query
      ? notes.filter((n) => n.title.toLowerCase().includes(query) || n.content.toLowerCase().includes(query))
      : notes;
    return [...filtered].sort((a, b) => {
      if (a.isPinned !== b.isPinned) return a.isPinned ? -1 : 1;
      return b.updatedAt - a.updatedAt;
    });
  }, [notes, searchQuery]);

  return {
    notes,
    filteredNotes,
    selectedNoteId,
    loading,
    searchQuery,
    setSearchQuery,
    fetchNotes,
    createNote,
    updateNote,
    deleteNote,
    selectNote,
    togglePin,
  };
}
