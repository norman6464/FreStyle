import SkillRadarChart from './SkillRadarChart';
import ScoreRankBadge from './ScoreRankBadge';
import ScoreStatsSummary from './ScoreStatsSummary';
import ScoreComparisonCard from './ScoreComparisonCard';
import ScoreImprovementAdvice from './ScoreImprovementAdvice';
import SkillSummaryCard from './SkillSummaryCard';
import CommunicationStyleCard from './CommunicationStyleCard';
import ScoreGoalCard from './ScoreGoalCard';
import SkillGapAnalysisCard from './SkillGapAnalysisCard';
import WeeklyComparisonCard from './WeeklyComparisonCard';
import SkillRadarOverlayCard from './SkillRadarOverlayCard';
import SessionFeedbackSummary from './SessionFeedbackSummary';
import WeakAxisAdviceCard from './WeakAxisAdviceCard';
import PersonalBestCard from './PersonalBestCard';
import type { ScoreHistoryItem, AxisScore } from '../types';

interface ScoreOverviewSectionProps {
  history: ScoreHistoryItem[];
  latestSession: ScoreHistoryItem;
  averageScore: number;
  weakestAxis: AxisScore | null;
  scoreGoal: number;
}

export default function ScoreOverviewSection({
  history,
  latestSession,
  averageScore,
  weakestAxis,
  scoreGoal,
}: ScoreOverviewSectionProps) {
  const hasScores = latestSession.scores.length > 0;

  return (
    <>
      <CommunicationStyleCard sessions={history} />

      <ScoreGoalCard averageScore={averageScore} />

      {hasScores && (
        <SkillGapAnalysisCard scores={latestSession.scores} goal={scoreGoal} />
      )}

      <ScoreStatsSummary history={history} />

      <PersonalBestCard history={history} />

      <WeeklyComparisonCard sessions={history} />

      <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 flex items-center justify-between">
        <div>
          <p className="text-xs text-[var(--color-text-muted)] mb-1">最新スコア</p>
          <p className="text-2xl font-bold text-[var(--color-text-primary)]">{latestSession.overallScore.toFixed(1)}</p>
        </div>
        <ScoreRankBadge score={latestSession.overallScore} />
      </div>

      {hasScores && (
        <SessionFeedbackSummary scores={latestSession.scores} overallScore={latestSession.overallScore} />
      )}

      {weakestAxis && <WeakAxisAdviceCard axis={weakestAxis.axis} />}

      {hasScores && (
        <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 flex justify-center">
          <SkillRadarChart scores={latestSession.scores} title="最新のスキルバランス" />
        </div>
      )}

      {history.length >= 2 && (
        <SkillRadarOverlayCard
          previousScores={history[history.length - 2].scores}
          currentScores={latestSession.scores}
        />
      )}

      {history.length >= 2 && history[0].scores.length > 0 && (
        <ScoreComparisonCard
          firstScores={history[0].scores}
          latestScores={latestSession.scores}
          firstOverall={history[0].overallScore}
          latestOverall={latestSession.overallScore}
        />
      )}

      {hasScores && <SkillSummaryCard scores={latestSession.scores} />}

      {hasScores && <ScoreImprovementAdvice scores={latestSession.scores} />}
    </>
  );
}
