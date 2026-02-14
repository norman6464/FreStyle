export function getScoreTextColor(score: number): string {
  if (score >= 8) return 'text-emerald-400';
  if (score >= 6) return 'text-amber-400';
  return 'text-rose-400';
}

export function getScoreBarColor(score: number): string {
  if (score >= 8) return 'bg-emerald-500';
  if (score >= 6) return 'bg-amber-500';
  return 'bg-rose-500';
}

export function getScoreLevel(score: number): { label: string; color: string } {
  if (score >= 8) return { label: '優秀レベル', color: 'bg-emerald-900/30 text-emerald-400 border-emerald-800' };
  if (score >= 5) return { label: '実務レベル', color: 'bg-amber-900/30 text-amber-400 border-amber-800' };
  return { label: '基礎レベル', color: 'bg-rose-900/30 text-rose-400 border-rose-800' };
}

export function getDeltaColor(delta: number): string {
  if (delta > 0) return 'text-emerald-400';
  if (delta < 0) return 'text-rose-400';
  return 'text-[var(--color-text-faint)]';
}

export function formatDelta(delta: number): string {
  if (delta === 0) return '±0';
  return delta > 0 ? `+${delta.toFixed(1)}` : `${delta.toFixed(1)}`;
}
