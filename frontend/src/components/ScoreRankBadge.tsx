interface ScoreRankBadgeProps {
  score: number;
}

interface Rank {
  letter: string;
  label: string;
  bgColor: string;
  textColor: string;
}

function getRank(score: number): Rank {
  if (score >= 9.0) return { letter: 'S', label: 'エキスパート', bgColor: 'bg-amber-100', textColor: 'text-amber-400' };
  if (score >= 8.0) return { letter: 'A', label: '上級', bgColor: 'bg-surface-3', textColor: 'text-[var(--color-text-secondary)]' };
  if (score >= 7.0) return { letter: 'B', label: '中級', bgColor: 'bg-orange-100', textColor: 'text-orange-700' };
  if (score >= 6.0) return { letter: 'C', label: '初級', bgColor: 'bg-emerald-100', textColor: 'text-emerald-400' };
  return { letter: 'D', label: '入門', bgColor: 'bg-surface-3', textColor: 'text-[var(--color-text-muted)]' };
}

export default function ScoreRankBadge({ score }: ScoreRankBadgeProps) {
  const rank = getRank(score);

  return (
    <div className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full ${rank.bgColor}`}>
      <span className={`text-sm font-bold ${rank.textColor}`}>{rank.letter}</span>
      <span className={`text-[10px] font-medium ${rank.textColor}`}>{rank.label}</span>
    </div>
  );
}
