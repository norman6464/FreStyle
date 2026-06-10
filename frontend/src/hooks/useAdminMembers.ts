import { useCallback, useEffect, useState } from 'react';
import AdminMemberRepository, { Member } from '../repositories/AdminMemberRepository';

/**
 * useAdminMembers — 従業員管理ページの状態管理フック。
 * 自社の従業員一覧の取得と、各従業員の AI 利用可否（個別上書き）の更新を扱う。
 */
export function useAdminMembers() {
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updatingId, setUpdatingId] = useState<number | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setMembers(await AdminMemberRepository.listMembers());
    } catch {
      setError('従業員一覧の取得に失敗しました');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  // AI 利用可否を個別更新する。enabled=null で会社設定に従う状態へ戻す。
  const setAiAccess = useCallback(
    async (userId: number, enabled: boolean | null) => {
      setUpdatingId(userId);
      // 楽観的に反映し、失敗したら再取得して整合を取る。
      setMembers((prev) => prev.map((m) => (m.id === userId ? { ...m, aiChatEnabled: enabled } : m)));
      try {
        await AdminMemberRepository.updateAiAccess(userId, enabled);
      } catch {
        await load();
      } finally {
        setUpdatingId(null);
      }
    },
    [load],
  );

  return { members, loading, error, updatingId, setAiAccess, reload: load };
}
