import { useCallback, useEffect, useRef, useState } from 'react';
import type { TeachingMaterial } from '../types';
import type { SaveStatus } from './useNoteEditor';

interface Args {
  selectedId: number | null;
  selected: TeachingMaterial | null;
  update: (
    id: number,
    payload: { title: string; content: string; orderInCourse: number; isPublished: boolean },
  ) => Promise<void>;
}

/**
 * useTeachingMaterialEditor — 教材詳細の編集 + autosave 制御。
 * 構造は useNoteEditor と同等で、 isPublished の切替も含む。
 */
export function useTeachingMaterialEditor({ selectedId, selected, update }: Args) {
  const [editTitle, setEditTitle] = useState('');
  const [editContent, setEditContent] = useState('');
  const [editIsPublished, setEditIsPublished] = useState(false);
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle');
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (selected) {
      setEditTitle(selected.title);
      setEditContent(selected.content);
      setEditIsPublished(selected.isPublished);
    } else {
      setEditTitle('');
      setEditContent('');
      setEditIsPublished(false);
    }
    setSaveStatus('idle');
  }, [selectedId, selected]);

  useEffect(() => {
    return () => {
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    };
  }, []);

  const scheduleSave = useCallback(
    (title: string, content: string, isPublished: boolean) => {
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
      setSaveStatus('unsaved');
      saveTimerRef.current = setTimeout(async () => {
        if (selectedId == null || !selected) return;
        setSaveStatus('saving');
        try {
          await update(selectedId, {
            title,
            content,
            orderInCourse: selected.orderInCourse,
            isPublished,
          });
          setSaveStatus('saved');
        } catch {
          setSaveStatus('idle');
        }
      }, 800);
    },
    [selectedId, selected, update],
  );

  const handleTitleChange = useCallback(
    (title: string) => {
      setEditTitle(title);
      scheduleSave(title, editContent, editIsPublished);
    },
    [scheduleSave, editContent, editIsPublished],
  );

  const handleContentChange = useCallback(
    (content: string) => {
      setEditContent(content);
      scheduleSave(editTitle, content, editIsPublished);
    },
    [scheduleSave, editTitle, editIsPublished],
  );

  const handleIsPublishedChange = useCallback(
    (isPublished: boolean) => {
      setEditIsPublished(isPublished);
      scheduleSave(editTitle, editContent, isPublished);
    },
    [scheduleSave, editTitle, editContent],
  );

  return {
    editTitle,
    editContent,
    editIsPublished,
    saveStatus,
    handleTitleChange,
    handleContentChange,
    handleIsPublishedChange,
  };
}
