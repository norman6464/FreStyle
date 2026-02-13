import { useState, useEffect, useMemo } from 'react';
import { MenuRepository } from '../repositories/MenuRepository';

interface ChatStats {
  chatPartnerCount: number;
}

interface ScoreHistory {
  sessionId: number;
  sessionTitle: string;
  overallScore: number;
  createdAt: string;
}

export function useMenuData() {
  const [stats, setStats] = useState<ChatStats | null>(null);
  const [totalUnread, setTotalUnread] = useState(0);
  const [latestScore, setLatestScore] = useState<ScoreHistory | null>(null);
  const [allScores, setAllScores] = useState<ScoreHistory[]>([]);

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

        if (roomsData.status === 'fulfilled' && roomsData.value?.chatUsers) {
          const unread = roomsData.value.chatUsers.reduce(
            (sum, u) => sum + u.unreadCount, 0
          );
          setTotalUnread(unread);
        }

        if (scoresData.status === 'fulfilled' && scoresData.value.length > 0) {
          setLatestScore(scoresData.value[0]);
          setAllScores(scoresData.value);
        }
      } catch (err) {
        console.error('ダッシュボードデータ取得エラー:', err);
      }
    };
    fetchAll();
  }, []);

  const totalSessions = allScores.length;
  const averageScore = useMemo(() => {
    if (totalSessions === 0) return 0;
    return Math.round((allScores.reduce((sum, s) => sum + s.overallScore, 0) / totalSessions) * 10) / 10;
  }, [allScores, totalSessions]);
  const uniqueDays = useMemo(() => {
    return new Set(allScores.map(s => s.createdAt.split('T')[0])).size;
  }, [allScores]);

  return {
    stats,
    totalUnread,
    latestScore,
    allScores,
    totalSessions,
    averageScore,
    uniqueDays,
  };
}
