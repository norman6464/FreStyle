import { useNavigate } from 'react-router-dom';
import {
  UserGroupIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  AcademicCapIcon,
  ChartBarIcon,
} from '@heroicons/react/24/outline';
import AchievementBadgeCard from '../components/AchievementBadgeCard';
import StreakCalendarCard from '../components/StreakCalendarCard';
import CommunicationTipCard from '../components/CommunicationTipCard';
import DailyChallengeCard from '../components/DailyChallengeCard';
import MotivationQuoteCard from '../components/MotivationQuoteCard';
import DailyGoalCard from '../components/DailyGoalCard';
import LearningInsightsCard from '../components/LearningInsightsCard';
import NextStepCard from '../components/NextStepCard';
import PracticeLevelCard from '../components/PracticeLevelCard';
import PracticeReminderCard from '../components/PracticeReminderCard';
import RecentNotesCard from '../components/RecentNotesCard';
import RecentSessionsCard from '../components/RecentSessionsCard';
import WeeklyGoalProgressCard from '../components/WeeklyGoalProgressCard';
import WeeklyReportCard from '../components/WeeklyReportCard';
import LearningPatternCard from '../components/LearningPatternCard';
import ScoreGrowthTrendCard from '../components/ScoreGrowthTrendCard';
import PracticeFrequencyCard from '../components/PracticeFrequencyCard';
import { useMenuData } from '../hooks/useMenuData';

export default function MenuPage() {
  const navigate = useNavigate();
  const { stats, totalUnread, latestScore, allScores, totalSessions, averageScore, uniqueDays, practiceDates, sessionsThisWeek } = useMenuData();

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

      {/* 成長トレンド */}
      {allScores.length >= 2 && (
        <div className="mb-6">
          <ScoreGrowthTrendCard scores={allScores.map(s => s.overallScore)} />
        </div>
      )}

      {/* 次のステップ提案 */}
      <div className="mb-6">
        <NextStepCard totalSessions={totalSessions} averageScore={averageScore} />
      </div>

      {/* 練習レベル */}
      <div className="mb-6">
        <PracticeLevelCard totalSessions={totalSessions} />
      </div>

      {/* 達成バッジ */}
      <div className="mb-6">
        <AchievementBadgeCard totalSessions={totalSessions} />
      </div>

      {/* 練習頻度 */}
      {practiceDates.length > 0 && (
        <div className="mb-6">
          <PracticeFrequencyCard dates={practiceDates} />
        </div>
      )}

      {/* 練習ストリークカレンダー */}
      <div className="mb-6">
        <StreakCalendarCard practiceDates={practiceDates} />
      </div>

      {/* サマリー */}
      <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 mb-6">
        <div className="flex items-center gap-3">
          <UserGroupIcon className="w-5 h-5 text-[#666666]" />
          <div>
            <p className="text-xs text-[#888888]">会話した人数</p>
            <p className="text-lg font-semibold text-[#F0F0F0]">
              {stats?.chatPartnerCount ?? '—'}
              <span className="text-sm font-normal text-[#888888] ml-1">人</span>
            </p>
          </div>
        </div>
      </div>

      {/* おすすめアクション */}
      {showRecommendation && (
        <div className="bg-surface-2 rounded-lg border border-[#444444] p-4 mb-6">
          <p className="text-xs font-medium text-primary-300 mb-1">はじめての方へ</p>
          <p className="text-xs text-primary-400">
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

      {/* 学習パターン分析 */}
      {practiceDates.length > 0 && (
        <div className="mb-6">
          <LearningPatternCard practiceDates={practiceDates} />
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

      {/* 最近のメモ */}
      <div className="mb-6">
        <RecentNotesCard />
      </div>

      {/* 本日のチャレンジ */}
      <div className="mb-6">
        <DailyChallengeCard />
      </div>

      {/* 今日の一言 */}
      <div className="mb-6">
        <MotivationQuoteCard />
      </div>

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
            className="w-full flex items-center gap-4 bg-surface-1 rounded-lg border border-surface-3 p-4 text-left hover:bg-surface-2 transition-colors"
          >
            <item.icon className="w-5 h-5 text-[#888888] flex-shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <p className="text-sm font-medium text-[#F0F0F0]">{item.label}</p>
                {item.badge && (
                  <span className="text-[10px] font-medium bg-primary-500 text-white px-1.5 py-0.5 rounded-full">
                    {item.badge}
                  </span>
                )}
              </div>
              <p className="text-xs text-[#888888]">{item.description}</p>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
