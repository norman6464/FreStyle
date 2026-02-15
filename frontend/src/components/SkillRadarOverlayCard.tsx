import Card from './Card';
import CardHeading from './CardHeading';
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

interface SkillRadarOverlayCardProps {
  previousScores: AxisScore[];
  currentScores: AxisScore[];
}

export default function SkillRadarOverlayCard({ previousScores, currentScores }: SkillRadarOverlayCardProps) {
  const axisCount = Math.max(previousScores.length, currentScores.length, 5);
  const angleStep = 360 / axisCount;
  const labels = currentScores.length > 0 ? currentScores : previousScores;

  return (
    <Card>
      <CardHeading>スキル変化レーダー</CardHeading>

      <div className="flex justify-center">
        <svg viewBox={`0 0 ${RADAR_SIZE} ${RADAR_SIZE}`} width={RADAR_SIZE} height={RADAR_SIZE}>
          {/* グリッド線 */}
          {RADAR_GRID_LEVELS.map((level) => (
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
            const point = polarToCartesian(i * angleStep, RADAR_RADIUS + 18);
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
