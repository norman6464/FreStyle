import { useEffect, useState, useMemo } from 'react';
import type { ScoreHistoryItem } from '../types';
import { useAiChat } from './useAiChat';

export interface ScoreHistoryItemWithDelta extends ScoreHistoryItem {
  delta: number | null;
}

const FILTERS = ['すべて', '練習', 'フリー'] as const;
export type FilterType = (typeof FILTERS)[number];
export { FILTERS };

const PERIOD_FILTERS = ['全期間', '1週間', '1ヶ月', '3ヶ月'] as const;
export type PeriodFilterType = (typeof PERIOD_FILTERS)[number];
export { PERIOD_FILTERS };

function isPracticeSession(title: string): boolean {
  return title.startsWith('練習:') || title.startsWith('練習：');
}

function getPeriodCutoff(period: PeriodFilterType): Date | null {
  const now = new Date();
  switch (period) {
    case '1週間':
      return new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
    case '1ヶ月':
      return new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
    case '3ヶ月':
      return new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000);
    default:
      return null;
  }
}

export function useScoreHistory() {
  const [history, setHistory] = useState<ScoreHistoryItem[]>([]);
  const [filter, setFilter] = useState<FilterType>('すべて');
  const [periodFilter, setPeriodFilter] = useState<PeriodFilterType>('全期間');
  const [selectedSession, setSelectedSession] = useState<ScoreHistoryItem | null>(null);
  const { fetchScoreHistory, loading } = useAiChat();

  useEffect(() => {
    const loadHistory = async () => {
      const data = await fetchScoreHistory();
      const items = Array.isArray(data) ? data : [];
      setHistory(items.map((item: ScoreHistoryItem) => ({
        ...item,
        scores: Array.isArray(item.scores) ? item.scores : [],
      })));
    };
    loadHistory();
  }, [fetchScoreHistory]);

  const filteredHistory = useMemo(() => {
    const cutoff = getPeriodCutoff(periodFilter);
    return history.filter((item) => {
      if (filter === '練習' && !isPracticeSession(item.sessionTitle)) return false;
      if (filter === 'フリー' && isPracticeSession(item.sessionTitle)) return false;
      if (cutoff && new Date(item.createdAt) < cutoff) return false;
      return true;
    });
  }, [history, filter, periodFilter]);

  const latestSession = history.length > 0 ? history[history.length - 1] : null;

  const averageScore = useMemo(() => {
    if (history.length === 0) return 0;
    return Math.round((history.reduce((sum, h) => sum + h.overallScore, 0) / history.length) * 10) / 10;
  }, [history]);

  const weakestAxis = useMemo(() => {
    if (!latestSession) return null;
    return [...latestSession.scores].sort((a, b) => a.score - b.score)[0] ?? null;
  }, [latestSession]);

  const filteredHistoryWithDelta = useMemo((): ScoreHistoryItemWithDelta[] => {
    return filteredHistory.map((item) => {
      const originalIndex = history.indexOf(item);
      const prevItem = originalIndex > 0 ? history[originalIndex - 1] : null;
      const delta = prevItem
        ? Math.round((item.overallScore - prevItem.overallScore) * 10) / 10
        : null;
      return { ...item, delta };
    });
  }, [filteredHistory, history]);

  return {
    history,
    filteredHistory,
    filteredHistoryWithDelta,
    filter,
    setFilter,
    periodFilter,
    setPeriodFilter,
    loading,
    latestSession,
    averageScore,
    weakestAxis,
    selectedSession,
    setSelectedSession,
  };
}
