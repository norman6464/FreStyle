import { useNavigate } from 'react-router-dom';
import {
  UserGroupIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  AcademicCapIcon,
  ChartBarIcon,
} from '@heroicons/react/24/outline';
import CommunicationTipCard from '../components/CommunicationTipCard';
import DailyGoalCard from '../components/DailyGoalCard';
import LearningInsightsCard from '../components/LearningInsightsCard';
import PracticeReminderCard from '../components/PracticeReminderCard';
import RecentSessionsCard from '../components/RecentSessionsCard';
import WeeklyGoalProgressCard from '../components/WeeklyGoalProgressCard';
import WeeklyReportCard from '../components/WeeklyReportCard';
import { useMenuData } from '../hooks/useMenuData';

export default function MenuPage() {
  const navigate = useNavigate();
  const { stats, totalUnread, latestScore, allScores, totalSessions, averageScore, uniqueDays, sessionsThisWeek } = useMenuData();

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

      {/* 週間レポート */}
      {totalSessions > 0 && (
        <div className="mb-6">
          <WeeklyReportCard allScores={allScores} />
        </div>
      )}

      {/* 練習リマインダー */}
      {latestScore && (
        <div className="mb-6">
          <PracticeReminderCard lastPracticeDate={latestScore.createdAt} />
        </div>
      )}

      {/* 週間練習目標 */}
      <div className="mb-6">
        <WeeklyGoalProgressCard sessionsThisWeek={sessionsThisWeek} weeklyGoal={5} />
      </div>

      {/* 日次学習目標 */}
      <div className="mb-6">
        <DailyGoalCard />
      </div>

      {/* 直近のセッション */}
      {allScores.length > 0 && (
        <div className="mb-6">
          <RecentSessionsCard sessions={allScores} />
        </div>
      )}

      {/* コミュニケーションTips */}
      <div className="mb-6">
        <CommunicationTipCard />
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
