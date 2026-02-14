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
  if (delta > 0) return 'text-emerald-600';
  if (delta < 0) return 'text-rose-600';
  return 'text-slate-400';
}

export default function ScoreComparisonCard({
  firstScores,
  latestScores,
  firstOverall,
  latestOverall,
}: ScoreComparisonCardProps) {
  const overallDelta = Math.round((latestOverall - firstOverall) * 10) / 10;

  return (
    <div className="bg-white rounded-lg border border-slate-200 p-4">
      <p className="text-xs font-medium text-slate-700 mb-3">成長の記録</p>

      <div className="flex items-center justify-between mb-4">
        <div className="text-center">
          <p className="text-[10px] text-slate-400 mb-1">初回</p>
          <p className="text-xl font-bold text-slate-600">{firstOverall.toFixed(1)}</p>
        </div>
        <div className="text-center">
          <span className={`text-sm font-bold ${deltaColor(overallDelta)}`}>
            {formatDelta(overallDelta)}
          </span>
        </div>
        <div className="text-center">
          <p className="text-[10px] text-slate-400 mb-1">最新</p>
          <p className="text-xl font-bold text-primary-600">{latestOverall.toFixed(1)}</p>
        </div>
      </div>

      <div className="space-y-2">
        {latestScores.map((latest) => {
          const first = firstScores.find((f) => f.axis === latest.axis);
          const delta = first ? Math.round((latest.score - first.score) * 10) / 10 : 0;
          return (
            <div key={latest.axis} className="flex items-center justify-between">
              <span className="text-xs text-slate-500 w-24 truncate">{latest.axis}</span>
              <div className="flex items-center gap-3">
                <span className="text-xs text-slate-400 w-6 text-right">{first?.score ?? '-'}</span>
                <span className="text-xs text-slate-300">→</span>
                <span className="text-xs text-slate-700 w-6">{latest.score}</span>
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
