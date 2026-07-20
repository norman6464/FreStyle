import { useCallback, useEffect, useRef, useState } from 'react';
import type { TeachingMaterial } from '@/entities/course';
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

  // 既にエディタへロード済みの material id を覚えておくための ref。
  // autosave 後の materials 再 fetch で `selected` ref が変わっても、
  // ref が同じ id を指している間は editor state を上書きしない。
  //
  // 上書きすると、 ユーザが入力中だった差分が autosave 完了タイミングで
  // 巻き戻り、 さらに textarea の undo 履歴が壊れて cmd+z が効かなくなる。
  const loadedIdRef = useRef<number | null>(null);

  useEffect(() => {
    if (selectedId == null) {
      loadedIdRef.current = null;
      setEditTitle('');
      setEditContent('');
      setEditIsPublished(false);
      setSaveStatus('idle');
      return;
    }
    if (loadedIdRef.current === selectedId) return;
    if (selected && selected.id === selectedId) {
      loadedIdRef.current = selectedId;
      setEditTitle(selected.title);
      setEditContent(selected.content);
      setEditIsPublished(selected.isPublished);
      setSaveStatus('idle');
    }
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
