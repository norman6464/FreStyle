export function toLocalDateKey(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

export function getCalendarDays(weeks: number): Date[] {
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const dayOfWeek = today.getDay();
  const endOfWeek = new Date(today);
  endOfWeek.setDate(today.getDate() + (6 - dayOfWeek));

  const startDate = new Date(endOfWeek);
  startDate.setDate(endOfWeek.getDate() - weeks * 7 + 1);

  const days: Date[] = [];
  const current = new Date(startDate);
  while (current <= endOfWeek) {
    days.push(new Date(current));
    current.setDate(current.getDate() + 1);
  }

  return days;
}

export function getIntensityClass(count: number): string {
  if (count === 0) return 'bg-surface-3';
  if (count === 1) return 'bg-emerald-200';
  if (count === 2) return 'bg-emerald-400';
  return 'bg-emerald-600';
}
