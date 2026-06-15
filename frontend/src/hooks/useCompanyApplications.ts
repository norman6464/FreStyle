import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  CompanyApplicationRepository,
  CompanyApplication,
  CompanyApplicationStatus,
} from '../repositories/CompanyApplicationRepository';

/**
 * useCompanyApplications — 利用申請（企業）の承認ページの状態管理フック。
 * super_admin 専用。一覧取得と status 更新（承認 / 却下 / 保留）を扱う。
 */
export function useCompanyApplications() {
  const [applications, setApplications] = useState<CompanyApplication[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updatingId, setUpdatingId] = useState<number | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setApplications(await CompanyApplicationRepository.adminList());
    } catch {
      setError('利用申請の取得に失敗しました');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  // 承認待ちの件数（サイドバー/ヘッダーのバッジ表示に使う）。
  const pendingCount = useMemo(
    () => applications.filter((a) => a.status === 'pending').length,
    [applications],
  );

  // status を更新する。楽観的に反映し、失敗時は更新前の status へ確実にロールバックする
  // （再取得 load 依存だと、再取得自体が失敗したときにサーバー未反映の状態が画面に残るため）。
  const setStatus = useCallback(
    async (id: number, status: CompanyApplicationStatus): Promise<boolean> => {
      setUpdatingId(id);
      setError(null);
      let previous: CompanyApplicationStatus | undefined;
      setApplications((prev) =>
        prev.map((a) => {
          if (a.id !== id) return a;
          previous = a.status;
          return { ...a, status };
        }),
      );
      try {
        await CompanyApplicationRepository.adminUpdateStatus(id, status);
        return true;
      } catch {
        if (previous !== undefined) {
          setApplications((prev) =>
            prev.map((a) => (a.id === id ? { ...a, status: previous as CompanyApplicationStatus } : a)),
          );
        }
        setError('ステータスの更新に失敗しました');
        return false;
      } finally {
        setUpdatingId(null);
      }
    },
    [],
  );

  return { applications, pendingCount, loading, error, updatingId, setStatus, reload: load };
}
