import { useEffect, useState } from 'react';

import { CompanyRepository, CompanyStat } from '@/entities/company';
import { logger } from '@/shared/lib/logger';

/**
 * useAdminCompanies — 会社横断ビュー（super_admin 専用）の状態管理フック。
 * 各社のメンバー集計の取得と、会社アカウントの有効/無効の切り替えを扱う。
 */
export function useAdminCompanies() {
  const [companies, setCompanies] = useState<CompanyStat[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updatingId, setUpdatingId] = useState<number | null>(null);

  useEffect(() => {
    CompanyRepository.listStats()
      .then(setCompanies)
      .catch((e) => {
        setError('会社一覧の取得に失敗しました');
        logger.error(e);
      })
      .finally(() => setLoading(false));
  }, []);

  // 会社アカウントの有効/無効を切り替える（super_admin 専用）。楽観的更新 + 失敗時ロールバック。
  const setActive = async (id: number, active: boolean) => {
    setUpdatingId(id);
    setError(null);
    setCompanies((prev) => prev.map((c) => (c.id === id ? { ...c, isActive: active } : c)));
    try {
      await CompanyRepository.updateActive(id, active);
    } catch (e) {
      logger.error(e);
      setCompanies((prev) => prev.map((c) => (c.id === id ? { ...c, isActive: !active } : c)));
      setError('会社状態の更新に失敗しました');
    } finally {
      setUpdatingId(null);
    }
  };

  return { companies, loading, error, updatingId, setActive };
}
