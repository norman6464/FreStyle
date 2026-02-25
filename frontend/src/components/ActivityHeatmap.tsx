import Card from './Card';

interface ActivityHeatmapProps {
  practiceDates: string[];
}

function formatDate(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

function getLast365Days(): Date[] {
  const days: Date[] = [];
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  for (let i = 364; i >= 0; i--) {
    const d = new Date(today);
    d.setDate(d.getDate() - i);
    days.push(d);
  }
  return days;
}

function getIntensityClass(count: number): string {
  if (count === 0) return 'bg-surface-3';
  if (count === 1) return 'bg-emerald-800';
  if (count === 2) return 'bg-emerald-600';
  return 'bg-emerald-400';
}

const MONTH_LABELS = ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月'];

export default function ActivityHeatmap({ practiceDates }: ActivityHeatmapProps) {
  const days = getLast365Days();
  const dateCountMap = new Map<string, number>();
  practiceDates.forEach((d) => {
    dateCountMap.set(d, (dateCountMap.get(d) || 0) + 1);
  });

  const uniqueDays = new Set(practiceDates).size;

  // 週ごとにグループ化（日曜始まり）
  const weeks: Date[][] = [];
  let currentWeek: Date[] = [];

  // 最初の週の前にパディング
  const firstDayOfWeek = days[0].getDay();
  for (let i = 0; i < firstDayOfWeek; i++) {
    currentWeek.push(null as unknown as Date);
  }

  days.forEach((day) => {
    currentWeek.push(day);
    if (currentWeek.length === 7) {
      weeks.push(currentWeek);
      currentWeek = [];
    }
  });
  if (currentWeek.length > 0) {
    weeks.push(currentWeek);
  }

  // 月ラベルの位置を計算
  const monthPositions: { label: string; weekIndex: number }[] = [];
  let lastMonth = -1;
  weeks.forEach((week, weekIndex) => {
    const validDay = week.find((d) => d !== null);
    if (validDay) {
      const month = validDay.getMonth();
      if (month !== lastMonth) {
        monthPositions.push({ label: MONTH_LABELS[month], weekIndex });
        lastMonth = month;
      }
    }
  });

  return (
    <Card>
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-[var(--color-text-secondary)]">年間の学習活動</p>
        <p className="text-xs text-[var(--color-text-muted)]">
          <span className="font-bold text-emerald-400">{uniqueDays}日</span> 練習
        </p>
      </div>

      {/* 月ラベル */}
      <div className="flex mb-1 text-[9px] text-[var(--color-text-faint)]" style={{ paddingLeft: '0px' }}>
        {monthPositions.map((pos, i) => (
          <span
            key={i}
            style={{
              position: 'relative',
              left: `${pos.weekIndex * 11}px`,
              marginRight: i < monthPositions.length - 1
                ? `${Math.max(0, (monthPositions[i + 1]?.weekIndex - pos.weekIndex) * 11 - 20)}px`
                : '0px',
            }}
          >
            {pos.label}
          </span>
        ))}
      </div>

      {/* ヒートマップグリッド */}
      <div className="flex gap-[2px] overflow-x-auto">
        {weeks.map((week, weekIndex) => (
          <div key={weekIndex} className="flex flex-col gap-[2px]">
            {week.map((day, dayIndex) => {
              if (!day) {
                return <div key={dayIndex} className="w-[9px] h-[9px]" />;
              }
              const dateStr = formatDate(day);
              const count = dateCountMap.get(dateStr) || 0;
              return (
                <div
                  key={dayIndex}
                  title={dateStr}
                  className={`w-[9px] h-[9px] rounded-[2px] ${getIntensityClass(count)}`}
                />
              );
            })}
          </div>
        ))}
      </div>
    </Card>
  );
}
