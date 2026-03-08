import { useState, useEffect, useCallback } from 'react';
import { RankingRepository } from '../repositories/RankingRepository';
import { Ranking } from '../types';

export function useRanking() {
  const [ranking, setRanking] = useState<Ranking | null>(null);
  const [period, setPeriod] = useState<'weekly' | 'monthly'>('weekly');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    RankingRepository.fetchRanking(period)
      .then((data) => {
        if (!cancelled) setRanking(data);
      })
      .catch(() => {
        if (!cancelled) setError('ランキングの取得に失敗しました');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, [period]);

  const changePeriod = useCallback((newPeriod: 'weekly' | 'monthly') => {
    setPeriod(newPeriod);
  }, []);

  return { ranking, period, changePeriod, loading, error };
}
