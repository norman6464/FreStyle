import { useState, useCallback, useMemo, useRef } from 'react';
import type { Note } from '../types';
import type { NoteSortOption } from '../constants/sortOptions';
import NoteRepository from '../repositories/NoteRepository';

export function useNotes() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [selectedNoteId, setSelectedNoteId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);
  const [noteSort, setNoteSort] = useState<NoteSortOption>('default');
  const notesRef = useRef<Note[]>(notes);
  notesRef.current = notes;

  const fetchNotes = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await NoteRepository.fetchNotes();
      setNotes(Array.isArray(data) ? data : []);
    } catch {
      setError('ノートの取得に失敗しました');
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
      setError('ノートの更新に失敗しました');
    }
  }, []);

  const deleteNote = useCallback(async (noteId: string) => {
    try {
      await NoteRepository.deleteNote(noteId);
      setNotes((prev) => prev.filter((n) => n.noteId !== noteId));
      setSelectedNoteId((prev) => (prev === noteId ? null : prev));
    } catch {
      setError('ノートの削除に失敗しました');
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
      setError('ピン留めの変更に失敗しました');
    }
  }, []);

  const selectNote = useCallback((noteId: string | null) => {
    setSelectedNoteId(noteId);
  }, []);

  const selectedNote = useMemo(() => {
    return notes.find((n) => n.noteId === selectedNoteId) || null;
  }, [notes, selectedNoteId]);

  const filteredNotes = useMemo(() => {
    const query = searchQuery.toLowerCase();
    const filtered = query
      ? notes.filter((n) => n.title.toLowerCase().includes(query) || n.content.toLowerCase().includes(query))
      : notes;
    return [...filtered].sort((a, b) => {
      if (a.isPinned !== b.isPinned) return a.isPinned ? -1 : 1;
      if (noteSort === 'updated-asc') return a.updatedAt - b.updatedAt;
      if (noteSort === 'title') return a.title.localeCompare(b.title, 'ja');
      if (noteSort === 'created-desc') return b.createdAt - a.createdAt;
      return b.updatedAt - a.updatedAt;
    });
  }, [notes, searchQuery, noteSort]);

  const requestDelete = useCallback((noteId: string) => {
    setDeleteTargetId(noteId);
  }, []);

  const confirmDelete = useCallback(async () => {
    if (!deleteTargetId) return;

    // 削除前に次の選択候補を決定（filteredNotesの順序で次→前）
    const idx = filteredNotes.findIndex((n) => n.noteId === deleteTargetId);
    const nextNote = idx >= 0
      ? filteredNotes[idx + 1] || filteredNotes[idx - 1] || null
      : null;

    await deleteNote(deleteTargetId);

    if (selectedNoteId === deleteTargetId && nextNote) {
      setSelectedNoteId(nextNote.noteId);
    }

    setDeleteTargetId(null);
  }, [deleteTargetId, deleteNote, filteredNotes, selectedNoteId]);

  const cancelDelete = useCallback(() => {
    setDeleteTargetId(null);
  }, []);

  return {
    notes,
    filteredNotes,
    selectedNoteId,
    selectedNote,
    loading,
    error,
    searchQuery,
    setSearchQuery,
    fetchNotes,
    createNote,
    updateNote,
    deleteNote,
    selectNote,
    togglePin,
    deleteTargetId,
    requestDelete,
    confirmDelete,
    cancelDelete,
    noteSort,
    setNoteSort,
  };
}
