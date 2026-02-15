import type { AxisScore } from '../types';
import Card from './Card';
import CardHeading from './CardHeading';

interface SkillGapAnalysisCardProps {
  scores: AxisScore[];
  goal: number;
}

export default function SkillGapAnalysisCard({ scores, goal }: SkillGapAnalysisCardProps) {
  if (scores.length === 0) return null;

  const gaps = scores
    .map((s) => ({
      axis: s.axis,
      score: s.score,
      gap: Math.round((goal - s.score) * 10) / 10,
      achieved: s.score >= goal,
    }))
    .sort((a, b) => b.gap - a.gap);

  const allAchieved = gaps.every((g) => g.achieved);

  return (
    <Card>
      <CardHeading>スキルギャップ分析</CardHeading>

      {allAchieved && (
        <p className="text-xs text-emerald-400 font-medium mb-3">
          全スキル目標達成！次の目標を設定しましょう！
        </p>
      )}

      <div className="space-y-2">
        {gaps.map((item) => (
          <div key={item.axis} className="flex items-center gap-2">
            <span data-testid="gap-axis" className="text-xs text-[var(--color-text-muted)] w-24 flex-shrink-0 truncate">
              {item.axis}
            </span>
            <div className="flex-1 bg-surface-3 rounded-full h-3">
              <div
                className={`h-3 rounded-full transition-all ${
                  item.achieved ? 'bg-emerald-400' : 'bg-rose-400'
                }`}
                style={{ width: `${Math.min((item.score / goal) * 100, 100)}%` }}
              />
            </div>
            {item.achieved ? (
              <span data-testid="gap-achieved" className="text-xs font-medium text-emerald-400 w-10 text-right">
                達成
              </span>
            ) : (
              <span data-testid="gap-value" className="text-xs font-medium text-rose-400 w-10 text-right">
                -{item.gap.toFixed(1)}
              </span>
            )}
          </div>
        ))}
      </div>
    </Card>
  );
}
