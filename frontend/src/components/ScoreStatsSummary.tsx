interface ScoreStatsSummaryProps {
  history: { sessionId: number; overallScore: number }[];
}

export default function ScoreStatsSummary({ history }: ScoreStatsSummaryProps) {
  if (history.length === 0) return null;

  const total = history.length;
  const avg = Math.round((history.reduce((sum, h) => sum + h.overallScore, 0) / total) * 10) / 10;
  const best = Math.max(...history.map((h) => h.overallScore));

  return (
    <div className="grid grid-cols-3 gap-2">
      <div className="bg-surface-1 rounded-lg border border-surface-3 p-3 text-center">
        <p className="text-lg font-bold text-[#F0F0F0]">{total}</p>
        <p className="text-[10px] text-[#888888]">総セッション</p>
      </div>
      <div className="bg-surface-1 rounded-lg border border-surface-3 p-3 text-center">
        <p className="text-lg font-bold text-[#F0F0F0]">{avg}</p>
        <p className="text-[10px] text-[#888888]">平均スコア</p>
      </div>
      <div className="bg-surface-1 rounded-lg border border-surface-3 p-3 text-center">
        <p className="text-lg font-bold text-amber-400">{best}</p>
        <p className="text-[10px] text-[#888888]">最高スコア</p>
      </div>
    </div>
  );
}
