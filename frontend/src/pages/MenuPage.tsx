import { useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import {
  UserGroupIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  AcademicCapIcon,
  ChartBarIcon,
} from '@heroicons/react/24/outline';
import DailyGoalCard from '../components/DailyGoalCard';
import LearningInsightsCard from '../components/LearningInsightsCard';

interface ChatStats {
  chatPartnerCount: number;
}

interface ScoreHistory {
  sessionId: number;
  sessionTitle: string;
  overallScore: number;
  createdAt: string;
}

export default function MenuPage() {
  const navigate = useNavigate();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const [stats, setStats] = useState<ChatStats | null>(null);
  const [totalUnread, setTotalUnread] = useState<number>(0);
  const [latestScore, setLatestScore] = useState<ScoreHistory | null>(null);
  const [allScores, setAllScores] = useState<ScoreHistory[]>([]);

  useEffect(() => {
    const fetchWithRetry = async (url: string) => {
      const res = await fetch(url, {
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
      });
      if (res.status === 401) {
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          { method: 'POST', credentials: 'include' }
        );
        if (!refreshRes.ok) {
          navigate('/login');
          return null;
        }
        await refreshRes.json();
        const retryRes = await fetch(url, {
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
        });
        if (!retryRes.ok) return null;
        return retryRes.json();
      }
      if (!res.ok) return null;
      return res.json();
    };

    const fetchAll = async () => {
      try {
        const [statsData, roomsData, scoresData] = await Promise.all([
          fetchWithRetry(`${API_BASE_URL}/api/chat/stats`),
          fetchWithRetry(`${API_BASE_URL}/api/chat/rooms`),
          fetchWithRetry(`${API_BASE_URL}/api/scores/history`),
        ]);

        if (statsData) setStats(statsData);

        if (roomsData?.chatUsers) {
          const unread = roomsData.chatUsers.reduce(
            (sum: number, u: { unreadCount: number }) => sum + u.unreadCount, 0
          );
          setTotalUnread(unread);
        }

        if (scoresData && scoresData.length > 0) {
          setLatestScore(scoresData[0]);
          setAllScores(scoresData);
        }
      } catch (err) {
        console.error('Error fetching dashboard data:', err);
      }
    };
    fetchAll();
  }, [navigate, API_BASE_URL]);

  const menuItems = [
    {
      icon: ChatBubbleLeftRightIcon,
      label: 'チャット',
      description: 'メンバーとメッセージをやり取り',
      to: '/chat',
      badge: totalUnread > 0 ? `${totalUnread}件の未読` : null,
    },
    {
      icon: SparklesIcon,
      label: 'AI アシスタント',
      description: 'AIにコミュニケーションを分析・フィードバック',
      to: '/chat/ask-ai',
      badge: null,
    },
    {
      icon: AcademicCapIcon,
      label: '練習モード',
      description: 'ビジネスシナリオでロールプレイ練習',
      to: '/practice',
      badge: null,
    },
    {
      icon: ChartBarIcon,
      label: 'スコア履歴',
      description: 'フィードバック結果の振り返り',
      to: '/scores',
      badge: latestScore ? `最新: ${latestScore.overallScore}` : null,
    },
  ];

  const showRecommendation = !latestScore && stats?.chatPartnerCount === 0;

  const totalSessions = allScores.length;
  const averageScore = totalSessions > 0
    ? Math.round((allScores.reduce((sum, s) => sum + s.overallScore, 0) / totalSessions) * 10) / 10
    : 0;
  const uniqueDays = new Set(allScores.map(s => s.createdAt.split('T')[0])).size;

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* 学習インサイト */}
      {totalSessions > 0 && (
        <div className="mb-6">
          <LearningInsightsCard
            totalSessions={totalSessions}
            averageScore={averageScore}
            streakDays={uniqueDays}
          />
        </div>
      )}

      {/* サマリー */}
      <div className="bg-white rounded-lg border border-slate-200 p-4 mb-6">
        <div className="flex items-center gap-3">
          <UserGroupIcon className="w-5 h-5 text-slate-400" />
          <div>
            <p className="text-xs text-slate-500">会話した人数</p>
            <p className="text-lg font-semibold text-slate-800">
              {stats?.chatPartnerCount ?? '—'}
              <span className="text-sm font-normal text-slate-500 ml-1">人</span>
            </p>
          </div>
        </div>
      </div>

      {/* おすすめアクション */}
      {showRecommendation && (
        <div className="bg-primary-50 rounded-lg border border-primary-200 p-4 mb-6">
          <p className="text-xs font-medium text-primary-700 mb-1">はじめての方へ</p>
          <p className="text-xs text-primary-600">
            まずは練習モードから始めてみましょう。AIが相手役を演じるビジネスシナリオで、コミュニケーションスキルを磨けます。
          </p>
        </div>
      )}

      {/* 日次学習目標 */}
      <div className="mb-6">
        <DailyGoalCard />
      </div>

      {/* メニュー */}
      <div className="space-y-2">
        {menuItems.map((item) => (
          <button
            key={item.to}
            onClick={() => navigate(item.to)}
            className="w-full flex items-center gap-4 bg-white rounded-lg border border-slate-200 p-4 text-left hover:bg-primary-50 transition-colors"
          >
            <item.icon className="w-5 h-5 text-slate-500 flex-shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <p className="text-sm font-medium text-slate-800">{item.label}</p>
                {item.badge && (
                  <span className="text-[10px] font-medium bg-primary-500 text-white px-1.5 py-0.5 rounded-full">
                    {item.badge}
                  </span>
                )}
              </div>
              <p className="text-xs text-slate-500">{item.description}</p>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
