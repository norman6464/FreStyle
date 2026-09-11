import { useCallback, useEffect, useRef, useState } from 'react';
import { KbRepository, type KbAdminWorkspaceMember, type KbGrantRole } from '@/entities/kb';
import { getApiError } from '@/shared/lib/classifyApiError';

export interface KbAdminMembersState {
  members: KbAdminWorkspaceMember[];
  loading: boolean;
  /** 失敗の理由。null なら失敗していない。'forbidden' は admin でない（画面ごと出し分ける）。 */
  error: 'forbidden' | 'unknown' | null;
  /** いま処理中の相手の userId。行ごとの操作ボタンを閉じるのに使う（useKbComments の saving と同じ思想）。 */
  busyUserId: number | null;
}

const EMPTY: KbAdminMembersState = { members: [], loading: false, error: null, busyUserId: null };

/**
 * useKbAdminMembers はメンバー管理画面（段 7）の一覧取得と、役割変更・停止・復帰・削除の
 * 書き込みをまとめる。
 *
 * workspaceSlug が変わったら（違うワークスペースの管理画面を開き直した）取得をやり直す。
 * 応答は**要求を始めたときの宛先**が今も見られているときだけ反映する
 * （useKbComments と同じ理由 — 先に別ワークスペースへ移っていたら古い応答で上書きしない）。
 *
 * 書き込みはすべて**楽観更新をせず、成功した後に一覧を丸ごと引き直す**。停止・復帰・役割変更・
 * 削除のどれも「最後の admin」の可否が他の行にも影響しうる操作の余地があり
 * （例えば admin を 1 人に減らした直後は他の行の削除ボタンの意味が変わる）、
 * 差分をこちらで組み立てるより引き直す方が確実。
 */
export function useKbAdminMembers(workspaceSlug: string | undefined) {
  const [state, setState] = useState<KbAdminMembersState>(EMPTY);
  const active = useRef<string | null>(null);
  const seq = useRef(0);

  const load = useCallback(async (slug: string) => {
    const request = ++seq.current;
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const members = await KbRepository.fetchAdminMembers(slug);
      if (active.current !== slug || seq.current !== request) return;
      setState({ members, loading: false, error: null, busyUserId: null });
    } catch (cause) {
      if (active.current !== slug || seq.current !== request) return;
      const forbidden = getApiError(cause).status === 403;
      setState({ ...EMPTY, error: forbidden ? 'forbidden' : 'unknown' });
    }
  }, []);

  useEffect(() => {
    active.current = workspaceSlug ?? null;
    if (!workspaceSlug) {
      seq.current += 1;
      setState(EMPTY);
      return;
    }
    void load(workspaceSlug);
  }, [workspaceSlug, load]);

  const retry = useCallback(() => {
    if (workspaceSlug) void load(workspaceSlug);
  }, [workspaceSlug, load]);

  /** mutate は 1 回の書き込みを行い、成功したら一覧を引き直す。失敗は投げる（知らせは呼び出し側）。 */
  const mutate = useCallback(
    async (userId: number, run: (slug: string) => Promise<void>) => {
      const slug = active.current;
      if (!slug) return;
      setState((prev) => ({ ...prev, busyUserId: userId }));
      try {
        await run(slug);
        if (active.current === slug) await load(slug);
      } finally {
        if (active.current === slug) setState((prev) => ({ ...prev, busyUserId: null }));
      }
    },
    [load],
  );

  const changeRole = useCallback(
    (principalId: string, userId: number, role: KbGrantRole | null) =>
      mutate(userId, (slug) =>
        role === null
          ? KbRepository.revokeWorkspaceRole(slug, principalId)
          : KbRepository.grantWorkspaceRole(slug, principalId, role),
      ),
    [mutate],
  );

  const suspend = useCallback(
    (userId: number) => mutate(userId, (slug) => KbRepository.suspendMember(slug, userId)),
    [mutate],
  );

  const restore = useCallback(
    (userId: number) => mutate(userId, (slug) => KbRepository.restoreMember(slug, userId)),
    [mutate],
  );

  const remove = useCallback(
    (userId: number) => mutate(userId, (slug) => KbRepository.removeMember(slug, userId)),
    [mutate],
  );

  return { ...state, retry, changeRole, suspend, restore, remove };
}
