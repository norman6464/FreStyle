import { useNavigate } from 'react-router-dom';
import { SparklesIcon } from '@heroicons/react/24/outline';
import { SkeletonCard } from '../components/Skeleton';
import EmptyState from '../components/EmptyState';
import ScoreOverviewSection from '../components/ScoreOverviewSection';
import SessionListSection from '../components/SessionListSection';
import { useScoreHistory } from '../hooks/useScoreHistory';
import { useScoreGoal } from '../hooks/useScoreGoal';

export default function ScoreHistoryPage() {
  const navigate = useNavigate();
  const { history, filteredHistoryWithDelta, filter, setFilter, loading, latestSession, averageScore, weakestAxis, selectedSession, setSelectedSession } = useScoreHistory();
  const { goal: scoreGoal } = useScoreGoal();

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
      <ScoreOverviewSection
        history={history}
        latestSession={latestSession!}
        averageScore={averageScore}
        weakestAxis={weakestAxis}
        scoreGoal={scoreGoal}
      />

      <SessionListSection
        history={history}
        filteredHistoryWithDelta={filteredHistoryWithDelta}
        filter={filter}
        setFilter={setFilter}
        selectedSession={selectedSession}
        setSelectedSession={setSelectedSession}
      />
    </div>
  );
}
