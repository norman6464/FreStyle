import { useEffect, useState, useMemo } from 'react';
import type { ScoreHistoryItem } from '../types';
import { useAiChat } from './useAiChat';

export interface ScoreHistoryItemWithDelta extends ScoreHistoryItem {
  delta: number | null;
}

const FILTERS = ['すべて', '練習', 'フリー'] as const;
export type FilterType = (typeof FILTERS)[number];
export { FILTERS };

function isPracticeSession(title: string): boolean {
  return title.startsWith('練習:') || title.startsWith('練習：');
}

export function useScoreHistory() {
  const [history, setHistory] = useState<ScoreHistoryItem[]>([]);
  const [filter, setFilter] = useState<FilterType>('すべて');
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
    return history.filter((item) => {
      if (filter === '練習') return isPracticeSession(item.sessionTitle);
      if (filter === 'フリー') return !isPracticeSession(item.sessionTitle);
      return true;
    });
  }, [history, filter]);

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
    loading,
    latestSession,
    averageScore,
    weakestAxis,
    selectedSession,
    setSelectedSession,
  };
}
