import { useState, useEffect, useMemo } from 'react';
import { MenuRepository } from '../repositories/MenuRepository';
import type { ScoreHistory } from '../types';
import { getMonday } from '../utils/weekUtils';

interface ChatStats {
  chatPartnerCount: number;
}

export function useMenuData() {
  const [stats, setStats] = useState<ChatStats | null>(null);
  const [totalUnread, setTotalUnread] = useState(0);
  const [latestScore, setLatestScore] = useState<ScoreHistory | null>(null);
  const [allScores, setAllScores] = useState<ScoreHistory[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchAll = async () => {
      try {
        const [statsData, roomsData, scoresData] = await Promise.allSettled([
          MenuRepository.fetchChatStats(),
          MenuRepository.fetchChatRooms(),
          MenuRepository.fetchScoreHistory(),
        ]);

        if (statsData.status === 'fulfilled') {
          setStats(statsData.value);
        }

        if (roomsData.status === 'fulfilled' && Array.isArray(roomsData.value?.chatUsers)) {
          const unread = roomsData.value.chatUsers.reduce(
            (sum, u) => sum + u.unreadCount, 0
          );
          setTotalUnread(unread);
        }

        if (scoresData.status === 'fulfilled' && Array.isArray(scoresData.value) && scoresData.value.length > 0) {
          setLatestScore(scoresData.value[0]);
          setAllScores(scoresData.value);
        }
      } catch {
        // サイレントに処理
      } finally {
        setLoading(false);
      }
    };
    fetchAll();
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
    stats,
    totalUnread,
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
