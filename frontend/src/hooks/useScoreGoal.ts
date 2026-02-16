import { useState, useCallback, useEffect } from 'react';
import apiClient from '../lib/axios';

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
    apiClient.get<{ goalScore: number }>('/api/score-goal')
      .then((response) => {
        if (!cancelled && response.data?.goalScore != null) {
          setGoal(response.data.goalScore);
          setLocalGoal(response.data.goalScore);
        }
      })
      .catch(() => {
        // API失敗時はlocalStorageの値を維持
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, []);

  const saveGoal = useCallback(async (newGoal: number) => {
    setGoal(newGoal);
    setLocalGoal(newGoal);
    try {
      await apiClient.put('/api/score-goal', { goalScore: newGoal });
    } catch {
      // API失敗時もlocalStorageには保存済み
    }
  }, []);

  return { goal, saveGoal, loading };
}
