import { useCallback, useEffect, useRef, useState } from 'react';
import {
  NoteRepository,
  type NoteGrantablePrincipal,
  type NoteGrantRole,
  type NotePageGrant,
} from '@/entities/note';

/** 一覧に並ぶ 1 行（張った権限と、その相手の表示名を突き合わせたもの）。 */
export interface NoteShareRow {
  principalId: string;
  role: NoteGrantRole;
  /** 表示名。引けなかった相手は空文字。 */
  name: string;
  kind: NoteGrantablePrincipal['kind'] | 'unknown';
}

export interface NoteShareState {
  rows: NoteShareRow[];
  /** まだ権限を張っていない相手（追加の候補）。 */
  candidates: NoteGrantablePrincipal[];
  loading: boolean;
  /** 失敗の理由。null なら失敗していない。 */
  error: string | null;
  /** 書き込み（付与・取り消し）が飛んでいる間 true。 */
  saving: boolean;
}

const LOAD_FAILED = '権限を読めませんでした。通信が切れたか、このページの権限を変える立場でなくなっています。開き直すと最新の状態が出ます。';
const WRITE_FAILED = '権限を変えられませんでした。もう一度お試しください。';

/**
 * useNoteShare はページ 1 枚の共有設定（ページ単位の付与）を読み書きする。
 *
 * # 2 本引いて突き合わせる理由
 *
 * 一覧（grants）は主体を ID でしか返さない。名前の正本は kind ごとに別の表にあり、
 * backend はそれを相手の一覧（principals）側で解決している。ここで ID を突き合わせて
 * 名前を付ける。
 *
 * **突き合わない ID は行を落とさず ID のまま出す。** 引いた直後に主体が消えた場合など、
 * 名前が引けないことは起こる。そこで行を消すと、消せない権限が画面から見えないまま残り、
 * 誰が見られるのかを人が説明できなくなる。
 *
 * # 一覧が空でも「誰も見られない」ではない
 *
 * 返るのはこのページ自身に張った行だけで、ワークスペース / スペース / 祖先のページから
 * 届いている相手は含まれない。空 = この段では何も足していない、という意味しか無い。
 * それを画面に書くのは呼び出し側（NoteSharePanel）の責任。
 */
export function useNoteShare(workspaceSlug: string | undefined, pageId: string | undefined) {
  const [state, setState] = useState<NoteShareState>({
    rows: [],
    candidates: [],
    loading: false,
    error: null,
    saving: false,
  });

  // 速く開き閉めしたときに、古い応答が新しいページの結果を上書きするのを防ぐ。
  const generation = useRef(0);

  const load = useCallback(async () => {
    if (!workspaceSlug || !pageId) return;
    const gen = ++generation.current;
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      // 2 本は独立しているので同時に投げる（順に待つと開くたびに待ち時間が倍になる）。
      const [grants, principals] = await Promise.all([
        NoteRepository.listPageGrants(workspaceSlug, pageId),
        NoteRepository.listGrantablePrincipals(workspaceSlug, pageId),
      ]);
      if (gen !== generation.current) return;
      setState({
        rows: joinRows(grants, principals),
        candidates: principals.filter((p) => !grants.some((g) => g.principalId === p.id)),
        loading: false,
        error: null,
        saving: false,
      });
    } catch {
      if (gen !== generation.current) return;
      setState({ rows: [], candidates: [], loading: false, error: LOAD_FAILED, saving: false });
    }
  }, [workspaceSlug, pageId]);

  useEffect(() => {
    void load();
  }, [load]);

  // 書き込みのあとは必ず引き直す。楽観的に画面だけ書き換えると、サーバーが
  // 別の答えを返したとき（弱い役割を張っても上位の役割で上書きされない等）に
  // 画面と実態がずれたまま残る。
  const mutate = useCallback(
    async (run: () => Promise<unknown>) => {
      setState((prev) => ({ ...prev, saving: true, error: null }));
      try {
        await run();
      } catch {
        setState((prev) => ({ ...prev, saving: false, error: WRITE_FAILED }));
        return;
      }
      await load();
    },
    [load],
  );

  const grant = useCallback(
    (principalId: string, role: NoteGrantRole) => {
      if (!workspaceSlug || !pageId) return Promise.resolve();
      return mutate(() => NoteRepository.grantPageRole(workspaceSlug, pageId, principalId, role));
    },
    [mutate, workspaceSlug, pageId],
  );

  const revoke = useCallback(
    (principalId: string) => {
      if (!workspaceSlug || !pageId) return Promise.resolve();
      return mutate(() => NoteRepository.revokePageRole(workspaceSlug, pageId, principalId));
    },
    [mutate, workspaceSlug, pageId],
  );

  return { ...state, grant, revoke, reload: load };
}

/** joinRows は張った権限に表示名を付ける（付かない行も落とさない）。 */
function joinRows(
  grants: NotePageGrant[],
  principals: NoteGrantablePrincipal[],
): NoteShareRow[] {
  const byID = new Map(principals.map((p) => [p.id, p]));
  return grants.map((g) => {
    const p = byID.get(g.principalId);
    return {
      principalId: g.principalId,
      role: g.role,
      name: p?.name ?? '',
      kind: p?.kind ?? 'unknown',
    };
  });
}
