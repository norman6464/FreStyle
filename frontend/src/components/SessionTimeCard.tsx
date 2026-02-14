interface SessionTimeCardProps {
  dates: string[];
}

const TIME_SLOTS = [
  { label: '朝', emoji: '🌅', min: 6, max: 11 },
  { label: '昼', emoji: '☀️', min: 12, max: 16 },
  { label: '夕方', emoji: '🌇', min: 17, max: 20 },
  { label: '夜', emoji: '🌙', min: 21, max: 5, isNight: true },
] as const;

const SLOT_COLORS = [
  'bg-amber-400',
  'bg-sky-400',
  'bg-orange-400',
  'bg-indigo-400',
];

function getTimeSlotIndex(hour: number): number {
  if (hour >= 6 && hour <= 11) return 0;
  if (hour >= 12 && hour <= 16) return 1;
  if (hour >= 17 && hour <= 20) return 2;
  return 3; // 21-5
}

function getMessage(topIndex: number): string {
  switch (topIndex) {
    case 0:
      return '朝型の学習スタイルです。集中力の高い朝に練習できています！';
    case 1:
      return '昼間に集中して練習するタイプです。休憩時間を活用しましょう！';
    case 2:
      return '夕方に練習するスタイルです。仕事終わりにしっかり復習できています！';
    default:
      return '夜型の学習スタイルです。夜の静かな時間に集中しています！';
  }
}

export default function SessionTimeCard({ dates }: SessionTimeCardProps) {
  if (dates.length === 0) return null;

  const counts = [0, 0, 0, 0];
  for (const dateStr of dates) {
    const hour = new Date(dateStr).getHours();
    counts[getTimeSlotIndex(hour)]++;
  }

  const maxCount = Math.max(...counts);
  const topIndex = counts.indexOf(maxCount);

  return (
    <div className="bg-white rounded-lg border border-slate-200 p-4">
      <p className="text-xs font-medium text-slate-700 mb-3">練習時間帯</p>

      <div className="space-y-2">
        {TIME_SLOTS.map((slot, i) => (
          <div key={slot.label} className="flex items-center gap-2">
            <span className="text-xs text-slate-500 w-8 text-right">{slot.label}</span>
            <div className="flex-1 bg-slate-100 rounded-full h-4">
              {maxCount > 0 && counts[i] > 0 && (
                <div
                  className={`h-4 rounded-full ${SLOT_COLORS[i]} transition-all`}
                  style={{ width: `${(counts[i] / maxCount) * 100}%` }}
                />
              )}
            </div>
            <span data-testid="time-count" className="text-xs font-medium text-slate-600 w-5 text-right">
              {counts[i]}
            </span>
          </div>
        ))}
      </div>

      <p className="text-xs text-slate-500 mt-3">{getMessage(topIndex)}</p>
    </div>
  );
}
