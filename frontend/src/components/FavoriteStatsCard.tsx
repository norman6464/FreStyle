import type { FavoritePhrase } from '../types';

interface FavoriteStatsCardProps {
  phrases: FavoritePhrase[];
}

const PATTERN_COLORS: Record<string, string> = {
  'フォーマル': 'bg-blue-900/300',
  'ソフト': 'bg-emerald-900/300',
  '簡潔': 'bg-amber-900/300',
  '質問型': 'bg-purple-500',
  '提案型': 'bg-rose-900/300',
};

export default function FavoriteStatsCard({ phrases }: FavoriteStatsCardProps) {
  if (phrases.length === 0) return null;

  const patternCounts = phrases.reduce<Record<string, number>>((acc, phrase) => {
    acc[phrase.pattern] = (acc[phrase.pattern] || 0) + 1;
    return acc;
  }, {});

  const sortedPatterns = Object.entries(patternCounts).sort((a, b) => b[1] - a[1]);
  const maxCount = Math.max(...Object.values(patternCounts));

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-[#D0D0D0]">お気に入り統計</p>
        <div className="flex items-baseline gap-1">
          <span className="text-lg font-bold text-[#F0F0F0]">{phrases.length}</span>
          <span className="text-xs text-[#888888]">フレーズ</span>
        </div>
      </div>

      <div className="space-y-2">
        {sortedPatterns.map(([pattern, count]) => (
          <div key={pattern} className="flex items-center gap-2">
            <span className="text-xs text-[#888888] w-16 text-right truncate">{pattern}</span>
            <div className="flex-1 bg-surface-3 rounded-full h-3">
              {maxCount > 0 && count > 0 && (
                <div
                  className={`h-3 rounded-full ${PATTERN_COLORS[pattern] || 'bg-[#666666]'} transition-all`}
                  style={{ width: `${(count / maxCount) * 100}%` }}
                />
              )}
            </div>
            <span data-testid="pattern-count" className="text-xs font-medium text-[#A0A0A0] w-5 text-right">
              {count}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
