import { SkeletonCard } from '../components/Skeleton';
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
import { useScoreHistory, FILTERS } from '../hooks/useScoreHistory';

export default function ScoreHistoryPage() {
  const { history, filteredHistory, filter, setFilter, loading, latestSession, weakestAxis, selectedSession, setSelectedSession } = useScoreHistory();

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
      <div className="flex flex-col items-center justify-center h-64 text-[var(--color-text-muted)]">
        <p className="text-sm font-medium">スコア履歴がありません</p>
        <p className="text-xs mt-1">AIアシスタントでフィードバックを受けるとスコアが記録されます</p>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-3">
      {/* コミュニケーションスタイル */}
      <CommunicationStyleCard sessions={history} />

      {/* 目標スコア */}
      <ScoreGoalCard
        averageScore={Math.round((history.reduce((sum, h) => sum + h.overallScore, 0) / history.length) * 10) / 10}
      />

      {/* スキルギャップ分析 */}
      {latestSession && latestSession.scores.length > 0 && (
        <SkillGapAnalysisCard
          scores={latestSession.scores}
          goal={(() => {
            try {
              const stored = localStorage.getItem('scoreGoal');
              return stored ? parseFloat(stored) : 8.0;
            } catch {
              return 8.0;
            }
          })()}
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
      <div className="flex gap-1 border-b border-surface-3">
        {FILTERS.map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-3 py-2 text-xs font-medium transition-colors border-b-2 -mb-px ${
              filter === f
                ? 'border-primary-500 text-primary-400'
                : 'border-transparent text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
            }`}
          >
            {f}
          </button>
        ))}
      </div>

      {/* フィルタサマリー */}
      <ScoreFilterSummary scores={filteredHistory.map(h => h.overallScore)} />

      {/* トレンドインジケーター */}
      <ScoreTrendIndicator scores={filteredHistory.map(h => h.overallScore)} />

      {filteredHistory.map((item) => {
        const originalIndex = history.indexOf(item);
        const prevItem = originalIndex > 0 ? history[originalIndex - 1] : null;
        const delta = prevItem
          ? Math.round((item.overallScore - prevItem.overallScore) * 10) / 10
          : null;

        return (
          <ScoreHistorySessionCard
            key={item.sessionId}
            item={item}
            delta={delta}
            onClick={() => setSelectedSession(item)}
          />
        );
      })}

      {selectedSession && (
        <SessionDetailModal
          session={selectedSession}
          onClose={() => setSelectedSession(null)}
        />
      )}
    </div>
  );
}
