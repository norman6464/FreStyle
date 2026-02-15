import { ArrowTrendingUpIcon, ArrowTrendingDownIcon } from '@heroicons/react/24/solid';
import Card from './Card';

interface ScoreSparklineProps {
  scores: number[];
}

export default function ScoreSparkline({ scores }: ScoreSparklineProps) {
  if (scores.length < 2) {
    return (
      <Card>
        <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-1">スコア推移</p>
        <p className="text-xs text-[var(--color-text-muted)]">データ不足</p>
      </Card>
    );
  }

  const recent = scores.slice(-5);
  const latest = recent[recent.length - 1];
  const previous = recent[recent.length - 2];
  const isUp = latest > previous;
  const isDown = latest < previous;

  const min = Math.min(...recent);
  const max = Math.max(...recent);
  const range = max - min || 1;

  const width = 120;
  const height = 32;
  const padding = 2;

  const points = recent.map((score, i) => {
    const x = padding + (i / (recent.length - 1)) * (width - padding * 2);
    const y = padding + (1 - (score - min) / range) * (height - padding * 2);
    return `${x},${y}`;
  }).join(' ');

  const strokeColor = isUp ? '#34d399' : isDown ? '#f87171' : '#9ba4b5';

  return (
    <Card>
      <div className="flex items-center justify-between mb-2">
        <p className="text-xs font-medium text-[var(--color-text-secondary)]">スコア推移</p>
        <div className="flex items-center gap-1.5">
          <span className="text-sm font-bold text-[var(--color-text-primary)]">{latest}</span>
          {isUp && (
            <ArrowTrendingUpIcon data-testid="trend-up" className="w-4 h-4 text-emerald-400" />
          )}
          {isDown && (
            <ArrowTrendingDownIcon data-testid="trend-down" className="w-4 h-4 text-rose-400" />
          )}
        </div>
      </div>

      <svg width={width} height={height} className="w-full" viewBox={`0 0 ${width} ${height}`} preserveAspectRatio="none">
        <polyline
          points={points}
          fill="none"
          stroke={strokeColor}
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        {recent.map((score, i) => {
          const x = padding + (i / (recent.length - 1)) * (width - padding * 2);
          const y = padding + (1 - (score - min) / range) * (height - padding * 2);
          return (
            <circle
              key={i}
              cx={x}
              cy={y}
              r={i === recent.length - 1 ? 3 : 2}
              fill={i === recent.length - 1 ? strokeColor : 'var(--color-surface-3)'}
              stroke={strokeColor}
              strokeWidth="1"
            />
          );
        })}
      </svg>
    </Card>
  );
}
