import type { AxisScore } from '../types';
import {
  polarToCartesian,
  getPolygonPoints,
  getGridPoints,
  RADAR_SIZE,
  RADAR_CENTER,
  RADAR_RADIUS,
  RADAR_GRID_LEVELS,
} from '../utils/skillRadarHelpers';

interface SkillRadarChartProps {
  scores: AxisScore[];
  title?: string;
}

export default function SkillRadarChart({ scores, title }: SkillRadarChartProps) {
  const axisCount = scores.length || 5;
  const angleStep = 360 / axisCount;

  return (
    <div className="flex flex-col items-center">
      {title && (
        <h4 className="text-xs font-semibold text-[var(--color-text-secondary)] mb-2">{title}</h4>
      )}
      <svg viewBox={`0 0 ${RADAR_SIZE} ${RADAR_SIZE}`} width={RADAR_SIZE} height={RADAR_SIZE}>
        {/* グリッド線 */}
        {RADAR_GRID_LEVELS.map((level) => (
          <polygon
            key={level}
            className="grid"
            points={getGridPoints(level, axisCount)}
            fill="none"
            stroke="var(--color-surface-3)"
            strokeWidth={level === 10 ? 1 : 0.5}
          />
        ))}

        {/* 軸線 */}
        {scores.map((_, i) => {
          const point = polarToCartesian(i * angleStep, RADAR_RADIUS);
          return (
            <line
              key={i}
              x1={RADAR_CENTER}
              y1={RADAR_CENTER}
              x2={point.x}
              y2={point.y}
              stroke="var(--color-surface-3)"
              strokeWidth={0.5}
            />
          );
        })}

        {/* スコアポリゴン */}
        {scores.length > 0 && (
          <polygon
            points={getPolygonPoints(
              scores.map((s) => s.score),
              10
            )}
            fill="rgba(99, 102, 241, 0.2)"
            stroke="rgb(99, 102, 241)"
            strokeWidth={1.5}
          />
        )}

        {/* スコアドット */}
        {scores.map((s, i) => {
          const ratio = s.score / 10;
          const point = polarToCartesian(i * angleStep, RADAR_RADIUS * ratio);
          return (
            <circle
              key={i}
              cx={point.x}
              cy={point.y}
              r={2.5}
              fill="rgb(99, 102, 241)"
            />
          );
        })}

        {/* ラベル */}
        {scores.map((s, i) => {
          const point = polarToCartesian(i * angleStep, RADAR_RADIUS + 18);
          return (
            <text
              key={i}
              x={point.x}
              y={point.y}
              textAnchor="middle"
              dominantBaseline="middle"
              className="fill-[var(--color-text-tertiary)]"
              fontSize={8}
            >
              {s.axis}
            </text>
          );
        })}
      </svg>
    </div>
  );
}
