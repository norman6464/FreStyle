export const RADAR_SIZE = 200;
export const RADAR_CENTER = RADAR_SIZE / 2;
export const RADAR_RADIUS = 70;
export const RADAR_GRID_LEVELS = [2, 4, 6, 8, 10];

export function polarToCartesian(angle: number, radius: number): { x: number; y: number } {
  const rad = ((angle - 90) * Math.PI) / 180;
  return {
    x: RADAR_CENTER + radius * Math.cos(rad),
    y: RADAR_CENTER + radius * Math.sin(rad),
  };
}

export function getPolygonPoints(values: number[], maxValue: number): string {
  if (values.length === 0) return '';
  const angleStep = 360 / values.length;
  return values
    .map((value, i) => {
      const ratio = value / maxValue;
      const point = polarToCartesian(i * angleStep, RADAR_RADIUS * ratio);
      return `${point.x},${point.y}`;
    })
    .join(' ');
}

export function getGridPoints(level: number, count: number): string {
  if (count === 0) return '';
  const angleStep = 360 / count;
  const ratio = level / 10;
  return Array.from({ length: count })
    .map((_, i) => {
      const point = polarToCartesian(i * angleStep, RADAR_RADIUS * ratio);
      return `${point.x},${point.y}`;
    })
    .join(' ');
}
