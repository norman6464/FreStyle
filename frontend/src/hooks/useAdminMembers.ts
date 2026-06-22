import { useCallback, useEffect, useState } from 'react';
import AdminMemberRepository, { Member } from '../repositories/AdminMemberRepository';
import { getApiError } from '../utils/classifyApiError';

// backend のエラーコードを日本語メッセージにする。cannot_manage_self は自己操作の防止。
function messageFor(e: unknown, fallback: string): string {
  const code = getApiError(e).serverCode;
  if (code === 'cannot_manage_self') return '自分自身は無効化・削除できません';
  if (code === 'forbidden') return 'この従業員を操作する権限がありません';
  if (code === 'member_not_found') return '対象の従業員が見つかりません';
  return fallback;
}

/**
 * useAdminMembers — 従業員管理ページの状態管理フック。
 * 自社の従業員一覧の取得、AI 利用可否の個別上書き、アカウントの有効/無効、論理削除を扱う。
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

  // アカウントの有効/無効を切り替える。楽観的更新 + 失敗時ロールバック。
  const setActive = useCallback(
    async (userId: number, active: boolean) => {
      setUpdatingId(userId);
      setError(null);
      setMembers((prev) => prev.map((m) => (m.id === userId ? { ...m, isActive: active } : m)));
      try {
        await AdminMemberRepository.updateActive(userId, active);
      } catch (e) {
        setMembers((prev) => prev.map((m) => (m.id === userId ? { ...m, isActive: !active } : m)));
        setError(messageFor(e, 'アカウント状態の更新に失敗しました'));
      } finally {
        setUpdatingId(null);
      }
    },
    [],
  );

  // 従業員を論理削除する。成功したら一覧から除く。
  const remove = useCallback(async (userId: number) => {
    setUpdatingId(userId);
    setError(null);
    try {
      await AdminMemberRepository.remove(userId);
      setMembers((prev) => prev.filter((m) => m.id !== userId));
    } catch (e) {
      setError(messageFor(e, '従業員の削除に失敗しました'));
    } finally {
      setUpdatingId(null);
    }
  }, []);

  return { members, loading, error, updatingId, setAiAccess, setActive, remove, reload: load };
}
