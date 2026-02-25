import { useState, useEffect, useCallback, useRef } from 'react';
import type { Note } from '../types';

export type SaveStatus = 'idle' | 'unsaved' | 'saving' | 'saved';

export function useNoteEditor(
  selectedNoteId: string | null,
  selectedNote: Note | null,
  updateNote: (noteId: string, data: { title: string; content: string; isPinned: boolean }) => Promise<void>
) {
  const [editTitle, setEditTitle] = useState('');
  const [editContent, setEditContent] = useState('');
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle');
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (selectedNote) {
      setEditTitle(selectedNote.title);
      setEditContent(selectedNote.content);
    } else {
      setEditTitle('');
      setEditContent('');
    }
    setSaveStatus('idle');
  }, [selectedNoteId]);

  useEffect(() => {
    return () => {
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    };
  }, []);

  const handleAutoSave = useCallback(
    (title: string, content: string) => {
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
      setSaveStatus('unsaved');
      saveTimerRef.current = setTimeout(async () => {
        if (selectedNoteId) {
          setSaveStatus('saving');
          try {
            await updateNote(selectedNoteId, {
              title,
              content,
              isPinned: selectedNote?.isPinned || false,
            });
            setSaveStatus('saved');
          } catch {
            setSaveStatus('idle');
          }
        }
      }, 800);
    },
    [selectedNoteId, selectedNote, updateNote]
  );

  const handleTitleChange = useCallback(
    (title: string) => {
      setEditTitle(title);
      handleAutoSave(title, editContent);
    },
    [handleAutoSave, editContent]
  );

  const handleContentChange = useCallback(
    (content: string) => {
      setEditContent(content);
      handleAutoSave(editTitle, content);
    },
    [handleAutoSave, editTitle]
  );

  const forceSave = useCallback(async () => {
    if (!selectedNoteId) return;
    if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    setSaveStatus('saving');
    try {
      await updateNote(selectedNoteId, {
        title: editTitle,
        content: editContent,
        isPinned: selectedNote?.isPinned || false,
      });
      setSaveStatus('saved');
    } catch {
      setSaveStatus('idle');
    }
  }, [selectedNoteId, editTitle, editContent, selectedNote, updateNote]);

  return {
    editTitle,
    editContent,
    saveStatus,
    handleTitleChange,
    handleContentChange,
    forceSave,
  };
}
