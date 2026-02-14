import { ArrowTrendingUpIcon, ArrowTrendingDownIcon, MinusIcon } from '@heroicons/react/24/outline';

interface ScoreTrendIndicatorProps {
  scores: number[];
}

type Trend = 'up' | 'down' | 'stable';

function calculateTrend(scores: number[]): Trend {
  const recent = scores.slice(-5);
  const firstHalf = recent.slice(0, Math.ceil(recent.length / 2));
  const secondHalf = recent.slice(Math.floor(recent.length / 2));

  const avgFirst = firstHalf.reduce((s, v) => s + v, 0) / firstHalf.length;
  const avgSecond = secondHalf.reduce((s, v) => s + v, 0) / secondHalf.length;

  const diff = avgSecond - avgFirst;
  if (diff > 0.3) return 'up';
  if (diff < -0.3) return 'down';
  return 'stable';
}

const TREND_CONFIG = {
  up: {
    label: '上昇傾向',
    icon: ArrowTrendingUpIcon,
    color: 'text-emerald-400',
  },
  down: {
    label: '下降傾向',
    icon: ArrowTrendingDownIcon,
    color: 'text-red-400',
  },
  stable: {
    label: '安定',
    icon: MinusIcon,
    color: 'text-[var(--color-text-muted)]',
  },
} as const;

export default function ScoreTrendIndicator({ scores }: ScoreTrendIndicatorProps) {
  if (!Array.isArray(scores) || scores.length < 2) return null;

  const trend = calculateTrend(scores);
  const config = TREND_CONFIG[trend];
  const Icon = config.icon;

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-3 flex items-center gap-3">
      <Icon className={`w-5 h-5 ${config.color} flex-shrink-0`} />
      <div>
        <p className="text-[10px] text-[var(--color-text-muted)]">直近トレンド</p>
        <p className={`text-sm font-semibold ${config.color}`}>{config.label}</p>
      </div>
    </div>
  );
}
