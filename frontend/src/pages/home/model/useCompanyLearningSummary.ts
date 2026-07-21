import { useEffect, useState } from 'react';
import { AdminMemberRepository, type CompanyLearningSummary } from '@/entities/member';

interface Options {
  enabled?: boolean;
}

/**
 * useCompanyLearningSummary — 自社メンバーの学習状況サマリーを取得する(FRESTYLE-103)。
 * company_admin のホームのサイドバー用。enabled=false のときはリクエストを発行しない。
 */
export function useCompanyLearningSummary(options?: Options) {
  const enabled = options?.enabled ?? true;
  const [summary, setSummary] = useState<CompanyLearningSummary | null>(null);
  const [loading, setLoading] = useState(enabled);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!enabled) return;
    let cancelled = false;
    setLoading(true);
    AdminMemberRepository.learningSummary()
      .then((data) => {
        if (!cancelled) setSummary(data);
      })
      .catch(() => {
        if (!cancelled) setError('学習状況の取得に失敗しました');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [enabled]);

  return { summary, loading, error };
}
