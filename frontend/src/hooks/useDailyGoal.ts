import { useState, useCallback, useMemo } from 'react';
import { DailyGoalRepository } from '../repositories/DailyGoalRepository';
import type { DailyGoal } from '../types';

export function useDailyGoal() {
  const [goal, setGoal] = useState<DailyGoal>(() => DailyGoalRepository.getToday());

  const setTarget = useCallback((target: number) => {
    DailyGoalRepository.setTarget(target);
    setGoal(DailyGoalRepository.getToday());
  }, []);

  const incrementCompleted = useCallback(() => {
    DailyGoalRepository.incrementCompleted();
    setGoal(DailyGoalRepository.getToday());
  }, []);

  const isAchieved = useMemo(() => goal.completed >= goal.target, [goal]);
  const progress = useMemo(() => {
    if (goal.target === 0) return 100;
    return Math.round((goal.completed / goal.target) * 100);
  }, [goal]);

  return { goal, setTarget, incrementCompleted, isAchieved, progress };
}
