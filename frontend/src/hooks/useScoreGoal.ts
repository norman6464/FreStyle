import { useState, useCallback, useEffect } from 'react';
import { ScoreGoalRepository } from '../repositories/ScoreGoalRepository';

const STORAGE_KEY = 'scoreGoal';
const DEFAULT_GOAL = 8.0;

function getLocalGoal(): number {
  try {
    const item = localStorage.getItem(STORAGE_KEY);
    return item !== null ? JSON.parse(item) : DEFAULT_GOAL;
  } catch {
    return DEFAULT_GOAL;
  }
}

function setLocalGoal(goal: number): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(goal));
  } catch {
    // QuotaExceededError等
  }
}

export function useScoreGoal() {
  const [goal, setGoal] = useState<number>(getLocalGoal());
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    ScoreGoalRepository.fetchGoal()
      .then((goalScore) => {
        if (!cancelled && goalScore != null) {
          setGoal(goalScore);
          setLocalGoal(goalScore);
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, []);

  const saveGoal = useCallback(async (newGoal: number) => {
    setGoal(newGoal);
    setLocalGoal(newGoal);
    await ScoreGoalRepository.saveGoal(newGoal);
  }, []);

  return { goal, saveGoal, loading };
}
