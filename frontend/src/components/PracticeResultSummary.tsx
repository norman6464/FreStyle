import { useNavigate } from 'react-router-dom';
import { getAdviceForAxis } from '../constants/axisAdvice';
import { useStartPracticeSession } from '../hooks/useStartPracticeSession';
import type { ScoreCard } from '../types';

interface PracticeResultSummaryProps {
  scoreCard: ScoreCard;
  scenarioName: string;
  scenarioId?: number;
}

export default function PracticeResultSummary({ scoreCard, scenarioName, scenarioId }: PracticeResultSummaryProps) {
  const navigate = useNavigate();
  const { startSession, starting } = useStartPracticeSession();

  const scores = Array.isArray(scoreCard.scores) ? scoreCard.scores : [];
  if (scores.length === 0) return null;

  const sorted = [...scores].sort((a, b) => b.score - a.score);
  const strongest = sorted[0];
  const weakest = sorted[sorted.length - 1];

  const advice = getAdviceForAxis(weakest.axis);

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 my-3 max-w-[85%] self-start">
      <h3 className="text-sm font-semibold text-[var(--color-text-primary)] mb-1">練習結果サマリー</h3>
      <p className="text-xs text-[var(--color-text-muted)] mb-3">「{scenarioName}」の練習結果</p>

      <div className="grid grid-cols-2 gap-3 mb-3">
        {/* 強み */}
        <div className="bg-emerald-900/30 rounded-lg p-3 border border-emerald-800">
          <p className="text-[10px] font-semibold text-emerald-400 mb-1">強み</p>
          <p className="text-sm font-medium text-emerald-800">{strongest.axis}</p>
          <p className="text-xs text-emerald-400 mt-0.5">スコア: {strongest.score}/10</p>
        </div>

        {/* 課題 */}
        <div className="bg-rose-900/30 rounded-lg p-3 border border-rose-800">
          <p className="text-[10px] font-semibold text-rose-400 mb-1">課題</p>
          <p className="text-sm font-medium text-rose-800">{weakest.axis}</p>
          <p className="text-xs text-rose-400 mt-0.5">スコア: {weakest.score}/10</p>
        </div>
      </div>

      {/* 改善アドバイス */}
      <div className="bg-amber-900/30 rounded-lg p-3 border border-amber-800 mb-3">
        <p className="text-[10px] font-semibold text-amber-400 mb-1">改善アドバイス</p>
        <p className="text-xs text-amber-800">{advice}</p>
      </div>

      <div className="flex flex-col gap-2">
        {scenarioId && (
          <button
            onClick={() => startSession({ id: scenarioId, name: scenarioName })}
            disabled={starting}
            className="w-full py-2 text-sm font-medium text-white bg-primary-600 hover:bg-primary-500 disabled:opacity-50 rounded-lg transition-colors"
          >
            もう一度練習
          </button>
        )}
        <button
          onClick={() => navigate('/practice')}
          className="w-full py-2 text-sm font-medium text-primary-400 bg-surface-2 hover:bg-surface-3 rounded-lg transition-colors"
        >
          次の練習へ
        </button>
      </div>
    </div>
  );
}
