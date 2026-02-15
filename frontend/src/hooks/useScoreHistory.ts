import { useEffect, useState, useMemo } from 'react';
import { useAiChat } from './useAiChat';

interface AxisScore {
  axis: string;
  score: number;
  comment: string;
}

export interface ScoreHistoryItem {
  sessionId: number;
  sessionTitle: string;
  overallScore: number;
  scores: AxisScore[];
  createdAt: string;
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

  return {
    history,
    filteredHistory,
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
