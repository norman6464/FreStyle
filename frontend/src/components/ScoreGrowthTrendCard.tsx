import { useMemo } from 'react';
import { getDeltaColor, formatDelta } from '../utils/scoreColor';

type TrendType = 'up' | 'down' | 'stable';

const TREND_CONFIG: Record<TrendType, { label: string; message: string; color: string; icon: string }> = {
  up: {
    label: '上昇',
    message: 'スコアが上昇傾向です。この調子で練習を続けましょう！',
    color: 'text-emerald-400',
    icon: '↑',
  },
  down: {
    label: '低下',
    message: 'スコアが少し低下しています。基礎に立ち返って練習してみましょう。',
    color: 'text-rose-400',
    icon: '↓',
  },
  stable: {
    label: '安定',
    message: 'スコアが安定しています。新しいシナリオに挑戦してレベルアップしましょう！',
    color: 'text-blue-400',
    icon: '→',
  },
};

interface Props {
  scores: number[];
}

export default function ScoreGrowthTrendCard({ scores }: Props) {
  const analysis = useMemo(() => {
    if (scores.length < 2) return null;

    const mid = Math.floor(scores.length / 2);
    const firstHalf = scores.slice(0, mid);
    const secondHalf = scores.slice(mid);

    const firstAvg = firstHalf.reduce((s, v) => s + v, 0) / firstHalf.length;
    const secondAvg = secondHalf.reduce((s, v) => s + v, 0) / secondHalf.length;
    const delta = Math.round((secondAvg - firstAvg) * 10) / 10;

    let trend: TrendType;
    if (delta > 0.3) trend = 'up';
    else if (delta < -0.3) trend = 'down';
    else trend = 'stable';

    return {
      trend,
      delta,
      latestScore: scores[scores.length - 1],
    };
  }, [scores]);

  if (!analysis) return null;

  const config = TREND_CONFIG[analysis.trend];

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <div className="flex items-center justify-between mb-2">
        <p className="text-xs font-medium text-[var(--color-text-secondary)]">成長トレンド</p>
        <span className={`text-sm font-bold ${config.color}`}>
          {config.icon} {config.label}
        </span>
      </div>

      <div className="flex items-baseline gap-2 mb-2">
        <span className="text-2xl font-bold text-[var(--color-text-primary)]">{analysis.latestScore}</span>
        <span className="text-xs text-[var(--color-text-faint)]">最新スコア</span>
        {analysis.delta !== 0 && (
          <span className={`text-xs font-medium ${getDeltaColor(analysis.delta)}`}>
            {formatDelta(analysis.delta)}
          </span>
        )}
      </div>

      <p className="text-xs text-[var(--color-text-muted)]">{config.message}</p>
    </div>
  );
}
