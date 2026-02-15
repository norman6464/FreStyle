import type { ScoreCard as ScoreCardType } from '../types';
import { LightBulbIcon } from '@heroicons/react/24/outline';
import { getScoreLevel, getScoreBarColor } from '../utils/scoreColor';
import Card from './Card';

interface ScoreCardProps {
  scoreCard: ScoreCardType;
}

export default function ScoreCard({ scoreCard }: ScoreCardProps) {
  const level = getScoreLevel(scoreCard.overallScore);

  return (
    <Card className="my-3 max-w-[85%] self-start">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-medium text-[var(--color-text-secondary)]">スコアカード</h3>
        <div className="flex items-center gap-2">
          <span className={`text-[10px] font-medium px-1.5 py-0.5 rounded border ${level.color}`}>
            {level.label}
          </span>
          <div className="flex items-center gap-1">
            <span className="text-xs text-[var(--color-text-muted)]">総合</span>
            <span className="text-lg font-semibold text-[var(--color-text-primary)]">
              {scoreCard.overallScore.toFixed(1)}
            </span>
            <span className="text-xs text-[var(--color-text-faint)]">/10</span>
          </div>
        </div>
      </div>

      <div className="space-y-2">
        {(Array.isArray(scoreCard.scores) ? scoreCard.scores : []).map((axisScore) => (
          <div key={axisScore.axis}>
            <div className="flex items-center justify-between mb-0.5">
              <span className="text-xs text-[var(--color-text-muted)]">{axisScore.axis}</span>
              <span className="text-xs font-medium text-[var(--color-text-tertiary)]">{axisScore.score}</span>
            </div>
            <div className="w-full bg-surface-3 rounded-full h-1.5">
              <div
                className={`h-1.5 rounded-full ${getScoreBarColor(axisScore.score)}`}
                style={{ width: `${axisScore.score * 10}%` }}
              />
            </div>
            <p className="text-[10px] text-[var(--color-text-faint)] mt-0.5">{axisScore.comment}</p>
            {axisScore.score <= 5 && (
              <p className="text-[10px] text-amber-400 mt-0.5 flex items-center gap-0.5">
                <LightBulbIcon className="w-3 h-3 inline-block" />
                この項目を重点的に練習しましょう
              </p>
            )}
          </div>
        ))}
      </div>
    </Card>
  );
}
