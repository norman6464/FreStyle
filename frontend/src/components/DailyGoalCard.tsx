import { useDailyGoal } from '../hooks/useDailyGoal';

export default function DailyGoalCard() {
  const { goal, isAchieved, progress } = useDailyGoal();

  return (
    <div className="bg-white rounded-lg border border-slate-200 p-4">
      <div className="flex items-center justify-between mb-2">
        <p className="text-xs font-medium text-slate-700">今日の目標</p>
        <span className="text-xs text-slate-500">
          <span className="font-semibold text-slate-800">{goal.completed}</span>
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
        className="w-full bg-slate-100 rounded-full h-2"
      >
        <div
          className={`h-2 rounded-full transition-all duration-300 ${
            isAchieved ? 'bg-emerald-500' : 'bg-primary-500'
          }`}
          style={{ width: `${Math.min(progress, 100)}%` }}
        />
      </div>

      {isAchieved && (
        <p className="text-xs text-emerald-600 font-medium mt-2">
          目標達成！お疲れさまでした
        </p>
      )}
    </div>
  );
}
