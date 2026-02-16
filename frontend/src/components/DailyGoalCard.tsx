import { useState, useRef, useEffect } from 'react';
import { useDailyGoal } from '../hooks/useDailyGoal';
import Card from './Card';
import ProgressBar from './ProgressBar';

export default function DailyGoalCard() {
  const { goal, setTarget, isAchieved, progress } = useDailyGoal();
  const [editing, setEditing] = useState(false);
  const [editValue, setEditValue] = useState(goal.target);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [editing]);

  const startEditing = () => {
    setEditValue(goal.target);
    setEditing(true);
  };

  const save = () => {
    const clamped = Math.max(1, Math.min(10, editValue));
    if (clamped !== goal.target) {
      setTarget(clamped);
    }
    setEditing(false);
  };

  const cancel = () => {
    setEditing(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      save();
    } else if (e.key === 'Escape') {
      cancel();
    }
  };

  return (
    <Card>
      <div className="flex items-center justify-between mb-2">
        <p className="text-xs font-medium text-[var(--color-text-secondary)]">今日の目標</p>
        <span className="text-xs text-[var(--color-text-muted)]">
          <span className="font-semibold text-[var(--color-text-primary)]">{goal.completed}</span>
          {' / '}
          {editing ? (
            <input
              ref={inputRef}
              type="number"
              min={1}
              max={10}
              value={editValue}
              onChange={(e) => setEditValue(Number(e.target.value))}
              onKeyDown={handleKeyDown}
              onBlur={save}
              className="w-10 text-center text-xs bg-surface-2 border border-primary-500 rounded px-1 py-0.5 text-[var(--color-text-primary)] outline-none"
            />
          ) : (
            <button
              onClick={startEditing}
              aria-label="目標を変更"
              className="hover:text-primary-400 transition-colors cursor-pointer"
            >
              {goal.target}
            </button>
          )}
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
