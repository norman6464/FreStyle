import type { ScoreCard as ScoreCardType } from '../types';

interface ScoreCardProps {
  scoreCard: ScoreCardType;
}

export default function ScoreCard({ scoreCard }: ScoreCardProps) {
  const getScoreColor = (score: number) => {
    if (score >= 8) return 'bg-green-500';
    if (score >= 6) return 'bg-yellow-500';
    if (score >= 4) return 'bg-orange-500';
    return 'bg-red-500';
  };

  const getOverallColor = (score: number) => {
    if (score >= 8) return 'text-green-600';
    if (score >= 6) return 'text-yellow-600';
    if (score >= 4) return 'text-orange-600';
    return 'text-red-600';
  };

  return (
    <div className="bg-white rounded-xl border border-gray-200 p-4 my-3 max-w-[85%] self-start">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-bold text-gray-700">スコアカード</h3>
        <div className="flex items-center gap-1">
          <span className="text-xs text-gray-500">総合</span>
          <span className={`text-lg font-bold ${getOverallColor(scoreCard.overallScore)}`}>
            {scoreCard.overallScore.toFixed(1)}
          </span>
          <span className="text-xs text-gray-400">/10</span>
        </div>
      </div>

      <div className="space-y-2">
        {scoreCard.scores.map((axisScore) => (
          <div key={axisScore.axis}>
            <div className="flex items-center justify-between mb-0.5">
              <span className="text-xs text-gray-600">{axisScore.axis}</span>
              <span className="text-xs font-bold text-gray-700">{axisScore.score}</span>
            </div>
            <div className="w-full bg-gray-100 rounded-full h-2">
              <div
                className={`h-2 rounded-full transition-all duration-500 ${getScoreColor(axisScore.score)}`}
                style={{ width: `${axisScore.score * 10}%` }}
              />
            </div>
            <p className="text-[10px] text-gray-400 mt-0.5">{axisScore.comment}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
