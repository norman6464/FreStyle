import { useCallback, useEffect, useRef, useState } from 'react';
import {
  CourseRepository,
  type MaterialGrant,
  type MaterialPrincipal,
} from '@/entities/course';
import type { ShareRole, ShareRow, SharePrincipal } from '@/features/permission-sharing';

export interface CourseShareState {
  rows: ShareRow[];
  candidates: SharePrincipal[];
  loading: boolean;
  /** 失敗の理由。null なら失敗していない。 */
  error: string | null;
  /** 書き込みが飛んでいる間 true。 */
  saving: boolean;
}

const EMPTY: CourseShareState = {
  rows: [],
  candidates: [],
  loading: false,
  error: null,
  saving: false,
};

const LOAD_FAILED =
  '権限を読めませんでした。通信が切れたか、このコースの権限を変える立場でなくなっています。開き直すと最新の状態が出ます。';
const WRITE_FAILED = '権限を変えられませんでした。もう一度お試しください。';

/**
 * useCourseShare はコース 1 つの共有設定（コース単位の付与）を読み書きする。
 *
 * # 2 本引いて突き合わせる理由
 *
 * 一覧は主体を ID でしか返さない。名前の正本は kind ごとに別の表にあり、backend は
 * それを相手の一覧側で解決している。ここで ID を突き合わせて名前を付ける。
 *
 * **突き合わない ID は行を落とさず ID のまま出す。** 引いた直後に主体が消えた場合など、
 * 名前が引けないことは起こる。そこで行を消すと、消せない権限が画面から見えないまま残り、
 * 誰が編集できるのかを人が説明できなくなる。
 *
 * # 応答は「いま見ているコース宛て」だけを受け取る
 *
 * 飛んでいる要求より先に別のコースへ移ったり、パネルを閉じたりする。着地してよいかは
 * **宛先（コース ID）と要求の連番の両方**で決める。片方だけでは足りない:
 *
 *   - 宛先だけ: 同じコースへの 2 本目（書き込みのあとの引き直し）が飛んでいる最中に
 *     1 本目が着地すると、古い一覧で上書きされる
 *   - 連番だけ: 書き込みのあとの引き直しは書き込みを始めた時点の宛先へ向かうので、
 *     その間に別のコースへ移ると、旧コースを引き直して新しいコースの結果を捨てる
 */
export function useCourseShare(courseId: number | undefined) {
  const [state, setState] = useState<CourseShareState>(EMPTY);

  // いま見ている宛先。応答が着地してよいかをこれで判定する。
  const active = useRef<number | undefined>(undefined);
  // 要求の連番。**宛先だけでは足りない。** 同じコースへの 2 本目（書き込みのあとの
  // 引き直し）が飛んでいる最中に 1 本目が着地すると、古い一覧で上書きされる。
  const seq = useRef(0);

  const load = useCallback(async (target: number) => {
    const request = ++seq.current;
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      // 2 本は独立しているので同時に投げる（順に待つと開くたびに待ち時間が倍になる）。
      const [grants, principals] = await Promise.all([
        CourseRepository.listGrants(target),
        CourseRepository.listGrantablePrincipals(target),
      ]);
      if (active.current !== target || seq.current !== request) return;
      const granted = new Set(grants.map((grant) => grant.principalId));
      setState({
        rows: joinRows(grants, principals),
        candidates: principals.filter((principal) => !granted.has(principal.id)),
        loading: false,
        error: null,
        saving: false,
      });
    } catch {
      if (active.current !== target || seq.current !== request) return;
      setState({ ...EMPTY, error: LOAD_FAILED });
    }
  }, []);

  useEffect(() => {
    active.current = courseId;
    if (courseId === undefined) {
      // 閉じた・コースが決まっていない。連番を進めて、飛んでいる応答を無効にする
      // （同じコースをすぐ開き直しても、前回の応答は着地しない）。
      seq.current += 1;
      setState(EMPTY);
      return;
    }
    void load(courseId);
  }, [courseId, load]);

  /**
   * mutate は書き込みを 1 回行い、成功したかを返す。
   * 宛先は押した時点のもの。引き直しは、その宛先がまだ見られている場合だけ行う。
   */
  const mutate = useCallback(
    async (run: (target: number) => Promise<unknown>): Promise<boolean> => {
      const target = active.current;
      if (target === undefined) return false;
      setState((prev) => ({ ...prev, saving: true, error: null }));
      try {
        await run(target);
      } catch {
        if (active.current !== target) return false;
        setState((prev) => ({ ...prev, saving: false, error: WRITE_FAILED }));
        return false;
      }
      if (active.current !== target) return false;
      await load(target);
      return true;
    },
    [load],
  );

  const grant = useCallback(
    (principalId: string, role: ShareRole) =>
      mutate((target) => CourseRepository.grantRole(target, principalId, role)),
    [mutate],
  );

  const revoke = useCallback(
    (principalId: string) => mutate((target) => CourseRepository.revokeRole(target, principalId)),
    [mutate],
  );

  return { ...state, grant, revoke };
}

/** joinRows は張った権限に表示名を付ける（付かない行も落とさない）。 */
function joinRows(grants: MaterialGrant[], principals: MaterialPrincipal[]): ShareRow[] {
  const principalsById = new Map(principals.map((principal) => [principal.id, principal]));
  return grants.map((grant) => {
    const principal = principalsById.get(grant.principalId);
    return {
      principalId: grant.principalId,
      role: grant.role,
      name: principal?.name ?? '',
      kind: principal?.kind ?? 'unknown',
    };
  });
}
