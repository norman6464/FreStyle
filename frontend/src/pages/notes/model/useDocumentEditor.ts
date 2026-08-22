import { useState, useEffect, useCallback, useRef } from 'react';
import axios from 'axios';
import { DocumentRepository, type RichDocument, type RichDocumentSummary } from '@/entities/document';
import { emptyRichDoc, type RichDocContent, type SaveStatus } from '@/shared/ui/RichTextEditor';

const AUTOSAVE_DELAY_MS = 800;

interface UseDocumentEditorOptions {
  /** 保存成功時に一覧サマリ（title / updatedAt / revision）を同期する。 */
  onSynced?: (summary: RichDocumentSummary) => void;
  /** 版不一致（409）を検出し、サーバの最新版で読み直したときに通知する。 */
  onConflict?: () => void;
}

/**
 * useDocumentEditor — 選択中のリッチ文書の本文編集と保存を担う hook。
 *
 * 選択が変わるたびに doc 本体を個別取得（一覧はサマリのみ）。編集は debounce 自動保存で、
 * 保存は revision による楽観ロック（PUT）。版不一致（409）はサーバの最新版を取り直して
 * エディタへ反映し、onConflict で利用者に知らせる（GitHub / Jira 方式）。
 */
export function useDocumentEditor(selectedId: string | null, options: UseDocumentEditorOptions = {}) {
  const { onSynced, onConflict } = options;
  const [editTitle, setEditTitle] = useState('');
  const [editDoc, setEditDoc] = useState<RichDocContent>(emptyRichDoc);
  const [revision, setRevision] = useState(0);
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle');
  const [loadingDoc, setLoadingDoc] = useState(false);

  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  // 保存で参照する最新値を ref に持つ（debounce のクロージャ陳腐化を防ぐ）。
  const titleRef = useRef(editTitle);
  const docRef = useRef(editDoc);
  const revisionRef = useRef(revision);
  titleRef.current = editTitle;
  docRef.current = editDoc;
  revisionRef.current = revision;
  const onSyncedRef = useRef(onSynced);
  const onConflictRef = useRef(onConflict);
  onSyncedRef.current = onSynced;
  onConflictRef.current = onConflict;

  const clearTimer = () => {
    if (saveTimerRef.current) {
      clearTimeout(saveTimerRef.current);
      saveTimerRef.current = null;
    }
  };

  // 選択が変わったら doc 本体を取得してエディタを初期化する。
  useEffect(() => {
    clearTimer();
    let cancelled = false;
    if (selectedId == null) {
      setEditTitle('');
      setEditDoc(emptyRichDoc());
      setRevision(0);
      setSaveStatus('idle');
      setLoadingDoc(false);
      return;
    }
    setLoadingDoc(true);
    setSaveStatus('idle');
    DocumentRepository.fetchDocument(selectedId)
      .then((doc) => {
        if (cancelled) return;
        setEditTitle(doc.title);
        setEditDoc(doc.doc);
        setRevision(doc.revision);
      })
      .catch(() => {
        // 取得失敗時は空のまま（利用者は再選択で再取得できる）。
      })
      .finally(() => {
        if (!cancelled) setLoadingDoc(false);
      });
    return () => {
      cancelled = true;
    };
  }, [selectedId]);

  // アンマウント時にタイマーを解除する。
  useEffect(() => () => clearTimer(), []);

  const applyServerDoc = useCallback((doc: RichDocument) => {
    setEditTitle(doc.title);
    setEditDoc(doc.doc);
    setRevision(doc.revision);
    const { doc: _doc, ...summary } = doc;
    void _doc;
    onSyncedRef.current?.(summary);
  }, []);

  const save = useCallback(async () => {
    if (selectedId == null) return;
    setSaveStatus('saving');
    try {
      const updated = await DocumentRepository.updateDocument(selectedId, {
        title: titleRef.current,
        doc: docRef.current,
        revision: revisionRef.current,
      });
      setRevision(updated.revision);
      setSaveStatus('saved');
      const { doc: _doc, ...summary } = updated;
      void _doc;
      onSyncedRef.current?.(summary);
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        // 版不一致: サーバの最新版で読み直し、利用者に知らせる（編集内容は最新版で上書き）。
        try {
          const fresh = await DocumentRepository.fetchDocument(selectedId);
          applyServerDoc(fresh);
          setSaveStatus('saved');
          onConflictRef.current?.();
          return;
        } catch {
          // 取り直しも失敗したら idle に戻す。
        }
      }
      setSaveStatus('idle');
    }
  }, [selectedId, applyServerDoc]);

  const scheduleSave = useCallback(() => {
    clearTimer();
    setSaveStatus('unsaved');
    saveTimerRef.current = setTimeout(() => {
      void save();
    }, AUTOSAVE_DELAY_MS);
  }, [save]);

  const handleTitleChange = useCallback(
    (title: string) => {
      setEditTitle(title);
      scheduleSave();
    },
    [scheduleSave],
  );

  const handleDocChange = useCallback(
    (doc: RichDocContent) => {
      setEditDoc(doc);
      scheduleSave();
    },
    [scheduleSave],
  );

  const forceSave = useCallback(() => {
    if (selectedId == null) return;
    clearTimer();
    void save();
  }, [selectedId, save]);

  return {
    editTitle,
    editDoc,
    saveStatus,
    loadingDoc,
    handleTitleChange,
    handleDocChange,
    forceSave,
  };
}
