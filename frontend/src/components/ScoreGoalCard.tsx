import Card from './Card';
import { useLocalStorage } from '../hooks/useLocalStorage';

interface Props {
  averageScore: number;
}

const GOAL_OPTIONS = [6.0, 6.5, 7.0, 7.5, 8.0, 8.5, 9.0, 9.5, 10.0];

export default function ScoreGoalCard({ averageScore }: Props) {
  const [goal, setGoal] = useLocalStorage('scoreGoal', 8.0);

  const handleGoalChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setGoal(parseFloat(e.target.value));
  };

  const achieved = averageScore >= goal;
  const gap = Math.round((goal - averageScore) * 10) / 10;
  const progress = Math.min((averageScore / goal) * 100, 100);

  return (
    <Card>
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-[var(--color-text-secondary)]">目標スコア</p>
        <select
          value={goal}
          onChange={handleGoalChange}
          className="text-xs border border-surface-3 rounded px-2 py-1 text-[var(--color-text-tertiary)]"
        >
          {GOAL_OPTIONS.map((opt) => (
            <option key={opt} value={opt}>
              {opt.toFixed(1)}
            </option>
          ))}
        </select>
      </div>

      <div className="flex items-end justify-between mb-2">
        <div>
          <p className="text-xs text-[var(--color-text-muted)]">現在の平均</p>
          <p className="text-xl font-bold text-[var(--color-text-primary)]">{averageScore.toFixed(1)}</p>
        </div>
        <div className="text-right">
          <p className="text-xs text-[var(--color-text-muted)]">目標</p>
          <p className="text-xl font-bold text-primary-400">{goal.toFixed(1)}</p>
        </div>
      </div>

      <div className="w-full bg-surface-3 rounded-full h-2 mb-2">
        <div
          className={`h-2 rounded-full transition-all ${achieved ? 'bg-emerald-900/300' : 'bg-primary-500'}`}
          style={{ width: `${progress}%` }}
        />
      </div>

      <p className={`text-xs ${achieved ? 'text-emerald-400 font-medium' : 'text-[var(--color-text-muted)]'}`}>
        {achieved
          ? '目標達成！素晴らしいです！'
          : `あと ${gap.toFixed(1)} ポイントで目標達成です`}
      </p>
    </Card>
  );
}
