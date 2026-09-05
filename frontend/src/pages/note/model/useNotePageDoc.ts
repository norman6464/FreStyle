import { useCallback, useEffect, useRef, useState } from 'react';
import { NoteRepository, emitNoteTreeEvent, type NoteResolvedPage } from '@/entities/note';
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
 * useNotePageDoc は /kb/{pageId} の URL からページを解決し、本文の保存も持つ。
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
  //
  // **宛先（どのページの本文か）は doc と一緒に束ねて持つ。** 別々の ref に置くと、
  // ページを移った瞬間に宛先だけが新しいページへ差し替わり、旧ページの書きかけが
  // 新しいページへ PUT される（丸ごと置換の API なので、移った先の本文が旧ページの
  // 全文で上書きされる）。書いた時点のページが宛先 — この束がそれを崩れなくする。
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  // 保留は宛先（ページ）ごとに最新の doc を 1 つずつ持つ（Map は挿入順を保つ）。
  // 1 枠だけだと、旧ページの PUT が飛んでいる間に旧ページを書き直し → 移動 → 新ページを
  // 書く、の並びで旧ページの最後の編集が新ページの doc に上書きされて消える。
  // ページ単位の丸ごと置換なので、ページごとに最後の doc が届けば十分。
  const pendingSaves = useRef(
    new Map<string, { workspaceSlug: string; pageId: string; doc: unknown }>(),
  );
  const saveTarget = useRef<{ workspaceSlug: string; pageId: string } | null>(null);
  // PUT が飛んでいる間 true。保存は**必ず 1 本ずつ**送る。並行に送ると、後から書いた
  // 本文の PUT が先に完了し、古い本文の PUT が後から着地して上書きすることがある
  //（丸ごと置換の API なので、順序が崩れる＝最後の入力が消える）。
  const saveInFlight = useRef(false);

  const flushSave = useCallback(() => {
    if (saveInFlight.current) return; // 完了ハンドラが残りを流す
    const head = pendingSaves.current.entries().next();
    if (head.done) return;
    const [key, pending] = head.value;
    pendingSaves.current.delete(key);
    saveInFlight.current = true;
    setSaveStatus('saving');
    NoteRepository.replaceContent(pending.workspaceSlug, pending.pageId, pending.doc)
      .then(() => {
        saveInFlight.current = false;
        if (pendingSaves.current.size === 0) {
          setSaveStatus('saved');
        } else {
          // 送信中にさらに書かれていた。次を続けて送る（書いた順を守る）。
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

  /**
   * renameTitle は題名を変える。**失敗は投げる**（呼び出し側が入力を保って知らせる）。
   * 成功したら画面の状態を確定後の値で差し替え、サイドバーの木にも知らせる。
   */
  const renameTitle = useCallback(async (title: string): Promise<void> => {
    const target = saveTarget.current;
    if (!target) return;
    const token = generation.current;
    const page = await NoteRepository.renamePage(target.workspaceSlug, target.pageId, title);
    // 応答が返る前に別ページへ移っていたら、画面の状態には触らない
    //（触ると、移った先の見出しと ID が前のページのもので上書きされる）。
    // 改名そのものはサーバーで成立しているので、木への知らせは出す。
    if (token === generation.current) {
      setState((prev) => (prev.data ? { ...prev, data: { ...prev.data, page } } : prev));
    }
    emitNoteTreeEvent({ type: 'page-renamed', page });
  }, []);

  /** onDocChange はエディタの onChange から呼ぶ。デバウンスして本文を保存する。 */
  const onDocChange = useCallback(
    (doc: unknown) => {
      // 宛先は**書いたこの瞬間**のページ。あとで読むとページ移動で差し替わっている。
      const target = saveTarget.current;
      if (!target) return;
      pendingSaves.current.set(target.pageId, { ...target, doc });
      setSaveStatus('unsaved');
      if (saveTimer.current) clearTimeout(saveTimer.current);
      saveTimer.current = setTimeout(() => {
        saveTimer.current = null;
        flushSave();
      }, SAVE_DEBOUNCE_MS);
    },
    [flushSave],
  );

  return { ...state, saveStatus, onDocChange, renameTitle };
}
