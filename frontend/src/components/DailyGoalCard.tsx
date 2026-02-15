import { useDailyGoal } from '../hooks/useDailyGoal';
import Card from './Card';
import ProgressBar from './ProgressBar';

export default function DailyGoalCard() {
  const { goal, isAchieved, progress } = useDailyGoal();

  return (
    <Card>
      <div className="flex items-center justify-between mb-2">
        <p className="text-xs font-medium text-[var(--color-text-secondary)]">今日の目標</p>
        <span className="text-xs text-[var(--color-text-muted)]">
          <span className="font-semibold text-[var(--color-text-primary)]">{goal.completed}</span>
          {' / '}
          <span>{goal.target}</span>
          {' 回'}
        </span>
      </div>

      <ProgressBar
        percentage={Math.min(progress, 100)}
        barColorClass={isAchieved ? 'bg-emerald-500' : 'bg-primary-500'}
      />

      {isAchieved && (
        <p className="text-xs text-emerald-400 font-medium mt-2">
          目標達成！お疲れさまでした
        </p>
      )}
    </Card>
  );
}
