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
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <div className="flex items-center gap-4 mb-3">
        <div className="text-center">
          <p className="text-2xl font-bold text-primary-400">{streakDays}</p>
          <p className="text-[10px] text-[#888888]">日連続</p>
        </div>
        <div className="h-8 w-px bg-surface-3" />
        <div className="text-center">
          <p className="text-sm font-semibold text-[#D0D0D0]">累計 {totalSessions}回</p>
          <p className="text-[10px] text-[#666666]">練習セッション</p>
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
                  ? 'bg-surface-2 text-primary-300 border-[#444444]'
                  : 'bg-surface-2 text-[#666666] border-surface-3 opacity-40'
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
