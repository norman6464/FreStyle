interface StreakCalendarCardProps {
  practiceDates: string[];
}

function formatDate(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

function calculateStreak(practiceDates: string[]): number {
  if (practiceDates.length === 0) return 0;

  const dateSet = new Set(practiceDates);
  let streak = 0;
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const current = new Date(today);
  while (dateSet.has(formatDate(current))) {
    streak++;
    current.setDate(current.getDate() - 1);
  }

  return streak;
}

function getLast28Days(): Date[] {
  const days: Date[] = [];
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  for (let i = 27; i >= 0; i--) {
    const d = new Date(today);
    d.setDate(d.getDate() - i);
    days.push(d);
  }
  return days;
}

const WEEKDAY_LABELS = ['月', '火', '水', '木', '金', '土', '日'];

export default function StreakCalendarCard({ practiceDates }: StreakCalendarCardProps) {
  const dateSet = new Set(practiceDates);
  const days = getLast28Days();
  const streak = calculateStreak(practiceDates);

  return (
    <div className="bg-white rounded-lg border border-slate-200 p-4">
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-slate-700">練習カレンダー</p>
        <div className="flex items-baseline gap-1">
          <span className="text-lg font-bold text-primary-600">{streak}</span>
          <span className="text-[10px] text-slate-500">日連続</span>
        </div>
      </div>

      <div className="grid grid-cols-7 gap-1 mb-1">
        {WEEKDAY_LABELS.map((label) => (
          <div key={label} className="text-[9px] text-slate-400 text-center">
            {label}
          </div>
        ))}
      </div>

      <div className="grid grid-cols-7 gap-1">
        {days.map((day) => {
          const dateStr = formatDate(day);
          const isActive = dateSet.has(dateStr);
          const isToday = formatDate(new Date()) === dateStr;

          return (
            <div
              key={dateStr}
              data-testid={`calendar-cell-${dateStr}`}
              className={`aspect-square rounded-sm ${
                isActive
                  ? 'bg-primary-500'
                  : 'bg-slate-100'
              } ${isToday ? 'ring-1 ring-primary-300' : ''}`}
              title={dateStr}
            />
          );
        })}
      </div>
    </div>
  );
}
