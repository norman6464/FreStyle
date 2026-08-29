import { useEffect, useState } from 'react';
import { CompanyRepository } from '@/entities/company';

/** 運営ダッシュボードの概況サマリ（既存 API の集計）。 */
export interface AdminDashboardSummary {
  companyTotal: number;
  companyActive: number;
  companyInactive: number;
}

/**
 * useAdminDashboard — 運営（super_admin）ダッシュボードの概況を取得するフック。
 * 専用の集計 API は持たず、会社一覧をクライアントで集計する（会社数が少ない前提）。
 * `enabled=false`（super_admin 以外 / 認証確認中）のときは admin API を叩かない。
 */
export function useAdminDashboard(enabled = true) {
  const [summary, setSummary] = useState<AdminDashboardSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!enabled) {
      setLoading(false);
      return;
    }
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const companies = await CompanyRepository.list();
        if (cancelled) return;
        const companyActive = companies.filter((c) => c.isActive).length;
        setSummary({
          companyTotal: companies.length,
          companyActive,
          companyInactive: companies.length - companyActive,
        });
      } catch {
        if (!cancelled) setError('ダッシュボードの取得に失敗しました');
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [enabled]);

  return { summary, loading, error };
}
