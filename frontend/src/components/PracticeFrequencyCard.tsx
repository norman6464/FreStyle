import { useMemo } from 'react';

interface PracticeFrequencyCardProps {
  dates: string[];
}

function getWeekStart(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay();
  const diff = day === 0 ? 6 : day - 1; // 月曜始まり
  d.setDate(d.getDate() - diff);
  d.setHours(0, 0, 0, 0);
  return d;
}

export default function PracticeFrequencyCard({ dates }: PracticeFrequencyCardProps) {
  const weeks = useMemo(() => {
    const now = new Date();
    const currentWeekStart = getWeekStart(now);

    const result: { label: string; count: number }[] = [];

    for (let i = 3; i >= 0; i--) {
      const weekStart = new Date(currentWeekStart);
      weekStart.setDate(weekStart.getDate() - i * 7);
      const weekEnd = new Date(weekStart);
      weekEnd.setDate(weekEnd.getDate() + 7);

      const count = dates.filter((d) => {
        const date = new Date(d);
        return date >= weekStart && date < weekEnd;
      }).length;

      const label = i === 0 ? '今週' : `${i}週前`;
      result.push({ label, count });
    }

    return result;
  }, [dates]);

  const maxCount = Math.max(...weeks.map((w) => w.count), 1);
  const currentWeek = weeks[weeks.length - 1];
  const lastWeek = weeks[weeks.length - 2];
  const delta = currentWeek.count - lastWeek.count;

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-[#D0D0D0]">週別練習頻度</p>
        <div className="flex items-center gap-2">
          <span className="text-lg font-bold text-[#F0F0F0]">{currentWeek.count}回</span>
          {delta !== 0 && (
            <span
              className={`text-xs font-medium ${delta > 0 ? 'text-emerald-400' : 'text-rose-400'}`}
            >
              {delta > 0 ? `+${delta}` : `\u2212${Math.abs(delta)}`}
            </span>
          )}
        </div>
      </div>

      <div className="flex items-end gap-2 h-16">
        {weeks.map((week) => (
          <div key={week.label} className="flex-1 flex flex-col items-center gap-1">
            <div className="w-full flex items-end h-12">
              <div
                data-testid="frequency-bar"
                className={`w-full rounded-t ${
                  week.label === '今週' ? 'bg-primary-500' : 'bg-surface-3'
                }`}
                style={{ height: `${(week.count / maxCount) * 100}%` }}
              />
            </div>
            <span
              className={`text-[10px] ${
                week.label === '今週' ? 'font-semibold text-primary-300' : 'text-[#888888]'
              }`}
            >
              {week.label}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
