import { useState } from 'react';

interface Props {
  averageScore: number;
}

const GOAL_OPTIONS = [6.0, 6.5, 7.0, 7.5, 8.0, 8.5, 9.0, 9.5, 10.0];
const STORAGE_KEY = 'scoreGoal';

function getStoredGoal(): number {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored) {
    const parsed = parseFloat(stored);
    if (!isNaN(parsed)) return parsed;
  }
  return 8.0;
}

export default function ScoreGoalCard({ averageScore }: Props) {
  const [goal, setGoal] = useState(getStoredGoal);

  const handleGoalChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newGoal = parseFloat(e.target.value);
    setGoal(newGoal);
    localStorage.setItem(STORAGE_KEY, String(newGoal));
  };

  const achieved = averageScore >= goal;
  const gap = Math.round((goal - averageScore) * 10) / 10;
  const progress = Math.min((averageScore / goal) * 100, 100);

  return (
    <div className="bg-white rounded-lg border border-slate-200 p-4">
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-medium text-slate-700">目標スコア</p>
        <select
          value={goal}
          onChange={handleGoalChange}
          className="text-xs border border-slate-200 rounded px-2 py-1 text-slate-600"
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
          <p className="text-xs text-slate-500">現在の平均</p>
          <p className="text-xl font-bold text-slate-800">{averageScore.toFixed(1)}</p>
        </div>
        <div className="text-right">
          <p className="text-xs text-slate-500">目標</p>
          <p className="text-xl font-bold text-primary-600">{goal.toFixed(1)}</p>
        </div>
      </div>

      <div className="w-full bg-slate-100 rounded-full h-2 mb-2">
        <div
          className={`h-2 rounded-full transition-all ${achieved ? 'bg-emerald-500' : 'bg-primary-500'}`}
          style={{ width: `${progress}%` }}
        />
      </div>

      <p className={`text-xs ${achieved ? 'text-emerald-600 font-medium' : 'text-slate-500'}`}>
        {achieved
          ? '目標達成！素晴らしいです！'
          : `あと ${gap.toFixed(1)} ポイントで目標達成です`}
      </p>
    </div>
  );
}
