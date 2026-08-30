import { useEffect, useState } from 'react';
import { AdminMemberRepository, type WorkspaceLearningSummary } from '@/entities/member';

interface Options {
  enabled?: boolean;
}

/**
 * useWorkspaceLearningSummary — 自社（自ワークスペース）メンバーの学習状況サマリーを取得する。
 * company_admin のホームのサイドバー用。enabled=false のときはリクエストを発行しない。
 */
export function useWorkspaceLearningSummary(options?: Options) {
  const enabled = options?.enabled ?? true;
  const [summary, setSummary] = useState<WorkspaceLearningSummary | null>(null);
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
