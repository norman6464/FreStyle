import { useState, useEffect, useMemo } from 'react';
import { MenuRepository } from '../repositories/MenuRepository';
import type { ScoreHistory } from '../types';
import { getMonday } from '../utils/weekUtils';

export function useMenuData() {
  const [latestScore, setLatestScore] = useState<ScoreHistory | null>(null);
  const [allScores, setAllScores] = useState<ScoreHistory[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    const fetchAll = async () => {
      try {
        const scoresData = await MenuRepository.fetchScoreHistory().catch(() => null);

        if (cancelled) return;

        if (scoresData && Array.isArray(scoresData) && scoresData.length > 0) {
          setLatestScore(scoresData[0]);
          setAllScores(scoresData);
        }
      } catch {
        // サイレントに処理
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };
    fetchAll();
    return () => { cancelled = true; };
  }, []);

  const totalSessions = allScores.length;
  const averageScore = useMemo(() => {
    if (totalSessions === 0) return 0;
    return Math.round((allScores.reduce((sum, s) => sum + s.overallScore, 0) / totalSessions) * 10) / 10;
  }, [allScores, totalSessions]);
  const practiceDates = useMemo(() => {
    return [...new Set(allScores.map(s => s.createdAt.split('T')[0]))];
  }, [allScores]);
  const uniqueDays = practiceDates.length;

  const sessionsThisWeek = useMemo(() => {
    const monday = getMonday(new Date());
    return allScores.filter(s => new Date(s.createdAt) >= monday).length;
  }, [allScores]);

  return {
    latestScore,
    allScores,
    totalSessions,
    averageScore,
    uniqueDays,
    practiceDates,
    sessionsThisWeek,
    loading,
  };
}
