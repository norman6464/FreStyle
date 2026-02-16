import Card from './Card';
import ProgressBar from './ProgressBar';
import { useScoreGoal } from '../hooks/useScoreGoal';

interface Props {
  averageScore: number;
}

const GOAL_OPTIONS = [6.0, 6.5, 7.0, 7.5, 8.0, 8.5, 9.0, 9.5, 10.0];

export default function ScoreGoalCard({ averageScore }: Props) {
  const { goal, saveGoal } = useScoreGoal();

  const handleGoalChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    saveGoal(parseFloat(e.target.value));
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

      <div className="mb-2">
        <ProgressBar
          percentage={progress}
          barColorClass={achieved ? 'bg-emerald-500' : 'bg-primary-500'}
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
