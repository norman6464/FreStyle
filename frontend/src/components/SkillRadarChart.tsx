import type { AxisScore } from '../types';

interface SkillRadarChartProps {
  scores: AxisScore[];
  title?: string;
}

const SIZE = 200;
const CENTER = SIZE / 2;
const RADIUS = 70;
const GRID_LEVELS = [2, 4, 6, 8, 10];

function polarToCartesian(angle: number, radius: number): { x: number; y: number } {
  // Start from top (-90 degrees)
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

export default function SkillRadarChart({ scores, title }: SkillRadarChartProps) {
  const axisCount = scores.length || 5;
  const angleStep = 360 / axisCount;

  return (
    <div className="flex flex-col items-center">
      {title && (
        <h4 className="text-xs font-semibold text-[#D0D0D0] mb-2">{title}</h4>
      )}
      <svg viewBox={`0 0 ${SIZE} ${SIZE}`} width={SIZE} height={SIZE}>
        {/* グリッド線 */}
        {GRID_LEVELS.map((level) => (
          <polygon
            key={level}
            className="grid"
            points={getGridPoints(level, axisCount)}
            fill="none"
            stroke="rgb(226, 232, 240)"
            strokeWidth={level === 10 ? 1 : 0.5}
          />
        ))}

        {/* 軸線 */}
        {scores.map((_, i) => {
          const point = polarToCartesian(i * angleStep, RADIUS);
          return (
            <line
              key={i}
              x1={CENTER}
              y1={CENTER}
              x2={point.x}
              y2={point.y}
              stroke="rgb(226, 232, 240)"
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
          const point = polarToCartesian(i * angleStep, RADIUS * ratio);
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
          const point = polarToCartesian(i * angleStep, RADIUS + 18);
          return (
            <text
              key={i}
              x={point.x}
              y={point.y}
              textAnchor="middle"
              dominantBaseline="middle"
              className="fill-slate-600"
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
