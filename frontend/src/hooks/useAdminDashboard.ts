import { useEffect, useState } from 'react';
import { CompanyRepository } from '@/entities/company';
import {
  CompanyApplicationRepository,
  CompanyApplication,
} from '@/entities/company/api/companyApplicationRepository';

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
        const [companies, applications] = await Promise.all([
          CompanyRepository.list(),
          CompanyApplicationRepository.adminList(),
        ]);
        if (cancelled) return;
        const companyActive = companies.filter((c) => c.isActive).length;
        // 直近順は backend の返却順に依存しないよう createdAt 降順で明示的にソートする。
        const pending = applications
          .filter((a) => a.status === 'pending')
          .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
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
  }, [enabled]);

  return { summary, loading, error };
}
