interface AxisScore {
  axis: string;
  score: number;
  comment: string;
}

interface SkillSummaryCardProps {
  scores: AxisScore[];
}

export default function SkillSummaryCard({ scores }: SkillSummaryCardProps) {
  if (scores.length === 0) return null;

  const sorted = [...scores].sort((a, b) => b.score - a.score);
  const strengths = sorted.slice(0, 2);
  const weaknesses = sorted.slice(-2).reverse();

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <p className="text-xs font-medium text-[#D0D0D0] mb-3">スキル強弱サマリー</p>

      <div className="grid grid-cols-2 gap-3">
        <div data-testid="strengths">
          <p className="text-[10px] font-medium text-emerald-400 mb-2">強み</p>
          <div className="space-y-2">
            {strengths.map((s) => (
              <div key={s.axis} className="flex items-center justify-between">
                <span className="text-xs text-[#D0D0D0] truncate mr-2">{s.axis}</span>
                <span className="text-sm font-bold text-emerald-400">{s.score}</span>
              </div>
            ))}
          </div>
        </div>

        <div data-testid="weaknesses">
          <p className="text-[10px] font-medium text-rose-400 mb-2">課題</p>
          <div className="space-y-2">
            {weaknesses.map((s) => (
              <div key={s.axis} className="flex items-center justify-between">
                <span className="text-xs text-[#D0D0D0] truncate mr-2">{s.axis}</span>
                <span className="text-sm font-bold text-rose-400">{s.score}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
