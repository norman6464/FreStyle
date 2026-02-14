interface WeeklyGoalProgressCardProps {
  sessionsThisWeek: number;
  weeklyGoal: number;
}

function getMessage(current: number, goal: number): string {
  if (current === 0) return '今週はまだ練習していません。始めてみましょう！';
  if (current >= goal) return '目標達成！素晴らしいです！';
  const remaining = goal - current;
  return `いい調子です！あと${remaining}回で目標達成！`;
}

export default function WeeklyGoalProgressCard({
  sessionsThisWeek,
  weeklyGoal,
}: WeeklyGoalProgressCardProps) {
  const percentage = Math.min((sessionsThisWeek / weeklyGoal) * 100, 100);
  const isCompleted = sessionsThisWeek >= weeklyGoal;

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-2">今週の練習目標</p>

      <div className="flex items-baseline gap-1 mb-2">
        <span className={`text-2xl font-bold ${isCompleted ? 'text-emerald-400' : 'text-primary-400'}`}>
          {sessionsThisWeek}
        </span>
        <span className="text-sm text-[var(--color-text-muted)]">/ {weeklyGoal} 回</span>
      </div>

      <div className="w-full bg-surface-3 rounded-full h-2 mb-2">
        <div
          data-testid="progress-bar"
          className={`h-2 rounded-full transition-all ${isCompleted ? 'bg-emerald-900/300' : 'bg-primary-500'}`}
          style={{ width: `${percentage}%` }}
        />
      </div>

      <p className={`text-xs ${isCompleted ? 'text-emerald-400 font-medium' : 'text-[var(--color-text-muted)]'}`}>
        {getMessage(sessionsThisWeek, weeklyGoal)}
      </p>
    </div>
  );
}
