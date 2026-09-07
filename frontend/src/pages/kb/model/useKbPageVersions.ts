import { useCallback, useEffect, useRef, useState } from 'react';
import {
  KbRepository,
  type KbPageContentSaveResult,
  type KbPageVersion,
  type KbPageVersionDetail,
} from '@/entities/kb';

export interface KbPageVersionsState {
  versions: KbPageVersion[];
  loading: boolean;
  /** 失敗の理由。null なら失敗していない。 */
  error: string | null;
  /** 「版を残す」（作成）が飛んでいる間 true。 */
  saving: boolean;
}

const EMPTY: KbPageVersionsState = {
  versions: [],
  loading: false,
  error: null,
  saving: false,
};

const LOAD_FAILED =
  '履歴を読み込めませんでした。通信が切れたか、このページを見る立場でなくなっています。開き直すと最新の状態が出ます。';

/** プレビュー中の版（doc 込み）。選んでいなければ null。 */
export interface KbSelectedVersionState {
  seq: number;
  detail: KbPageVersionDetail | null;
  loading: boolean;
  error: string | null;
}

/** 一覧取得の宛先。応答が着地してよいかの判定にこれを使う（useKbComments と同じ形）。 */
interface VersionsTarget {
  key: string;
  workspaceSlug: string;
  pageId: string;
}

function targetOf(
  workspaceSlug: string | undefined,
  pageId: string | undefined,
  open: boolean,
): VersionsTarget | null {
  // パネルが開いている間だけ取りに行く。useKbComments と違い、版一覧はエディタ側に
  // 常設のバッジ等を出さない（件数表示もしない設計）ので、閉じている間まで
  // 毎ページ引く理由が無い。
  if (!open || !workspaceSlug || !pageId) return null;
  return { key: `${workspaceSlug} ${pageId}`, workspaceSlug, pageId };
}

/**
 * useKbPageVersions はページ 1 枚の版一覧（新しい順）と、選択中の版のプレビュー・
 * 明示的な作成（版を残す）・復元を持つ。
 *
 * **一覧の取得は open（パネルが開いているか）ゲート付き** — useKbComments はバッジ表示の
 * ため開閉に関わらず常時取得するが、版一覧にはその用途が無いので、パネルを開いた人だけが
 * 必要とする。
 *
 * 一方で、**取得と書き込みの競合ガード（宛先 + 連番 + writeCount）は useKbComments と
 * 同じ理由で要る**と判断した。「開いているときだけ」が加わっても、宛先（workspaceSlug +
 * pageId + open）がパネルの開閉・ページの行き来のたびに変わり、飛んでいる応答が新しい
 * 宛先の画面を古い結果で上書きし得る事情そのものは無くならない。作成（版を残す）で
 * 一覧の先頭へ足した直後に、作成前から飛んでいた取得が遅れて着地すると、追加した版が
 * 一覧から消えてしまう — useKbComments の createThread と同じ理由で writeCount を見る。
 *
 * **プレビュー中の版（selected）・復元（restoreVersion）は open に依存しない。**
 * 選んだ版を表示するバナーはパネルを閉じても消えるべきではない
 * （「現在の版に戻る」ボタンだけが明示的な退出手段 — 画面設計の約束）。ページが変わった
 * ときだけ畳む。復元がページ本文（useKbPageDoc が持つ data.doc）へ与える影響は
 * ここでは扱わない — 呼び出し側（KbPage）が応答を useKbPageDoc.applyRestoredContent へ
 * 渡して反映する。
 */
