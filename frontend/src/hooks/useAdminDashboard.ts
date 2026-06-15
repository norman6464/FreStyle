import { useEffect, useState } from 'react';
import CompanyRepository from '../repositories/CompanyRepository';
import {
  CompanyApplicationRepository,
  CompanyApplication,
} from '../repositories/CompanyApplicationRepository';

/** 運営ダッシュボードの概況サマリ（既存 API の集計）。 */
export interface AdminDashboardSummary {
  companyTotal: number;
  companyActive: number;
  companyInactive: number;
  applicationTotal: number;
  pendingApplications: number;
  /** 承認待ちの申請（新しい順・最大 5 件）。すぐ承認できるよう一覧に出す。 */
  recentPending: CompanyApplication[];
}

/**
 * useAdminDashboard — 運営（super_admin）ダッシュボードの概況を取得するフック。
 * 専用の集計 API は持たず、会社一覧 + 利用申請一覧をクライアントで集計する（会社数が少ない前提）。
 */
export function useAdminDashboard() {
  const [summary, setSummary] = useState<AdminDashboardSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const [companies, applications] = await Promise.all([
          CompanyRepository.list(),
          CompanyApplicationRepository.adminList(),
        ]);
        if (cancelled) return;
        const companyActive = companies.filter((c) => c.isActive).length;
        const pending = applications.filter((a) => a.status === 'pending');
        setSummary({
          companyTotal: companies.length,
          companyActive,
          companyInactive: companies.length - companyActive,
          applicationTotal: applications.length,
          pendingApplications: pending.length,
          recentPending: pending.slice(0, 5),
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
  }, []);

  return { summary, loading, error };
}
