import { useCallback, useEffect, useRef, useState } from 'react';
import {
  KbRepository,
  emitKbTreeEvent,
  rememberVisitedPage,
  forgetVisitedPageIfMatches,
  type KbIcon,
  type KbResolvedPage,
} from '@/entities/kb';
import { getApiError } from '@/shared/lib/classifyApiError';
import type { SaveStatus } from '@/shared/ui/RichTextEditor';

export interface KbPageDocState {
  data: KbResolvedPage | null;
  loading: boolean;
  /** 失敗の理由。null なら失敗していない。 */
  error: string | null;
}

/** 打鍵が止まってから保存を撃つまでの間合い。 */
const SAVE_DEBOUNCE_MS = 800;

/**
 * useKbPageDoc は /kb/{pageId} の URL からページを解決し、本文の保存も持つ。
 *
 * URL にはページ ID しか無いので、所属ワークスペース（以降の API に要る slug）と
 * 編集可否はサーバーの解決 API が一緒に返す。404 は「無い」と「見えない」の両方を
 * 意味する。backend が撃ち分けていないので（撃ち分けると ID の総当たりで実在が
 * 分かる）、**フロントで「見る権限がありません」と書いてはいけない。**
 */
export function useKbPageDoc(pageId: string | undefined) {
  const [state, setState] = useState<KbPageDocState>({ data: null, loading: false, error: null });
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle');
  // 本文保存が block_id_conflict（409）で失敗した回数。0 は「まだ起きていない」。
  // 呼び出し側（KbPage）はこの値が変わるたびに再読み込みを促す通知を出す
  // （boolean だと同じ真値が続くだけで2回目以降の発火を検知できないため回数にする）。
  const [contentConflictCount, setContentConflictCount] = useState(0);

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
    KbRepository.replaceContent(pending.workspaceSlug, pending.pageId, pending.doc)
      .then((res) => {
        saveInFlight.current = false;
        // 画面は現在ユーザーの名前を持っていないので、保存後の「最終編集」はこの応答で
        // 更新する。移った先で戻ってきた応答（pending.pageId が古い画面のページ）は
        // 反映しない — 反映すると、いま見ているページの最終編集が別ページのものになる。
        setState((prev) => {
          if (!prev.data || prev.data.page.id !== pending.pageId) return prev;
          return {
            ...prev,
            data: { ...prev.data, lastEditedBy: res.lastEditedBy, lastEditedAt: res.lastEditedAt },
          };
        });
        if (pendingSaves.current.size === 0) {
          setSaveStatus('saved');
        } else {
          // 送信中にさらに書かれていた。次を続けて送る（書いた順を守る）。
          setSaveStatus('unsaved');
          flushSave();
        }
      })
      .catch((err) => {
        saveInFlight.current = false;
        setSaveStatus('unsaved');
        // ブロック id の衝突（別ページの id を乗っ取ろうとした・他クライアントとの
        // 並行編集で起きるレース）は、この画面の状態を書き換えても再送で直らない
        // （エディタ側が古いページの block id を持ったままの可能性がある）。
        // 呼び出し側で再読み込みを促す通知を出せるよう、原因を区別して伝える。
        if (getApiError(err).serverCode === 'block_id_conflict') {
          setContentConflictCount((n) => n + 1);
        }
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

    KbRepository.resolvePage(pageId)
      .then((data) => {
        if (token !== generation.current) return;
        saveTarget.current = { workspaceSlug: data.workspaceSlug, pageId: data.page.id };
        setState({ data, loading: false, error: null });
        // ヘッダーの「ナレッジ」ボタンや素の /kb が「前回の続き」へ戻れるよう覚えておく
        // (entities/kb/lib/lastVisitedPage.ts)。開けた時点で覚える — 編集は必ず
        // 開いた後に起きるので、これで「閲覧・編集した」の両方をカバーできる。
        rememberVisitedPage(pageId);
      })
      .catch(() => {
        if (token !== generation.current) return;
        setState({
          data: null,
          loading: false,
          error: 'このページを開けませんでした。移動または削除された可能性があります。',
        });
        // 覚えていたのがこのページ(削除・移動済み)なら忘れる。次回はここへ戻らず、
        // 最初に見つかるページへ自動で移れるようにする。
        forgetVisitedPageIfMatches(pageId);
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
    const page = await KbRepository.renamePage(target.workspaceSlug, target.pageId, title);
    // 応答が返る前に別ページへ移っていたら、画面の状態には触らない
    //（触ると、移った先の見出しと ID が前のページのもので上書きされる）。
    // 改名そのものはサーバーで成立しているので、木への知らせは出す。
    if (token === generation.current) {
      setState((prev) => (prev.data ? { ...prev, data: { ...prev.data, page } } : prev));
    }
    emitKbTreeEvent({ type: 'page-updated', page });
  }, []);

  /**
   * changeIcon はページのアイコンを設定・解除する（`icon` が null なら解除）。
   * **失敗は投げる**（renameTitle と同じ理由 — 呼び出し側がトーストで知らせる）。
   * 成功したら画面の状態を確定後の値で差し替え、サイドバーの木にも知らせる。
   */
  const changeIcon = useCallback(async (icon: KbIcon | null): Promise<void> => {
    const target = saveTarget.current;
    if (!target) return;
    const token = generation.current;
    const page = icon
      ? await KbRepository.setPageIcon(target.workspaceSlug, target.pageId, icon)
      : await KbRepository.clearPageIcon(target.workspaceSlug, target.pageId);
    // 応答が返る前に別ページへ移っていたら、画面の状態には触らない（renameTitle と同じ守り）。
    if (token === generation.current) {
      setState((prev) => (prev.data ? { ...prev, data: { ...prev.data, page } } : prev));
    }
    emitKbTreeEvent({ type: 'page-updated', page });
  }, []);

  /**
   * changeCover はページのカバー画像を設定・解除する（`key` が null なら解除）。
   *
   * **アップロード自体はここでは行わない。** 渡す key は呼び出し側
   * （KbPageCoverButton）が KbRepository.uploadPageImage で S3 へ上げ終えた後のもの
   * — バリデーション・アップロード・設定の一連の流れは 1 箇所（呼び出し側）にまとめる。
   *
   * **失敗は投げる**（changeIcon と同じ理由 — 呼び出し側がトーストで知らせる）。
   * 成功したら画面の状態を確定後の値（page・cover）で差し替え、サイドバーの木にも知らせる。
   */
  const changeCover = useCallback(async (key: string | null): Promise<void> => {
    const target = saveTarget.current;
    if (!target) return;
    const token = generation.current;
    const { page, cover } = key
      ? await KbRepository.setPageCover(target.workspaceSlug, target.pageId, key)
      : await KbRepository.clearPageCover(target.workspaceSlug, target.pageId);
    // 応答が返る前に別ページへ移っていたら、画面の状態には触らない（changeIcon と同じ守り）。
    if (token === generation.current) {
      setState((prev) => (prev.data ? { ...prev, data: { ...prev.data, page, cover } } : prev));
    }
    emitKbTreeEvent({ type: 'page-updated', page });
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

  return {
    ...state,
    saveStatus,
    contentConflictCount,
    onDocChange,
    renameTitle,
    changeIcon,
    changeCover,
  };
}