export function useKbPageVersions(
  workspaceSlug: string | undefined,
  pageId: string | undefined,
  open: boolean,
) {
  const [state, setState] = useState<KbPageVersionsState>(EMPTY);
  const [selected, setSelected] = useState<KbSelectedVersionState | null>(null);
  const [restoring, setRestoring] = useState(false);

  // 一覧取得の宛先（open ゲート付き）。
  const active = useRef<VersionsTarget | null>(null);
  const seq = useRef(0);
  const writeCount = useRef(0);
  const target = targetOf(workspaceSlug, pageId, open);
  const targetKey = target?.key ?? null;

  // プレビュー・復元の宛先（open に依存しない、ページが決まっていれば常に有効）。
  const pageTarget = useRef<{ workspaceSlug: string; pageId: string } | null>(null);
  // 選択中の版の詳細取得（GET .../versions/:seq）の連番。ページ変更時にも進めて、
  // 飛んでいる取得を無効化する。
  const detailSeq = useRef(0);

  const load = useCallback(async (to: VersionsTarget) => {
    const request = ++seq.current;
    const writesAtStart = writeCount.current;
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const versions = await KbRepository.listPageVersions(to.workspaceSlug, to.pageId);
      if (active.current?.key !== to.key || seq.current !== request) return;
      if (writeCount.current !== writesAtStart) {
        // 取得中に書き込み（版を残す・復元）が成功していた。取得結果はそれより前の
        // スナップショットで古いので上書きしない（useKbComments と同じ理由）。
        setState((prev) => ({ ...prev, loading: false }));
        return;
      }
      setState({ versions, loading: false, error: null, saving: false });
    } catch {
      if (active.current?.key !== to.key || seq.current !== request) return;
      if (writeCount.current !== writesAtStart) {
        setState((prev) => ({ ...prev, loading: false }));
        return;
      }
      setState({ ...EMPTY, error: LOAD_FAILED });
    }
  }, []);

  useEffect(() => {
    active.current = target;
    if (!target) {
      seq.current += 1;
      setState(EMPTY);
      return;
    }
    void load(target);
    // target は毎描画で作り直すオブジェクトなので、鍵で比べる。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [targetKey, load]);

  // ページが変わったら、プレビュー中の版と進行中の詳細取得を畳む（前のページの版を
  // 持ち越さない）。パネルの開閉（open）には反応しない —
  // 上の JSDoc の通り、プレビューはパネルの開閉と独立に生きる。
  useEffect(() => {
    pageTarget.current = workspaceSlug && pageId ? { workspaceSlug, pageId } : null;
    detailSeq.current += 1;
    setSelected(null);
  }, [workspaceSlug, pageId]);

  /**
   * createVersion は今の本文を明示的な版として残す。**失敗は投げる**
   * （呼び出し側 KbPage がトーストで知らせる約束、他の書き込みと同じ）。
   * 成功したら一覧の先頭へ足す（新しい順を保つ）。
   *
   * パネルが開いているとき（= フォームが画面に存在するとき）にしか呼ばれない前提
   * なので、宛先は一覧取得と同じ active（open ゲート付き）を使う。
   */
  const createVersion = useCallback(async (note?: string): Promise<void> => {
    const to = active.current;
    if (!to) return;
    const request = seq.current;
    setState((prev) => ({ ...prev, saving: true }));
    try {
      const created = await KbRepository.createPageVersion(to.workspaceSlug, to.pageId, note);
      if (active.current?.key === to.key && seq.current === request) {
        writeCount.current += 1;
        setState((prev) => ({ ...prev, saving: false, versions: [created, ...prev.versions] }));
      }
    } catch (cause) {
      if (active.current?.key === to.key && seq.current === request) {
        setState((prev) => ({ ...prev, saving: false }));
      }
      throw cause;
    }
  }, []);

  /** selectVersion は行クリックで版のプレビューを始める（doc 込みの詳細を取得する）。 */
  const selectVersion = useCallback((versionSeq: number) => {
    const to = pageTarget.current;
    if (!to) return;
    const request = ++detailSeq.current;
    setSelected({ seq: versionSeq, detail: null, loading: true, error: null });
    KbRepository.getPageVersion(to.workspaceSlug, to.pageId, versionSeq)
      .then((detail) => {
        if (detailSeq.current !== request) return;
        setSelected({ seq: versionSeq, detail, loading: false, error: null });
      })
      .catch(() => {
        if (detailSeq.current !== request) return;
        setSelected({
          seq: versionSeq,
          detail: null,
          loading: false,
          error: 'この版を読み込めませんでした。',
        });
      });
  }, []);

  /** clearSelection は「現在の版に戻る（閉じる）」。プレビューをやめる。 */
  const clearSelection = useCallback(() => {
    detailSeq.current += 1; // 飛んでいる detail 取得があれば無効化する
    setSelected(null);
  }, []);

  /**
   * restoreVersion はその版を今の本文として復元する。**失敗は投げる**
   * （呼び出し側がトーストで知らせる約束）。
   *
   * 成功したら: プレビューを終える・一覧を引き直す（復元自体が新しい版になるため —
   * 一覧の先頭に足すだけでは「復元によって生まれた版」の note/author/createdAt が
   * 分からないので、作成のような手元での追記はせず、素直に取り直す）。**本体ページの
   * doc/lastEditedBy 等への反映はここでは行わない** — useKbPageVersions はページ本文の
   * 状態を持たない（useKbPageDoc の持ち物）。呼び出し側（KbPage）が応答を
   * useKbPageDoc.applyRestoredContent へ渡して反映する。
   */
  const restoreVersion = useCallback(
    async (versionSeq: number): Promise<KbPageContentSaveResult> => {
      const to = pageTarget.current;
      if (!to) throw new Error('ページが確定していないため復元できません。');
      setRestoring(true);
      try {
        const result = await KbRepository.restorePageVersion(to.workspaceSlug, to.pageId, versionSeq);
        const stillSamePage =
          pageTarget.current?.workspaceSlug === to.workspaceSlug &&
          pageTarget.current?.pageId === to.pageId;
        if (stillSamePage) {
          detailSeq.current += 1;
          setSelected(null);
          // パネルが開いていれば（active が有効なら）一覧を引き直す。閉じていれば
          // 次に開いたときの load() が最新を取るので、ここでは何もしない。
          if (active.current) {
            writeCount.current += 1;
            void load(active.current);
          }
        }
        return result;
      } finally {
        setRestoring(false);
      }
    },
    [load],
  );

  return {
    ...state,
    selected,
    restoring,
    createVersion,
    selectVersion,
    clearSelection,
    restoreVersion,
  };
}
