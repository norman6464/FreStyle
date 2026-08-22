import { useState, useEffect, useCallback, useRef } from 'react';
import axios from 'axios';
import {
  DocumentRepository,
  toRichDocumentSummary,
  type RichDocument,
} from '@/entities/document';
import { emptyRichDoc, type RichDocContent, type SaveStatus } from '@/shared/ui/RichTextEditor';

const AUTOSAVE_DELAY_MS = 800;

interface UseDocumentEditorOptions {
  /** 保存成功時に一覧サマリ（title / updatedAt / revision）を同期する。 */
  onSynced?: (summary: ReturnType<typeof toRichDocumentSummary>) => void;
  /** 版不一致（409）を検出し、サーバの最新版で読み直したときに通知する。 */
  onConflict?: () => void;
}

/**
 * useDocumentEditor — 選択中のリッチ文書の本文編集と保存を担う hook。
 *
 * 選択が変わるたびに doc 本体を個別取得（一覧はサマリのみ）。編集は debounce 自動保存で、
 * 保存は revision による楽観ロック（PUT）。設計上の要点:
 *   - 別文書へ切り替える直前に、保留中の保存を「前の文書」に対して実行する（編集を失わない）。
 *   - 保存の多重実行を防ぐ（in-flight 中は保留し、完了後に 1 回だけ再実行）。自分自身の保存で
 *     409 を誘発しないため。
 *   - 版不一致（409）はサーバ最新版を取り直してエディタへ反映＋onConflict 通知。
 *   - 保存失敗（409 以外）は saveStatus を 'unsaved' に戻す（無表示にせず再試行を促す）。
 *   - doc 取得失敗は loadError で示し、空エディタを編集可能にしない。
 */
