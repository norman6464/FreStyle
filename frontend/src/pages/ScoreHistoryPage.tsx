import { useNavigate } from 'react-router-dom';
import { SparklesIcon } from '@heroicons/react/24/outline';
import { SkeletonCard } from '../components/Skeleton';
import EmptyState from '../components/EmptyState';
import SkillRadarChart from '../components/SkillRadarChart';
import PracticeCalendar from '../components/PracticeCalendar';
import ScoreRankBadge from '../components/ScoreRankBadge';
import ScoreStatsSummary from '../components/ScoreStatsSummary';
import SkillMilestoneCard from '../components/SkillMilestoneCard';
import ScoreComparisonCard from '../components/ScoreComparisonCard';
import ScoreImprovementAdvice from '../components/ScoreImprovementAdvice';
import SkillSummaryCard from '../components/SkillSummaryCard';
import CommunicationStyleCard from '../components/CommunicationStyleCard';
import ScoreGoalCard from '../components/ScoreGoalCard';
import ScoreDistributionCard from '../components/ScoreDistributionCard';
import SessionTimeCard from '../components/SessionTimeCard';
import SkillGapAnalysisCard from '../components/SkillGapAnalysisCard';
import WeeklyComparisonCard from '../components/WeeklyComparisonCard';
import SkillTrendChart from '../components/SkillTrendChart';
import SkillRadarOverlayCard from '../components/SkillRadarOverlayCard';
import SessionDetailModal from '../components/SessionDetailModal';
import ScoreHistorySessionCard from '../components/ScoreHistorySessionCard';
import ScoreFilterSummary from '../components/ScoreFilterSummary';
import ScoreTrendIndicator from '../components/ScoreTrendIndicator';
import SessionFeedbackSummary from '../components/SessionFeedbackSummary';
import WeakAxisAdviceCard from '../components/WeakAxisAdviceCard';
import PersonalBestCard from '../components/PersonalBestCard';
import FilterTabs from '../components/FilterTabs';
import { useScoreHistory, FILTERS } from '../hooks/useScoreHistory';
import { useLocalStorage } from '../hooks/useLocalStorage';

export default function ScoreHistoryPage() {
  const navigate = useNavigate();
  const { history, filteredHistoryWithDelta, filter, setFilter, loading, latestSession, averageScore, weakestAxis, selectedSession, setSelectedSession } = useScoreHistory();
  const [scoreGoal] = useLocalStorage('scoreGoal', 8.0);

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto p-4 space-y-4">
        <div className="text-sm text-[var(--color-text-muted)]">スコア履歴を読み込み中...</div>
        <SkeletonCard />
        <SkeletonCard />
        <SkeletonCard />
      </div>
    );
  }

  if (history.length === 0) {
    return (
      <EmptyState
        icon={SparklesIcon}
        title="スコア履歴がありません"
        description="AIアシスタントでフィードバックを受けるとスコアが記録されます"
        action={{
          label: 'AIアシスタントで練習を開始',
          onClick: () => navigate('/chat/ask-ai'),
        }}
      />
    );
  }

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-3">
      {/* コミュニケーションスタイル */}
      <CommunicationStyleCard sessions={history} />

      {/* 目標スコア */}
      <ScoreGoalCard
        averageScore={averageScore}
      />

      {/* スキルギャップ分析 */}
      {latestSession && latestSession.scores.length > 0 && (
        <SkillGapAnalysisCard
          scores={latestSession.scores}
          goal={scoreGoal}
        />
      )}

      {/* 統計サマリー */}
      <ScoreStatsSummary history={history} />

      {/* 自己ベスト */}
      <PersonalBestCard history={history} />

      {/* 週間比較 */}
      <WeeklyComparisonCard sessions={history} />

      {/* 最新スコアとランクバッジ */}
      {latestSession && (
        <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 flex items-center justify-between">
          <div>
            <p className="text-xs text-[var(--color-text-muted)] mb-1">最新スコア</p>
            <p className="text-2xl font-bold text-[var(--color-text-primary)]">{latestSession.overallScore.toFixed(1)}</p>
          </div>
          <ScoreRankBadge score={latestSession.overallScore} />
        </div>
      )}

      {/* セッションフィードバックサマリー */}
      {latestSession && latestSession.scores.length > 0 && (
        <SessionFeedbackSummary scores={latestSession.scores} overallScore={latestSession.overallScore} />
      )}

      {/* 弱点ベースのおすすめ練習 */}
      {weakestAxis && <WeakAxisAdviceCard axis={weakestAxis.axis} />}

      {/* スキルレーダーチャート */}
      {latestSession && latestSession.scores.length > 0 && (
        <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 flex justify-center">
          <SkillRadarChart scores={latestSession.scores} title="最新のスキルバランス" />
        </div>
      )}

      {/* スキル変化レーダー（前回vs今回） */}
      {history.length >= 2 && latestSession && (
        <SkillRadarOverlayCard
          previousScores={history[history.length - 2].scores}
          currentScores={latestSession.scores}
        />
      )}

      {/* 初回vs最新スコア比較 */}
      {history.length >= 2 && latestSession && history[0].scores.length > 0 && (
        <ScoreComparisonCard
          firstScores={history[0].scores}
          latestScores={latestSession.scores}
          firstOverall={history[0].overallScore}
          latestOverall={latestSession.overallScore}
        />
      )}

      {/* スキル強弱サマリー */}
      {latestSession && latestSession.scores.length > 0 && (
        <SkillSummaryCard scores={latestSession.scores} />
      )}

      {/* 改善アドバイス */}
      {latestSession && latestSession.scores.length > 0 && (
        <ScoreImprovementAdvice scores={latestSession.scores} />
      )}

      {/* スコア分布 */}
      <ScoreDistributionCard scores={history.map(h => h.overallScore)} />

      {/* スコア推移グラフ */}
      <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
        <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-3">スコア推移</p>
        <div className="flex items-end gap-1 h-16">
          {history.map((item) => (
            <div
              key={item.sessionId}
              data-testid="trend-bar"
              className="flex-1 bg-primary-500 rounded-t"
              style={{ height: `${item.overallScore * 10}%` }}
              title={`${item.sessionTitle}: ${item.overallScore}`}
            />
          ))}
        </div>
      </div>

      {/* スキル別マイルストーン */}
      {latestSession && latestSession.scores.length > 0 && (
        <SkillMilestoneCard scores={latestSession.scores} />
      )}

      {/* スキル別推移 */}
      <SkillTrendChart history={history} />

      {/* セッション時間帯 */}
      <SessionTimeCard dates={history.map(h => h.createdAt)} />

      {/* 練習カレンダー */}
      <PracticeCalendar practiceDates={history.map(h => h.createdAt)} />

      {/* フィルタタブ */}
      <FilterTabs tabs={[...FILTERS]} selected={filter} onSelect={setFilter} />

      {/* フィルタサマリー */}
      <ScoreFilterSummary scores={filteredHistoryWithDelta.map(h => h.overallScore)} />

      {/* トレンドインジケーター */}
      <ScoreTrendIndicator scores={filteredHistoryWithDelta.map(h => h.overallScore)} />

      {filteredHistoryWithDelta.map((item) => (
        <ScoreHistorySessionCard
          key={item.sessionId}
          item={item}
          delta={item.delta}
          onClick={() => setSelectedSession(item)}
        />
      ))}

      {selectedSession && (
        <SessionDetailModal
          session={selectedSession}
          onClose={() => setSelectedSession(null)}
        />
      )}
    </div>
  );
}
