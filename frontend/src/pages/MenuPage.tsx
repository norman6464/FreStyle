import { UserGroupIcon } from '@heroicons/react/24/outline';
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
import MenuNavigationCard from '../components/MenuNavigationCard';
import SessionCountMilestoneCard from '../components/SessionCountMilestoneCard';
import ScoreSparkline from '../components/ScoreSparkline';
import RecommendedScenarioCard from '../components/RecommendedScenarioCard';
import BookmarkedScenariosCard from '../components/BookmarkedScenariosCard';
import Loading from '../components/Loading';
import { useMenuData } from '../hooks/useMenuData';
import { useScoreHistory } from '../hooks/useScoreHistory';
import { useRecommendedScenario } from '../hooks/useRecommendedScenario';

export default function MenuPage() {
  const { stats, totalUnread, latestScore, allScores, totalSessions, averageScore, uniqueDays, practiceDates, sessionsThisWeek, loading } = useMenuData();
  const { weakestAxis } = useScoreHistory();
  const { scenario: recommendedScenario } = useRecommendedScenario(weakestAxis);

  if (loading) {
    return <Loading message="読み込み中..." className="min-h-[calc(100vh-3.5rem)]" />;
  }

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

      {/* スコアスパークライン */}
      {allScores.length >= 2 && (
        <div className="mb-6">
          <ScoreSparkline scores={allScores.map(s => s.overallScore)} />
        </div>
      )}

      {/* 成長トレンド */}
      {allScores.length >= 2 && (
        <div className="mb-6">
          <ScoreGrowthTrendCard scores={allScores.map(s => s.overallScore)} />
        </div>
      )}

      {/* 弱点シナリオ推薦 */}
      {recommendedScenario && weakestAxis && (
        <div className="mb-6">
          <RecommendedScenarioCard scenario={recommendedScenario} weakAxis={weakestAxis.axis} />
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

      {/* セッション数マイルストーン */}
      <div className="mb-6">
        <SessionCountMilestoneCard sessionCount={totalSessions} />
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
          <UserGroupIcon className="w-5 h-5 text-[var(--color-text-faint)]" />
          <div>
            <p className="text-xs text-[var(--color-text-muted)]">会話した人数</p>
            <p className="text-lg font-semibold text-[var(--color-text-primary)]">
              {stats?.chatPartnerCount ?? '—'}
              <span className="text-sm font-normal text-[var(--color-text-muted)] ml-1">人</span>
            </p>
          </div>
        </div>
      </div>

      {/* おすすめアクション */}
      {showRecommendation && (
        <div className="bg-surface-2 rounded-lg border border-[var(--color-border-hover)] p-4 mb-6">
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

      {/* ブックマーク済みシナリオ */}
      <BookmarkedScenariosCard />

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
      <MenuNavigationCard totalUnread={totalUnread} latestScore={latestScore ? latestScore.overallScore : null} />
    </div>
  );
}
