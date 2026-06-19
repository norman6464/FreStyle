import { useState, useEffect } from 'react';
import DashboardRepository from '../repositories/DashboardRepository';
import type { UserDashboard } from '../types';

/** useUserDashboard はログインユーザーの学習サマリーを取得する。 */
export function useUserDashboard() {
  const [dashboard, setDashboard] = useState<UserDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
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
  }, []);

  return { dashboard, loading, error };
}
