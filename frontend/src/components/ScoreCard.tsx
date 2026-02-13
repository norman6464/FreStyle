import type { ScoreCard as ScoreCardType } from '../types';

interface ScoreCardProps {
  scoreCard: ScoreCardType;
}

export default function ScoreCard({ scoreCard }: ScoreCardProps) {
  return (
    <div className="bg-white rounded-lg border border-slate-200 p-4 my-3 max-w-[85%] self-start">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-medium text-slate-700">スコアカード</h3>
        <div className="flex items-center gap-1">
          <span className="text-xs text-slate-500">総合</span>
          <span className="text-lg font-semibold text-slate-800">
            {scoreCard.overallScore.toFixed(1)}
          </span>
          <span className="text-xs text-slate-400">/10</span>
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
                className="h-1.5 rounded-full bg-primary-500"
                style={{ width: `${axisScore.score * 10}%` }}
              />
            </div>
            <p className="text-[10px] text-slate-400 mt-0.5">{axisScore.comment}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
