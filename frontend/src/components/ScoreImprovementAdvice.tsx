import type { AxisScore } from '../types';
import Card from './Card';
import CardHeading from './CardHeading';
import { IMPROVEMENT_ADVICE, SCORE_THRESHOLD } from '../constants/axisAdvice';

interface ScoreImprovementAdviceProps {
  scores: AxisScore[];
}

export default function ScoreImprovementAdvice({ scores }: ScoreImprovementAdviceProps) {
  const weakAxes = scores.filter((s) => s.score < SCORE_THRESHOLD);

  return (
    <Card>
      <CardHeading>改善アドバイス</CardHeading>

      {weakAxes.length === 0 ? (
        <p className="text-xs text-emerald-400 font-medium">
          素晴らしい成績です！この調子で続けましょう。
        </p>
      ) : (
        <div className="space-y-3">
          {weakAxes.map((axis) => (
            <div key={axis.axis} data-testid="advice-item" className="flex gap-3">
              <div className="flex-shrink-0 w-5 h-5 rounded-full bg-amber-900/30 flex items-center justify-center mt-0.5">
                <span className="text-amber-400 text-[10px] font-bold">!</span>
              </div>
              <div>
                <p className="text-xs font-medium text-[var(--color-text-primary)]">
                  {axis.axis}（{axis.score.toFixed(1)}）
                </p>
                <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                  {IMPROVEMENT_ADVICE[axis.axis] || `${axis.axis}を伸ばすための練習を続けましょう。`}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
    </Card>
  );
}
