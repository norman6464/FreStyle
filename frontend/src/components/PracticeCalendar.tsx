import { useMemo } from 'react';

interface PracticeCalendarProps {
  practiceDates: string[];
}

function toLocalDateKey(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

function getCalendarDays(weeks: number): Date[] {
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

function getIntensityClass(count: number): string {
  if (count === 0) return 'bg-surface-3';
  if (count === 1) return 'bg-emerald-200';
  if (count === 2) return 'bg-emerald-400';
  return 'bg-emerald-600';
}

export default function PracticeCalendar({ practiceDates }: PracticeCalendarProps) {
  const dateCountMap = useMemo(() => {
    const map: Record<string, number> = {};
    for (const date of practiceDates) {
      const key = date.split('T')[0];
      map[key] = (map[key] || 0) + 1;
    }
    return map;
  }, [practiceDates]);

  const days = useMemo(() => getCalendarDays(12), []);

  const weeks: Date[][] = useMemo(() => {
    const result: Date[][] = [];
    for (let i = 0; i < days.length; i += 7) {
      result.push(days.slice(i, i + 7));
    }
    return result;
  }, [days]);

  const dayLabels = ['日', '月', '火', '水', '木', '金', '土'];

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-3">練習カレンダー</p>

      <div className="flex gap-0.5">
        {/* 曜日ラベル */}
        <div className="flex flex-col gap-0.5 mr-1">
          {dayLabels.map((label, i) => (
            <div
              key={label}
              className={`h-3 text-[9px] leading-3 text-[var(--color-text-faint)] ${
                i % 2 === 1 ? '' : 'invisible'
              }`}
            >
              {label}
            </div>
          ))}
        </div>

        {/* カレンダーグリッド */}
        {weeks.map((week, wi) => (
          <div key={wi} className="flex flex-col gap-0.5">
            {week.map((day) => {
              const key = toLocalDateKey(day);
              const count = dateCountMap[key] || 0;
              const today = new Date();
              today.setHours(0, 0, 0, 0);
              const isFuture = day > today;

              return (
                <div
                  key={key}
                  className={`w-3 h-3 rounded-sm ${
                    isFuture ? 'bg-transparent' : getIntensityClass(count)
                  }`}
                  data-active={count > 0 ? 'true' : 'false'}
                  data-count={count}
                  title={`${key}: ${count}回`}
                />
              );
            })}
          </div>
        ))}
      </div>
    </div>
  );
}
