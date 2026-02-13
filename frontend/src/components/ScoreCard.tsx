import type { ScoreCard as ScoreCardType } from '../types';

interface ScoreCardProps {
  scoreCard: ScoreCardType;
}

function getOverallLevel(score: number): { label: string; color: string } {
  if (score >= 8) return { label: '優秀レベル', color: 'bg-emerald-50 text-emerald-700 border-emerald-200' };
  if (score >= 5) return { label: '実務レベル', color: 'bg-amber-50 text-amber-700 border-amber-200' };
  return { label: '基礎レベル', color: 'bg-rose-50 text-rose-700 border-rose-200' };
}

function getBarColor(score: number): string {
  if (score >= 8) return 'bg-emerald-500';
  if (score >= 6) return 'bg-amber-500';
  return 'bg-rose-500';
}

export default function ScoreCard({ scoreCard }: ScoreCardProps) {
  const level = getOverallLevel(scoreCard.overallScore);

  return (
    <div className="bg-white rounded-lg border border-slate-200 p-4 my-3 max-w-[85%] self-start">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-medium text-slate-700">スコアカード</h3>
        <div className="flex items-center gap-2">
          <span className={`text-[10px] font-medium px-1.5 py-0.5 rounded border ${level.color}`}>
            {level.label}
          </span>
          <div className="flex items-center gap-1">
            <span className="text-xs text-slate-500">総合</span>
            <span className="text-lg font-semibold text-slate-800">
              {scoreCard.overallScore.toFixed(1)}
            </span>
            <span className="text-xs text-slate-400">/10</span>
          </div>
        </div>
      </div>

      <div className="space-y-2">
        {scoreCard.scores.map((axisScore) => (
          <div key={axisScore.axis}>
            <div className="flex items-center justify-between mb-0.5">
              <span className="text-xs text-slate-500">{axisScore.axis}</span>
              <span className="text-xs font-medium text-slate-600">{axisScore.score}</span>
            </div>
            <div className="w-full bg-slate-100 rounded-full h-1.5">
              <div
                className={`h-1.5 rounded-full ${getBarColor(axisScore.score)}`}
                style={{ width: `${axisScore.score * 10}%` }}
              />
            </div>
            <p className="text-[10px] text-slate-400 mt-0.5">{axisScore.comment}</p>
            {axisScore.score <= 5 && (
              <p className="text-[10px] text-amber-600 mt-0.5">
                💡 この項目を重点的に練習しましょう
              </p>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
