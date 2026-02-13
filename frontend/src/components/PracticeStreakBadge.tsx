interface PracticeStreakBadgeProps {
  streakDays: number;
  totalSessions: number;
}

const BADGES = [
  { label: '初回練習', condition: (streak: number, total: number) => total >= 1 },
  { label: '3日連続', condition: (streak: number) => streak >= 3 },
  { label: '7日連続', condition: (streak: number) => streak >= 7 },
  { label: '10回達成', condition: (_streak: number, total: number) => total >= 10 },
] as const;

export default function PracticeStreakBadge({ streakDays, totalSessions }: PracticeStreakBadgeProps) {
  return (
    <div className="bg-white rounded-lg border border-slate-200 p-4">
      <div className="flex items-center gap-4 mb-3">
        <div className="text-center">
          <p className="text-2xl font-bold text-primary-600">{streakDays}</p>
          <p className="text-[10px] text-slate-500">日連続</p>
        </div>
        <div className="h-8 w-px bg-slate-200" />
        <div className="text-center">
          <p className="text-sm font-semibold text-slate-700">累計 {totalSessions}回</p>
          <p className="text-[10px] text-slate-400">練習セッション</p>
        </div>
      </div>

      <div className="flex gap-2 flex-wrap">
        {BADGES.map(({ label, condition }) => {
          const achieved = condition(streakDays, totalSessions);
          return (
            <div
              key={label}
              className={`text-[10px] font-medium px-2 py-1 rounded-full border ${
                achieved
                  ? 'bg-primary-50 text-primary-700 border-primary-200'
                  : 'bg-slate-50 text-slate-400 border-slate-200 opacity-40'
              }`}
            >
              {label}
            </div>
          );
        })}
      </div>
    </div>
  );
}
