interface AxisScore {
  axis: string;
  score: number;
  comment: string;
}

interface ScoreComparisonCardProps {
  firstScores: AxisScore[];
  latestScores: AxisScore[];
  firstOverall: number;
  latestOverall: number;
}

function formatDelta(delta: number): string {
  if (delta === 0) return '±0';
  return delta > 0 ? `+${delta.toFixed(1)}` : `${delta.toFixed(1)}`;
}

function deltaColor(delta: number): string {
  if (delta > 0) return 'text-emerald-400';
  if (delta < 0) return 'text-rose-400';
  return 'text-[var(--color-text-faint)]';
}

export default function ScoreComparisonCard({
  firstScores,
  latestScores,
  firstOverall,
  latestOverall,
}: ScoreComparisonCardProps) {
  const overallDelta = Math.round((latestOverall - firstOverall) * 10) / 10;

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-3">成長の記録</p>

      <div className="flex items-center justify-between mb-4">
        <div className="text-center">
          <p className="text-[10px] text-[var(--color-text-faint)] mb-1">初回</p>
          <p className="text-xl font-bold text-[var(--color-text-tertiary)]">{firstOverall.toFixed(1)}</p>
        </div>
        <div className="text-center">
          <span className={`text-sm font-bold ${deltaColor(overallDelta)}`}>
            {formatDelta(overallDelta)}
          </span>
        </div>
        <div className="text-center">
          <p className="text-[10px] text-[var(--color-text-faint)] mb-1">最新</p>
          <p className="text-xl font-bold text-primary-400">{latestOverall.toFixed(1)}</p>
        </div>
      </div>

      <div className="space-y-2">
        {latestScores.map((latest) => {
          const first = firstScores.find((f) => f.axis === latest.axis);
          const delta = first ? Math.round((latest.score - first.score) * 10) / 10 : 0;
          return (
            <div key={latest.axis} className="flex items-center justify-between">
              <span className="text-xs text-[var(--color-text-muted)] w-24 truncate">{latest.axis}</span>
              <div className="flex items-center gap-3">
                <span className="text-xs text-[var(--color-text-faint)] w-6 text-right">{first?.score ?? '-'}</span>
                <span className="text-xs text-[var(--color-text-subtle)]">→</span>
                <span className="text-xs text-[var(--color-text-secondary)] w-6">{latest.score}</span>
                <span className={`text-xs font-medium w-8 text-right ${deltaColor(delta)}`}>
                  {formatDelta(delta)}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
