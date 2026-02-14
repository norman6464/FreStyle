import { useNavigate } from 'react-router-dom';
import type { ScoreCard } from '../types';

interface PracticeResultSummaryProps {
  scoreCard: ScoreCard;
  scenarioName: string;
}

const IMPROVEMENT_ADVICE: Record<string, string> = {
  '論理的構成力': '論理的構成力を伸ばすために、結論→理由→具体例の構成を意識して練習しましょう。',
  '配慮表現': '配慮表現を伸ばすために、クッション言葉や敬語のバリエーションを増やしましょう。',
  '要約力': '要約力を伸ばすために、要点を3つに絞って伝える練習をしましょう。',
  '提案力': '提案力を伸ばすために、代替案を複数用意してから発言する習慣をつけましょう。',
  '質問・傾聴力': '質問・傾聴力を伸ばすために、相手の発言を復唱してから質問する練習をしましょう。',
};

export default function PracticeResultSummary({ scoreCard, scenarioName }: PracticeResultSummaryProps) {
  const navigate = useNavigate();

  const sorted = [...scoreCard.scores].sort((a, b) => b.score - a.score);
  const strongest = sorted[0];
  const weakest = sorted[sorted.length - 1];

  const advice = IMPROVEMENT_ADVICE[weakest.axis] || `${weakest.axis}を伸ばすために、繰り返し練習しましょう。`;

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

      {/* 次の練習へ */}
      <button
        onClick={() => navigate('/practice')}
        className="w-full py-2 text-sm font-medium text-primary-400 bg-surface-2 hover:bg-surface-3 rounded-lg transition-colors"
      >
        次の練習へ
      </button>
    </div>
  );
}
