import { useState, useEffect, useCallback } from 'react';
import ProfileStatsRepository, { type ProfileStats } from '@/entities/user/api/profileStatsRepository';

export function useProfileStats() {
  const [stats, setStats] = useState<ProfileStats | null>(null);
  const [loading, setLoading] = useState(true);

  const loadStats = useCallback(async () => {
    try {
      const data = await ProfileStatsRepository.fetchStats();
      setStats(data);
    } catch {
      // 統計取得エラーは無視
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadStats();
  }, [loadStats]);

  return { stats, loading };
}