export function useDocumentEditor(selectedId: string | null, options: UseDocumentEditorOptions = {}) {
  const { onSynced, onConflict } = options;
  const [editTitle, setEditTitle] = useState('');
  const [editDoc, setEditDoc] = useState<RichDocContent>(emptyRichDoc);
  const [revision, setRevision] = useState(0);
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle');
  const [loadingDoc, setLoadingDoc] = useState(false);
  const [loadError, setLoadError] = useState(false);

  // 保存で参照する最新値を ref に持つ。React 19 の並行レンダーでコミットされない値を残さないよう、
  // ref の同期はレンダー中ではなく useEffect で行う。
  const titleRef = useRef(editTitle);
  const docRef = useRef(editDoc);
  const revisionRef = useRef(revision);
  const selectedIdRef = useRef<string | null>(selectedId);
  const onSyncedRef = useRef(onSynced);
  const onConflictRef = useRef(onConflict);
  useEffect(() => {
    titleRef.current = editTitle;
  }, [editTitle]);
  useEffect(() => {
    docRef.current = editDoc;
  }, [editDoc]);
  useEffect(() => {
    revisionRef.current = revision;
  }, [revision]);
  useEffect(() => {
    onSyncedRef.current = onSynced;
  }, [onSynced]);
  useEffect(() => {
    onConflictRef.current = onConflict;
  }, [onConflict]);

  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const savingRef = useRef(false);
  const pendingRef = useRef(false);

  const clearTimer = () => {
    if (saveTimerRef.current) {
      clearTimeout(saveTimerRef.current);
      saveTimerRef.current = null;
    }
  };

  // targetId の文書を現在のスナップショット（title/doc/revision ref）で保存する。
  // 完了時に選択が別文書へ変わっていたらエディタ state は触らない（PUT と一覧同期だけ行う）。
  const runSave = useCallback(async (targetId: string) => {
    if (savingRef.current) {
      // 実行中は多重に走らせない。完了後に 1 回だけ再実行する。
      pendingRef.current = true;
      return;
    }
    savingRef.current = true;
    const isCurrent = () => selectedIdRef.current === targetId;
    if (isCurrent()) setSaveStatus('saving');
    try {
      const updated = await DocumentRepository.updateDocument(targetId, {
        title: titleRef.current,
        doc: docRef.current,
        revision: revisionRef.current,
      });
      revisionRef.current = updated.revision;
      if (isCurrent()) {
        setRevision(updated.revision);
        setSaveStatus('saved');
      }
      onSyncedRef.current?.(toRichDocumentSummary(updated));
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        try {
          const fresh = await DocumentRepository.fetchDocument(targetId);
          revisionRef.current = fresh.revision;
          onSyncedRef.current?.(toRichDocumentSummary(fresh));
          if (isCurrent()) {
            applyServerDoc(fresh);
            setSaveStatus('saved');
            onConflictRef.current?.();
          }
          return;
        } catch {
          // 取り直しも失敗したら未保存へ戻す（下の共通処理へ流す）。
        }
      }
      // 409 以外の失敗・取り直し失敗は未保存に戻し、再試行を促す（無表示にしない）。
      if (isCurrent()) setSaveStatus('unsaved');
    } finally {
      savingRef.current = false;
      if (pendingRef.current) {
        pendingRef.current = false;
        const next = selectedIdRef.current;
        if (next != null) void runSave(next);
      }
    }
    // applyServerDoc は下で定義した安定参照。runSave の依存には含めない。
  }, []);

  // サーバの文書でエディタ state と ref を同期する（load / 409 取り直しで共用）。
  function applyServerDoc(document: RichDocument) {
    setEditTitle(document.title);
    setEditDoc(document.doc);
    setRevision(document.revision);
    titleRef.current = document.title;
    docRef.current = document.doc;
    revisionRef.current = document.revision;
  }

  // 選択が変わったら、保留中の保存を「前の文書」へ実行してから新しい doc を取得する。
  useEffect(() => {
    const prevId = selectedIdRef.current;
    if (saveTimerRef.current) {
      clearTimer();
      if (prevId != null) void runSave(prevId);
    }
    selectedIdRef.current = selectedId;

    let cancelled = false;
    if (selectedId == null) {
      setEditTitle('');
      setEditDoc(emptyRichDoc());
      setRevision(0);
      setSaveStatus('idle');
      setLoadingDoc(false);
      setLoadError(false);
      return;
    }
    setLoadingDoc(true);
    setLoadError(false);
    setSaveStatus('idle');
    DocumentRepository.fetchDocument(selectedId)
      .then((document) => {
        if (cancelled) return;
        applyServerDoc(document);
      })
      .catch(() => {
        if (cancelled) return;
        // 取得失敗を明示する（空エディタを編集可能にしない）。
        setLoadError(true);
      })
      .finally(() => {
        if (!cancelled) setLoadingDoc(false);
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedId]);

  // アンマウント時にタイマーを解除する。
  useEffect(() => () => clearTimer(), []);

  const scheduleSave = useCallback(() => {
    clearTimer();
    setSaveStatus('unsaved');
    saveTimerRef.current = setTimeout(() => {
      const id = selectedIdRef.current;
      if (id != null) void runSave(id);
    }, AUTOSAVE_DELAY_MS);
  }, [runSave]);

  const handleTitleChange = useCallback(
    (title: string) => {
      setEditTitle(title);
      titleRef.current = title;
      scheduleSave();
    },
    [scheduleSave],
  );

  const handleDocChange = useCallback(
    (doc: RichDocContent) => {
      setEditDoc(doc);
      docRef.current = doc;
      scheduleSave();
    },
    [scheduleSave],
  );

  const forceSave = useCallback(() => {
    const id = selectedIdRef.current;
    if (id == null) return;
    clearTimer();
    void runSave(id);
  }, [runSave]);

  // reload は loadError からの再試行。選択中の文書をもう一度取得する。
  const reload = useCallback(() => {
    const id = selectedIdRef.current;
    if (id == null) return;
    setLoadingDoc(true);
    setLoadError(false);
    DocumentRepository.fetchDocument(id)
      .then((document) => applyServerDoc(document))
      .catch(() => setLoadError(true))
      .finally(() => setLoadingDoc(false));
  }, []);

  return {
    editTitle,
    editDoc,
    saveStatus,
    loadingDoc,
    loadError,
    handleTitleChange,
    handleDocChange,
    forceSave,
    reload,
  };
}
