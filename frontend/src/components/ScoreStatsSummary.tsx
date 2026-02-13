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
      <div className="bg-white rounded-lg border border-slate-200 p-3 text-center">
        <p className="text-lg font-bold text-slate-800">{total}</p>
        <p className="text-[10px] text-slate-500">総セッション</p>
      </div>
      <div className="bg-white rounded-lg border border-slate-200 p-3 text-center">
        <p className="text-lg font-bold text-slate-800">{avg}</p>
        <p className="text-[10px] text-slate-500">平均スコア</p>
      </div>
      <div className="bg-white rounded-lg border border-slate-200 p-3 text-center">
        <p className="text-lg font-bold text-amber-600">{best}</p>
        <p className="text-[10px] text-slate-500">最高スコア</p>
      </div>
    </div>
  );
}
