interface ScoreDistributionCardProps {
  scores: number[];
}

const RANGES = [
  { label: '9-10', min: 9, max: 10 },
  { label: '7-8', min: 7, max: 8.9 },
  { label: '4-6', min: 4, max: 6.9 },
  { label: '1-3', min: 1, max: 3.9 },
] as const;

const RANGE_COLORS = [
  'bg-emerald-900/300',
  'bg-blue-900/300',
  'bg-amber-900/300',
  'bg-rose-900/300',
];

function getMessage(topRangeIndex: number): string {
  switch (topRangeIndex) {
    case 0:
      return '高スコア帯に集中しています。素晴らしい成果です！';
    case 1:
      return '安定した実力が身についています。さらに上を目指しましょう！';
    case 2:
      return '基礎力がついてきています。練習を続けてレベルアップしましょう！';
    default:
      return '伸びしろがたくさんあります。練習を重ねてスキルを磨きましょう！';
  }
}

export default function ScoreDistributionCard({ scores }: ScoreDistributionCardProps) {
  if (scores.length === 0) return null;

  const counts = RANGES.map((range) =>
    scores.filter((s) => s >= range.min && s <= range.max).length
  );

  const maxCount = Math.max(...counts);
  const topRangeIndex = counts.indexOf(maxCount);

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <p className="text-xs font-medium text-[#D0D0D0] mb-3">スコア分布</p>

      <div className="space-y-2">
        {RANGES.map((range, i) => (
          <div key={range.label} className="flex items-center gap-2">
            <span className="text-xs text-[#888888] w-8 text-right">{range.label}</span>
            <div className="flex-1 bg-surface-3 rounded-full h-4">
              {maxCount > 0 && counts[i] > 0 && (
                <div
                  className={`h-4 rounded-full ${RANGE_COLORS[i]} transition-all`}
                  style={{ width: `${(counts[i] / maxCount) * 100}%` }}
                />
              )}
            </div>
            <span data-testid="range-count" className="text-xs font-medium text-[#A0A0A0] w-6 text-right">
              {counts[i]}
            </span>
          </div>
        ))}
      </div>

      <p className="text-xs text-[#888888] mt-3">{getMessage(topRangeIndex)}</p>
    </div>
  );
}
