import type { ScoreCard as ScoreCardType } from '../types';

interface ScoreCardProps {
  scoreCard: ScoreCardType;
}

function getOverallLevel(score: number): { label: string; color: string } {
  if (score >= 8) return { label: '優秀レベル', color: 'bg-emerald-900/30 text-emerald-400 border-emerald-800' };
  if (score >= 5) return { label: '実務レベル', color: 'bg-amber-900/30 text-amber-400 border-amber-800' };
  return { label: '基礎レベル', color: 'bg-rose-900/30 text-rose-400 border-rose-800' };
}

function getBarColor(score: number): string {
  if (score >= 8) return 'bg-emerald-900/300';
  if (score >= 6) return 'bg-amber-900/300';
  return 'bg-rose-900/300';
}

export default function ScoreCard({ scoreCard }: ScoreCardProps) {
  const level = getOverallLevel(scoreCard.overallScore);

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 my-3 max-w-[85%] self-start">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-medium text-[#D0D0D0]">スコアカード</h3>
        <div className="flex items-center gap-2">
          <span className={`text-[10px] font-medium px-1.5 py-0.5 rounded border ${level.color}`}>
            {level.label}
          </span>
          <div className="flex items-center gap-1">
            <span className="text-xs text-[#888888]">総合</span>
            <span className="text-lg font-semibold text-[#F0F0F0]">
              {scoreCard.overallScore.toFixed(1)}
            </span>
            <span className="text-xs text-[#666666]">/10</span>
          </div>
        </div>
      </div>

      <div className="space-y-2">
        {scoreCard.scores.map((axisScore) => (
          <div key={axisScore.axis}>
            <div className="flex items-center justify-between mb-0.5">
              <span className="text-xs text-[#888888]">{axisScore.axis}</span>
              <span className="text-xs font-medium text-[#A0A0A0]">{axisScore.score}</span>
            </div>
            <div className="w-full bg-surface-3 rounded-full h-1.5">
              <div
                className={`h-1.5 rounded-full ${getBarColor(axisScore.score)}`}
                style={{ width: `${axisScore.score * 10}%` }}
              />
            </div>
            <p className="text-[10px] text-[#666666] mt-0.5">{axisScore.comment}</p>
            {axisScore.score <= 5 && (
              <p className="text-[10px] text-amber-400 mt-0.5">
                💡 この項目を重点的に練習しましょう
              </p>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
