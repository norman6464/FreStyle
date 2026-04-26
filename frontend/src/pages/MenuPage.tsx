import { UserGroupIcon, HomeIcon, ChartBarIcon } from '@heroicons/react/24/outline';
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
import QuickStartButton from '../components/QuickStartButton';
import Loading from '../components/Loading';
import { useMemo } from 'react';
import { useMenuData } from '../hooks/useMenuData';
import { useScoreHistory } from '../hooks/useScoreHistory';
import { useRecommendedScenario } from '../hooks/useRecommendedScenario';
import { PageIntro, FirstTimeWelcome, GlossaryTerm } from '../components/ui';
import { GLOSSARY } from '../constants/glossary';

/**
 * 初心者向けに上から順に並ぶホーム画面。
 *
 * 情報設計:
 *   1. ウェルカム（初回のみ） / クイックスタート
 *   2. 今日やること（デイリーゴール・チャレンジ）
 *   3. 成長サマリー（インサイト・スコア推移）
 *   4. 継続を促すエンゲージメント（バッジ・ストリーク・マイルストーン）
 *   5. 振り返り（週次レポート・パターン）
 *   6. 参考情報（ヒント・名言）
 */
export default function MenuPage() {
  const {
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
  } = useMenuData();
  const { weakestAxis } = useScoreHistory();
  const { scenario: recommendedScenario } = useRecommendedScenario(weakestAxis);
  // Hooks は条件付きの early return より前に呼ぶ（Rules of Hooks）。
  // 以前は loading 早期 return 直後に useMemo を置いており、再レンダー時に
  // フック呼び出し数が変わって React error #310 を引き起こしていた。
  const overallScores = useMemo(() => allScores.map((s) => s.overallScore), [allScores]);

  if (loading) {
    return <Loading message="読み込み中..." className="min-h-[calc(100vh-3.5rem)]" />;
  }

  const isFirstTimeUser = totalSessions === 0;

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      <PageIntro
        icon={<HomeIcon className="h-6 w-6" />}
        title="ホーム"
        description={
          <>
            今日の{' '}
            <GlossaryTerm
              term={GLOSSARY.practiceMode.term}
              definition={GLOSSARY.practiceMode.definition}
            />{' '}
            と学習の振り返りがここにまとまります。
          </>
        }
      />

      {/* 初回ユーザー向けウェルカム */}
      {isFirstTimeUser && (
        <FirstTimeWelcome
          storageKey="welcome:menu:v1"
          steps={[
            {
              title: 'シナリオを選んで AI と練習',
              description:
                '「練習モード」には 12 件のビジネスシーンがあります。まずは 1 つ選んで会話してみましょう。',
            },
            {
              title: '5 軸評価で自分の強みを知る',
              description:
                '会話が終わると AI が自動でスコアカードを作成。論理的構成力や配慮表現などを可視化します。',
            },
            {
              title: '毎日少しずつ続ける',
              description:
                '日次目標・週次目標・ストリークカレンダーで継続を応援します。1 日 1 回が理想的です。',
            },
          ]}
        />
      )}

      {/* クイックスタート */}
      <section aria-label="今すぐ始める">
        <QuickStartButton scenario={recommendedScenario} />
      </section>

      {/* 学習インサイト */}
      {totalSessions > 0 && (
        <section aria-labelledby="menu-insights">
          <h2 id="menu-insights" className="sr-only">
            学習の進捗サマリー
          </h2>
          <div className="space-y-6">
            <LearningInsightsCard
              totalSessions={totalSessions}
              averageScore={averageScore}
              streakDays={uniqueDays}
            />

            {overallScores.length >= 2 && <ScoreSparkline scores={overallScores} />}
            {overallScores.length >= 2 && <ScoreGrowthTrendCard scores={overallScores} />}
          </div>
        </section>
      )}

      {/* 弱点シナリオ推薦 */}
      {recommendedScenario && weakestAxis && (
        <RecommendedScenarioCard scenario={recommendedScenario} weakAxis={weakestAxis.axis} />
      )}

      {/* 次のステップ */}
      <section aria-labelledby="menu-nextstep">
        <h2 id="menu-nextstep" className="sr-only">
          次にやるといいこと
        </h2>
        <div className="space-y-6">
          <NextStepCard totalSessions={totalSessions} averageScore={averageScore} />
          <PracticeLevelCard totalSessions={totalSessions} />
          <AchievementBadgeCard totalSessions={totalSessions} />
          <SessionCountMilestoneCard sessionCount={totalSessions} />
        </div>
      </section>

      {/* 練習頻度・継続 */}
      {(practiceDates.length > 0 || latestScore) && (
        <section aria-labelledby="menu-engagement">
          <h2 id="menu-engagement" className="sr-only">
            継続とエンゲージメント
          </h2>
          <div className="space-y-6">
            {practiceDates.length > 0 && <PracticeFrequencyCard dates={practiceDates} />}
            <StreakCalendarCard practiceDates={practiceDates} />
            {latestScore && <PracticeReminderCard lastPracticeDate={latestScore.createdAt} />}
          </div>
        </section>
      )}

      {/* 会話した人数サマリー */}
      <section className="rounded-lg border border-surface-3 bg-surface-1 p-4" aria-label="会話した人数">
        <div className="flex items-center gap-3">
          <UserGroupIcon className="w-5 h-5 text-[var(--color-text-faint)]" aria-hidden="true" />
          <div>
            <p className="text-xs text-[var(--color-text-muted)]">会話した人数</p>
            <p className="text-lg font-semibold text-[var(--color-text-primary)]">
              {stats?.chatPartnerCount ?? '—'}
              <span className="text-sm font-normal text-[var(--color-text-muted)] ml-1">人</span>
            </p>
          </div>
        </div>
      </section>

      {/* 週次レポート / パターン分析 */}
      {totalSessions > 0 && (
        <section aria-labelledby="menu-reports">
          <h2
            id="menu-reports"
            className="mb-3 flex items-center gap-2 text-sm font-semibold text-[var(--color-text-secondary)]"
          >
            <ChartBarIcon className="h-4 w-4 text-primary-300" aria-hidden="true" />
            振り返り
          </h2>
          <div className="space-y-6">
            <WeeklyReportCard allScores={allScores} />
            {practiceDates.length > 0 && <LearningPatternCard practiceDates={practiceDates} />}
          </div>
        </section>
      )}

      {/* 目標 */}
      <section aria-labelledby="menu-goals">
        <h2 id="menu-goals" className="sr-only">
          日々の目標
        </h2>
        <div className="space-y-6">
          <WeeklyGoalProgressCard sessionsThisWeek={sessionsThisWeek} weeklyGoal={5} />
          <DailyGoalCard />
        </div>
      </section>

      {/* 直近のコンテンツ */}
      <section aria-labelledby="menu-recent">
        <h2 id="menu-recent" className="sr-only">
          最近のアクティビティ
        </h2>
        <div className="space-y-6">
          {allScores.length > 0 && <RecentSessionsCard sessions={allScores} />}
          <BookmarkedScenariosCard />
          <RecentNotesCard />
        </div>
      </section>

      {/* 参考情報 */}
      <section aria-labelledby="menu-tips">
        <h2 id="menu-tips" className="sr-only">
          毎日のヒント
        </h2>
        <div className="space-y-6">
          <DailyChallengeCard />
          <MotivationQuoteCard />
          <CommunicationTipCard />
        </div>
      </section>

      <MenuNavigationCard
        totalUnread={totalUnread}
        latestScore={latestScore ? latestScore.overallScore : null}
      />
    </div>
  );
}
