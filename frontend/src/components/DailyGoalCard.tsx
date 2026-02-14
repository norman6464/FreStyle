import { useDailyGoal } from '../hooks/useDailyGoal';

export default function DailyGoalCard() {
  const { goal, isAchieved, progress } = useDailyGoal();

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <div className="flex items-center justify-between mb-2">
        <p className="text-xs font-medium text-[#D0D0D0]">今日の目標</p>
        <span className="text-xs text-[#888888]">
          <span className="font-semibold text-[#F0F0F0]">{goal.completed}</span>
          {' / '}
          <span>{goal.target}</span>
          {' 回'}
        </span>
      </div>

      <div
        role="progressbar"
        aria-valuenow={progress}
        aria-valuemin={0}
        aria-valuemax={100}
        className="w-full bg-surface-3 rounded-full h-2"
      >
        <div
          className={`h-2 rounded-full transition-all duration-300 ${
            isAchieved ? 'bg-emerald-900/300' : 'bg-primary-500'
          }`}
          style={{ width: `${Math.min(progress, 100)}%` }}
        />
      </div>

      {isAchieved && (
        <p className="text-xs text-emerald-400 font-medium mt-2">
          目標達成！お疲れさまでした
        </p>
      )}
    </div>
  );
}
