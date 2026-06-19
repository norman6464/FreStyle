import { useState, useEffect } from 'react';
import DashboardRepository from '../repositories/DashboardRepository';
import type { UserDashboard } from '../types';

interface Options {
  enabled?: boolean;
}

/** useUserDashboard はログインユーザーの学習サマリーを取得する。enabled=false のときはリクエストを発行しない。 */
export function useUserDashboard(options?: Options) {
  const enabled = options?.enabled ?? true;
  const [dashboard, setDashboard] = useState<UserDashboard | null>(null);
  const [loading, setLoading] = useState(enabled);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!enabled) return;
    let cancelled = false;
    setLoading(true);
    DashboardRepository.get()
      .then((data) => {
        if (!cancelled) setDashboard(data);
      })
      .catch(() => {
        if (!cancelled) setError('ダッシュボードの取得に失敗しました');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, [enabled]);

  return { dashboard, loading, error };
}
