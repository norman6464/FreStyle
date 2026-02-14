import { useNavigate } from 'react-router-dom';
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
import SessionDetailModal from '../components/SessionDetailModal';
import { useScoreHistory, FILTERS } from '../hooks/useScoreHistory';

const AXIS_ADVICE: Record<string, string> = {
  '論理的構成力': '論理的構成力を伸ばすシナリオで練習しましょう',
  '配慮表現': '配慮表現を伸ばすシナリオで練習しましょう',
  '要約力': '要約力を伸ばすシナリオで練習しましょう',
  '提案力': '提案力を伸ばすシナリオで練習しましょう',
  '質問・傾聴力': '質問・傾聴力を伸ばすシナリオで練習しましょう',
};

export default function ScoreHistoryPage() {
  const { history, filteredHistory, filter, setFilter, loading, latestSession, weakestAxis, selectedSession, setSelectedSession } = useScoreHistory();
  const navigate = useNavigate();

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto p-4 space-y-4">
        <div className="text-sm text-[#888888]">スコア履歴を読み込み中...</div>
        <SkeletonCard />
        <SkeletonCard />
        <SkeletonCard />
      </div>
    );
  }

  if (history.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-[#888888]">
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

      {/* 週間比較 */}
      <WeeklyComparisonCard sessions={history} />

      {/* 最新スコアとランクバッジ */}
      {latestSession && (
        <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 flex items-center justify-between">
          <div>
            <p className="text-xs text-[#888888] mb-1">最新スコア</p>
            <p className="text-2xl font-bold text-[#F0F0F0]">{latestSession.overallScore.toFixed(1)}</p>
          </div>
          <ScoreRankBadge score={latestSession.overallScore} />
        </div>
      )}

      {/* 弱点ベースのおすすめ練習 */}
      {weakestAxis && (
        <div className="bg-surface-2 rounded-lg border border-[#444444] p-4">
          <p className="text-xs font-semibold text-primary-300 mb-1">おすすめ練習</p>
          <p className="text-xs text-primary-400 mb-2">
            {AXIS_ADVICE[weakestAxis.axis] || `${weakestAxis.axis}を伸ばすシナリオで練習しましょう`}
          </p>
          <button
            onClick={() => navigate('/practice')}
            className="text-xs font-medium text-primary-300 bg-surface-1 px-3 py-1.5 rounded-lg border border-[#444444] hover:bg-surface-3 transition-colors"
          >
            練習一覧を見る
          </button>
        </div>
      )}

      {/* スキルレーダーチャート */}
      {latestSession && latestSession.scores.length > 0 && (
        <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 flex justify-center">
          <SkillRadarChart scores={latestSession.scores} title="最新のスキルバランス" />
        </div>
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
        <p className="text-xs font-medium text-[#D0D0D0] mb-3">スコア推移</p>
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
                : 'border-transparent text-[#888888] hover:text-[#D0D0D0]'
            }`}
          >
            {f}
          </button>
        ))}
      </div>

      <p className="text-xs text-[#888888]">
        全 {filteredHistory.length} 件のフィードバック履歴
      </p>

      {filteredHistory.map((item) => {
        const originalIndex = history.indexOf(item);
        const prevItem = originalIndex > 0 ? history[originalIndex - 1] : null;
        const delta = prevItem
          ? Math.round((item.overallScore - prevItem.overallScore) * 10) / 10
          : null;

        return (
          <div
            key={item.sessionId}
            className="bg-surface-1 rounded-lg border border-surface-3 p-4 cursor-pointer hover:border-[#444444] transition-colors"
            onClick={() => setSelectedSession(item)}
          >
            <div className="flex items-center justify-between mb-3">
              <div>
                <h3 className="text-sm font-medium text-[#F0F0F0]">
                  {item.sessionTitle || `セッション #${item.sessionId}`}
                </h3>
                <p className="text-xs text-[#666666] mt-0.5">
                  {new Date(item.createdAt).toLocaleDateString('ja-JP', {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                  })}
                </p>
              </div>
              <div className="flex items-center gap-2">
                {delta !== null && delta !== 0 && (
                  <span className={`text-xs font-medium ${delta > 0 ? 'text-emerald-400' : 'text-rose-400'}`}>
                    {delta > 0 ? `+${delta.toFixed(1)}` : `\u2212${Math.abs(delta).toFixed(1)}`}
                  </span>
                )}
                <div className="flex items-center gap-1">
                  <span className="text-xs text-[#888888]">総合</span>
                  <span className="text-lg font-semibold text-[#F0F0F0]">
                    {item.overallScore.toFixed(1)}
                  </span>
                  <span className="text-xs text-[#666666]">/10</span>
                </div>
              </div>
            </div>

            <div className="space-y-1.5">
              {item.scores.map((axisScore) => (
                <div key={axisScore.axis} className="flex items-center gap-2">
                  <span className="text-xs text-[#888888] w-24 flex-shrink-0 truncate">
                    {axisScore.axis}
                  </span>
                  <div className="flex-1 bg-surface-3 rounded-full h-1.5">
                    <div
                      className="h-1.5 rounded-full bg-primary-500"
                      style={{ width: `${axisScore.score * 10}%` }}
                    />
                  </div>
                  <span className="text-xs text-[#A0A0A0] w-5 text-right">
                    {axisScore.score}
                  </span>
                </div>
              ))}
            </div>
          </div>
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
