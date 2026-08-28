import { useCallback, useEffect, useRef, useState } from 'react';
import { NoteRepository, type NoteResolvedPage } from '@/entities/note';
import type { SaveStatus } from '@/shared/ui/RichTextEditor';

export interface NotePageDocState {
  data: NoteResolvedPage | null;
  loading: boolean;
  /** 失敗の理由。null なら失敗していない。 */
  error: string | null;
}

/** 打鍵が止まってから保存を撃つまでの間合い。 */
const SAVE_DEBOUNCE_MS = 800;

/**
 * useNotePageDoc は /p/{pageId} の URL からページを解決し、本文の保存も持つ。
 *
 * URL にはページ ID しか無いので、所属ワークスペース（以降の API に要る slug）と
 * 編集可否はサーバーの解決 API が一緒に返す。404 は「無い」と「見えない」の両方を
 * 意味する。backend が撃ち分けていないので（撃ち分けると ID の総当たりで実在が
 * 分かる）、**フロントで「見る権限がありません」と書いてはいけない。**
 */
export function useNotePageDoc(pageId: string | undefined) {
  const [state, setState] = useState<NotePageDocState>({ data: null, loading: false, error: null });
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle');

  // 速く行き来したときに、古い応答が新しいページを上書きするのを防ぐ。
  const generation = useRef(0);

  // 保存のデバウンスと「最後に書かれた doc」。タイマーは 1 本だけ持ち、
  // 発火時点の最新 doc を送る（打鍵ごとに PUT しない）。
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pendingDoc = useRef<unknown>(null);
  const saveTarget = useRef<{ workspaceSlug: string; pageId: string } | null>(null);
  // PUT が飛んでいる間 true。保存は**必ず 1 本ずつ**送る。並行に送ると、後から書いた
  // 本文の PUT が先に完了し、古い本文の PUT が後から着地して上書きすることがある
  //（丸ごと置換の API なので、順序が崩れる＝最後の入力が消える）。
  const saveInFlight = useRef(false);

  const flushSave = useCallback(() => {
    if (saveInFlight.current) return; // 完了ハンドラが残りを流す
    const target = saveTarget.current;
    const doc = pendingDoc.current;
    if (!target || doc == null) return;
    pendingDoc.current = null;
    saveInFlight.current = true;
    setSaveStatus('saving');
    NoteRepository.replaceContent(target.workspaceSlug, target.pageId, doc)
      .then(() => {
        saveInFlight.current = false;
        if (pendingDoc.current == null) {
          setSaveStatus('saved');
        } else {
          // 送信中にさらに書かれていた。次を続けて送る（編集順を守る）。
          setSaveStatus('unsaved');
          flushSave();
        }
      })
      .catch(() => {
        saveInFlight.current = false;
        setSaveStatus('unsaved');
      });
  }, []);


  useEffect(() => {
    if (!pageId) {
      setState({ data: null, loading: false, error: null });
      setSaveStatus('idle');
      return;
    }
    const token = ++generation.current;
    setState((prev) => ({ ...prev, loading: true, error: null }));
    setSaveStatus('idle');

    NoteRepository.resolvePage(pageId)
      .then((data) => {
        if (token !== generation.current) return;
        saveTarget.current = { workspaceSlug: data.workspaceSlug, pageId: data.page.id };
        setState({ data, loading: false, error: null });
      })
      .catch(() => {
        if (token !== generation.current) return;
        setState({
          data: null,
          loading: false,
          error: 'このページを開けませんでした。移動または削除された可能性があります。',
        });
      });

    return () => {
      // ページを離れるとき、書きかけがあれば待たずに送る（デバウンス分の取りこぼし防止）。
      if (saveTimer.current) {
        clearTimeout(saveTimer.current);
        saveTimer.current = null;
        flushSave();
      }
    };
  }, [pageId, flushSave]);

  /** onDocChange はエディタの onChange から呼ぶ。デバウンスして本文を保存する。 */
  const onDocChange = useCallback(
    (doc: unknown) => {
      pendingDoc.current = doc;
      setSaveStatus('unsaved');
      if (saveTimer.current) clearTimeout(saveTimer.current);
      saveTimer.current = setTimeout(() => {
        saveTimer.current = null;
        flushSave();
      }, SAVE_DEBOUNCE_MS);
    },
    [flushSave],
  );

  return { ...state, saveStatus, onDocChange };
}
