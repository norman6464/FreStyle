import { useState, useCallback, useMemo, useEffect } from 'react';
import { DailyGoalRepository } from '../repositories/DailyGoalRepository';
import type { DailyGoal } from '../types';

export function useDailyGoal() {
  const [goal, setGoal] = useState<DailyGoal>({ date: '', target: 3, completed: 0 });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    DailyGoalRepository.getToday().then((data) => {
      if (!cancelled) {
        setGoal(data);
        setLoading(false);
      }
    });
    return () => { cancelled = true; };
  }, []);

  const setTarget = useCallback(async (target: number) => {
    await DailyGoalRepository.setTarget(target);
    const updated = await DailyGoalRepository.getToday();
    setGoal(updated);
  }, []);

  const incrementCompleted = useCallback(async () => {
    const updated = await DailyGoalRepository.incrementCompleted();
    setGoal(updated);
  }, []);

  const isAchieved = useMemo(() => goal.completed >= goal.target, [goal]);
  const progress = useMemo(() => {
    if (goal.target === 0) return 100;
    return Math.round((goal.completed / goal.target) * 100);
  }, [goal]);

  return { goal, setTarget, incrementCompleted, isAchieved, progress, loading };
}
