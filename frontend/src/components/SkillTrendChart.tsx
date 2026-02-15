interface AxisScore {
  axis: string;
  score: number;
}

interface HistoryItem {
  sessionId: number;
  scores: AxisScore[];
}

interface SkillTrendChartProps {
  history: HistoryItem[];
}

export default function SkillTrendChart({ history }: SkillTrendChartProps) {
  if (history.length === 0) return null;

  const latest = history[history.length - 1];
  const previous = history.length > 1 ? history[history.length - 2] : null;
  const latestScores = Array.isArray(latest.scores) ? latest.scores : [];
  const previousScores = Array.isArray(previous?.scores) ? previous.scores : [];

  if (latestScores.length === 0) return null;

  const skills = latestScores.map((s) => {
    const prevScore = previousScores.find((p) => p.axis === s.axis)?.score;
    const delta = prevScore !== undefined ? s.score - prevScore : null;

    return {
      axis: s.axis,
      score: s.score,
      delta,
    };
  });

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-3">スキル別推移</p>
      <div className="space-y-2">
        {skills.map((skill) => (
          <div key={skill.axis} className="flex items-center gap-2">
            <span className="text-xs text-[var(--color-text-muted)] w-24 flex-shrink-0 truncate">
              {skill.axis}
            </span>
            <div className="flex-1 bg-surface-3 rounded-full h-2">
              <div
                className="h-2 rounded-full bg-primary-500 transition-all"
                style={{ width: `${skill.score * 10}%` }}
              />
            </div>
            <span
              data-testid="skill-latest-score"
              className="text-xs font-medium text-[var(--color-text-secondary)] w-5 text-right"
            >
              {skill.score}
            </span>
            {skill.delta !== null && skill.delta !== 0 && (
              <span
                className={`text-xs font-medium w-6 text-right ${
                  skill.delta > 0 ? 'text-emerald-400' : 'text-rose-400'
                }`}
              >
                {skill.delta > 0 ? `+${skill.delta}` : skill.delta}
              </span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
