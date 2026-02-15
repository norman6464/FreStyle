import { useMemo } from 'react';
import CardHeading from './CardHeading';
import { toLocalDateKey, getCalendarDays, getIntensityClass } from '../utils/calendarHelpers';

interface PracticeCalendarProps {
  practiceDates: string[];
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
      <CardHeading>練習カレンダー</CardHeading>

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
