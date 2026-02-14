import Card from './Card';
import type { AxisScore } from '../types';

interface SkillRadarOverlayCardProps {
  previousScores: AxisScore[];
  currentScores: AxisScore[];
}

const SIZE = 200;
const CENTER = SIZE / 2;
const RADIUS = 70;
const GRID_LEVELS = [2, 4, 6, 8, 10];

function polarToCartesian(angle: number, radius: number): { x: number; y: number } {
  const rad = ((angle - 90) * Math.PI) / 180;
  return {
    x: CENTER + radius * Math.cos(rad),
    y: CENTER + radius * Math.sin(rad),
  };
}

function getPolygonPoints(values: number[], maxValue: number): string {
  if (values.length === 0) return '';
  const angleStep = 360 / values.length;
  return values
    .map((value, i) => {
      const ratio = value / maxValue;
      const point = polarToCartesian(i * angleStep, RADIUS * ratio);
      return `${point.x},${point.y}`;
    })
    .join(' ');
}

function getGridPoints(level: number, count: number): string {
  if (count === 0) return '';
  const angleStep = 360 / count;
  const ratio = level / 10;
  return Array.from({ length: count })
    .map((_, i) => {
      const point = polarToCartesian(i * angleStep, RADIUS * ratio);
      return `${point.x},${point.y}`;
    })
    .join(' ');
}

export default function SkillRadarOverlayCard({ previousScores, currentScores }: SkillRadarOverlayCardProps) {
  const axisCount = Math.max(previousScores.length, currentScores.length, 5);
  const angleStep = 360 / axisCount;
  const labels = currentScores.length > 0 ? currentScores : previousScores;

  return (
    <Card>
      <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-3">スキル変化レーダー</p>

      <div className="flex justify-center">
        <svg viewBox={`0 0 ${SIZE} ${SIZE}`} width={SIZE} height={SIZE}>
          {/* グリッド線 */}
          {GRID_LEVELS.map((level) => (
            <polygon
              key={level}
              points={getGridPoints(level, axisCount)}
              fill="none"
              stroke="var(--color-surface-3)"
              strokeWidth={level === 10 ? 1 : 0.5}
            />
          ))}

          {/* 軸線 */}
          {labels.map((_, i) => {
            const point = polarToCartesian(i * angleStep, RADIUS);
            return (
              <line
                key={i}
                x1={CENTER}
                y1={CENTER}
                x2={point.x}
                y2={point.y}
                stroke="var(--color-surface-3)"
                strokeWidth={0.5}
              />
            );
          })}

          {/* 前回のポリゴン */}
          {previousScores.length > 0 && (
            <polygon
              data-testid="prev-polygon"
              points={getPolygonPoints(previousScores.map((s) => s.score), 10)}
              fill="rgba(148, 163, 184, 0.15)"
              stroke="rgba(148, 163, 184, 0.6)"
              strokeWidth={1}
              strokeDasharray="4 2"
            />
          )}

          {/* 今回のポリゴン */}
          {currentScores.length > 0 && (
            <polygon
              data-testid="current-polygon"
              points={getPolygonPoints(currentScores.map((s) => s.score), 10)}
              fill="rgba(99, 102, 241, 0.2)"
              stroke="rgb(99, 102, 241)"
              strokeWidth={1.5}
            />
          )}

          {/* ラベル */}
          {labels.map((s, i) => {
            const point = polarToCartesian(i * angleStep, RADIUS + 18);
            return (
              <text
                key={i}
                x={point.x}
                y={point.y}
                textAnchor="middle"
                dominantBaseline="middle"
                fill="var(--color-text-tertiary)"
                fontSize={8}
              >
                {s.axis}
              </text>
            );
          })}
        </svg>
      </div>

      {/* 凡例 */}
      <div className="flex justify-center gap-4 mt-2">
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-0.5 bg-[#94a3b8] opacity-60" style={{ borderTop: '2px dashed #94a3b8' }} />
          <span className="text-[10px] text-[var(--color-text-muted)]">前回</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-0.5 bg-indigo-500" />
          <span className="text-[10px] text-[var(--color-text-muted)]">今回</span>
        </div>
      </div>
    </Card>
  );
}
