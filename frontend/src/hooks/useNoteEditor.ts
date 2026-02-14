import { useState, useEffect, useCallback, useRef } from 'react';
import type { Note } from '../types';

export function useNoteEditor(
  selectedNoteId: string | null,
  selectedNote: Note | null,
  updateNote: (noteId: string, data: { title: string; content: string; isPinned: boolean }) => Promise<void>
) {
  const [editTitle, setEditTitle] = useState('');
  const [editContent, setEditContent] = useState('');
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (selectedNote) {
      setEditTitle(selectedNote.title);
      setEditContent(selectedNote.content);
    } else {
      setEditTitle('');
      setEditContent('');
    }
  }, [selectedNoteId]);

  const handleAutoSave = useCallback(
    (title: string, content: string) => {
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
      saveTimerRef.current = setTimeout(() => {
        if (selectedNoteId) {
          updateNote(selectedNoteId, {
            title,
            content,
            isPinned: selectedNote?.isPinned || false,
          });
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

  return {
    editTitle,
    editContent,
    handleTitleChange,
    handleContentChange,
  };
}
