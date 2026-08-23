import { useCallback, useEffect, useRef, useState } from 'react';
import axios from 'axios';
import { TeachingMaterialRepository, type TeachingMaterial } from '@/entities/course';
import type { SaveStatus } from '@/entities/note';
import { emptyRichDoc, type RichDocContent } from '@/shared/ui/RichTextEditor';

const AUTOSAVE_DELAY_MS = 800;

interface Args {
  selectedId: number | null;
  selected: TeachingMaterial | null;
  update: (
    id: number,
    payload: { title: string; content: string; orderInCourse: number; isPublished: boolean },
  ) => Promise<void>;
  /** doc 保存成功・409 取り直しで、詳細キャッシュへ最新の doc / revision を反映する。 */
  onDocSynced?: (material: TeachingMaterial) => void;
  /** 版不一致（409）でサーバ最新版へ読み直したときの通知（トースト等）。 */
  onConflict?: () => void;
}

/**
 * useTeachingMaterialEditor — 教材詳細の編集 + autosave 制御。
 *
 * 本文はリッチ本文（doc / tiptap JSON）を正本として編集する（docMode）。
 * doc が無く Markdown content だけを持つ章は、従来の textarea 編集に
 * フォールバックする（過渡期互換。フェーズ E で撤去）。新規章（content 空）は
 * 空 doc から docMode で始める。
 *
 * - doc の保存は revision の楽観ロック付き PUT。多重実行を防ぎ（in-flight 保留）、
 *   409 はサーバ最新版を取り直してエディタへ反映 + onConflict 通知
 *   （ノートの useDocumentEditor と同じ設計）。
 * - title / isPublished は従来 PUT（content 併送）で保存する（doc とは別タイマー）。
 */
export function useTeachingMaterialEditor({ selectedId, selected, update, onDocSynced, onConflict }: Args) {
  const [editTitle, setEditTitle] = useState('');
  const [editContent, setEditContent] = useState('');
  const [editDoc, setEditDoc] = useState<RichDocContent>(emptyRichDoc);
  const [editIsPublished, setEditIsPublished] = useState(false);
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle');
  const [docMode, setDocMode] = useState(false);
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const docTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // doc 保存で参照する最新値（React の並行レンダーに備え effect で同期）。
  const docRef = useRef(editDoc);
  const revisionRef = useRef(1);
  const selectedIdRef = useRef<number | null>(selectedId);
  const onDocSyncedRef = useRef(onDocSynced);
  const onConflictRef = useRef(onConflict);
  useEffect(() => {
    docRef.current = editDoc;
  }, [editDoc]);
  useEffect(() => {
    selectedIdRef.current = selectedId;
  }, [selectedId]);
  useEffect(() => {
    onDocSyncedRef.current = onDocSynced;
  }, [onDocSynced]);
  useEffect(() => {
    onConflictRef.current = onConflict;
  }, [onConflict]);

  // 既にエディタへロード済みの material id を覚えておくための ref。
  // autosave 後の materials 再 fetch で `selected` ref が変わっても、
  // ref が同じ id を指している間は editor state を上書きしない。
  //
  // 上書きすると、 ユーザが入力中だった差分が autosave 完了タイミングで
  // 巻き戻り、 さらに undo 履歴が壊れて cmd+z が効かなくなる。
  const loadedIdRef = useRef<number | null>(null);

  useEffect(() => {
    if (selectedId == null) {
      loadedIdRef.current = null;
      setEditTitle('');
      setEditContent('');
      setEditDoc(emptyRichDoc());
      setEditIsPublished(false);
      setSaveStatus('idle');
      setDocMode(false);
      return;
    }
    if (loadedIdRef.current === selectedId) return;
    if (selected && selected.id === selectedId) {
      loadedIdRef.current = selectedId;
      setEditTitle(selected.title);
      setEditContent(selected.content);
      setEditIsPublished(selected.isPublished);
      setSaveStatus('idle');
      // doc がある章、または新規章（content 空）はリッチ編集。doc 無し + content ありは
      // 従来の Markdown 編集へフォールバック（フェーズ C 未変換の章を壊さない）。
      const useDoc = selected.doc != null || selected.content.trim() === '';
      setDocMode(useDoc);
      const initialDoc = selected.doc ?? emptyRichDoc();
      setEditDoc(initialDoc);
      docRef.current = initialDoc;
      revisionRef.current = selected.revision ?? 1;
    }
  }, [selectedId, selected]);

  useEffect(() => {
    return () => {
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
      if (docTimerRef.current) clearTimeout(docTimerRef.current);
    };
  }, []);

  // ---- doc（リッチ本文）の保存。多重実行防止 + 409 の取り直しつき。 ----
  const savingRef = useRef(false);
  const pendingRef = useRef(false);

  const runDocSave = useCallback(async (targetId: number) => {
    if (savingRef.current) {
      pendingRef.current = true;
      return;
    }
    savingRef.current = true;
    const isCurrent = () => selectedIdRef.current === targetId;
    if (isCurrent()) setSaveStatus('saving');
    try {
      const updated = await TeachingMaterialRepository.updateDoc(targetId, {
        doc: docRef.current,
        expectedRevision: revisionRef.current,
      });
      revisionRef.current = updated.revision ?? revisionRef.current + 1;
      onDocSyncedRef.current?.(updated);
      if (isCurrent()) setSaveStatus('saved');
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        try {
          const fresh = await TeachingMaterialRepository.get(targetId);
          revisionRef.current = fresh.revision ?? 1;
          onDocSyncedRef.current?.(fresh);
          if (isCurrent()) {
            const freshDoc = fresh.doc ?? emptyRichDoc();
            setEditDoc(freshDoc);
            docRef.current = freshDoc;
            setEditTitle(fresh.title);
            setSaveStatus('saved');
            onConflictRef.current?.();
          }
          return;
        } catch {
          // 取り直しも失敗したら未保存へ戻す（下の共通処理へ流す）。
        }
      }
      // 409 以外の失敗・取り直し失敗は未保存に戻し、再試行を促す。
      if (isCurrent()) setSaveStatus('unsaved');
    } finally {
      savingRef.current = false;
      if (pendingRef.current) {
        pendingRef.current = false;
        const next = selectedIdRef.current;
        if (next != null) void runDocSave(next);
      }
    }
  }, []);

  const handleDocChange = useCallback(
    (doc: RichDocContent) => {
      setEditDoc(doc);
      docRef.current = doc;
      if (docTimerRef.current) clearTimeout(docTimerRef.current);
      setSaveStatus('unsaved');
      docTimerRef.current = setTimeout(() => {
        const id = selectedIdRef.current;
        if (id != null) void runDocSave(id);
      }, AUTOSAVE_DELAY_MS);
    },
    [runDocSave],
  );

  // ---- title / content(旧モード) / isPublished の保存（従来 PUT）。 ----
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
      }, AUTOSAVE_DELAY_MS);
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
    editDoc,
    editIsPublished,
    saveStatus,
    docMode,
    handleTitleChange,
    handleContentChange,
    handleDocChange,
    handleIsPublishedChange,
  };
}
