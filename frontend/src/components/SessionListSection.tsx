import ScoreDistributionCard from './ScoreDistributionCard';
import SessionTimeCard from './SessionTimeCard';
import SkillTrendChart from './SkillTrendChart';
import SkillMilestoneCard from './SkillMilestoneCard';
import PracticeCalendar from './PracticeCalendar';
import SessionDetailModal from './SessionDetailModal';
import ScoreHistorySessionCard from './ScoreHistorySessionCard';
import ScoreFilterSummary from './ScoreFilterSummary';
import ScoreTrendIndicator from './ScoreTrendIndicator';
import FilterTabs from './FilterTabs';
import { FILTERS, PERIOD_FILTERS } from '../hooks/useScoreHistory';
import type { ScoreHistoryItem } from '../types';
import type { ScoreHistoryItemWithDelta } from '../hooks/useScoreHistory';

interface SessionListSectionProps {
  history: ScoreHistoryItem[];
  filteredHistoryWithDelta: ScoreHistoryItemWithDelta[];
  filter: string;
  setFilter: (filter: string) => void;
  periodFilter: string;
  setPeriodFilter: (period: string) => void;
  selectedSession: ScoreHistoryItem | null;
  setSelectedSession: (session: ScoreHistoryItem | null) => void;
}

export default function SessionListSection({
  history,
  filteredHistoryWithDelta,
  filter,
  setFilter,
  periodFilter,
  setPeriodFilter,
  selectedSession,
  setSelectedSession,
}: SessionListSectionProps) {
  const latestSession = history[history.length - 1];
  const hasScores = latestSession && latestSession.scores.length > 0;

  return (
    <>
      <ScoreDistributionCard scores={history.map(h => h.overallScore)} />

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

      {hasScores && <SkillMilestoneCard scores={latestSession.scores} />}

      <SkillTrendChart history={history} />

      <SessionTimeCard dates={history.map(h => h.createdAt)} />

      <PracticeCalendar practiceDates={history.map(h => h.createdAt)} />

      <FilterTabs tabs={[...PERIOD_FILTERS]} selected={periodFilter} onSelect={setPeriodFilter} />

      <FilterTabs tabs={[...FILTERS]} selected={filter} onSelect={setFilter} />

      <ScoreFilterSummary scores={filteredHistoryWithDelta.map(h => h.overallScore)} />

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
    </>
  );
}
