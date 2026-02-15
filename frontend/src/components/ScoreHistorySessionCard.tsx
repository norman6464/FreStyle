import type { ScoreHistoryItem } from '../types';
import Card from './Card';
import AxisScoreBar from './AxisScoreBar';

interface ScoreHistorySessionCardProps {
  item: ScoreHistoryItem;
  delta: number | null;
  onClick: () => void;
}

export default function ScoreHistorySessionCard({ item, delta, onClick }: ScoreHistorySessionCardProps) {
  return (
    <Card
      className="cursor-pointer hover:border-[var(--color-border-hover)] transition-colors"
      onClick={onClick}
    >
      <div className="flex items-center justify-between mb-3">
        <div>
          <h3 className="text-sm font-medium text-[var(--color-text-primary)]">
            {item.sessionTitle || `セッション #${item.sessionId}`}
          </h3>
          <p className="text-xs text-[var(--color-text-faint)] mt-0.5">
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
            <span className="text-xs text-[var(--color-text-muted)]">総合</span>
            <span className="text-lg font-semibold text-[var(--color-text-primary)]">
              {item.overallScore.toFixed(1)}
            </span>
            <span className="text-xs text-[var(--color-text-faint)]">/10</span>
          </div>
        </div>
      </div>

      <div className="space-y-1.5">
        {item.scores.map((axisScore) => (
          <div key={axisScore.axis} className="flex items-center gap-2">
            <span className="text-xs text-[var(--color-text-muted)] w-24 flex-shrink-0 truncate">
              {axisScore.axis}
            </span>
            <div className="flex-1">
              <AxisScoreBar score={axisScore.score} />
            </div>
            <span className="text-xs text-[var(--color-text-tertiary)] w-5 text-right">
              {axisScore.score}
            </span>
          </div>
        ))}
      </div>
    </Card>
  );
}
