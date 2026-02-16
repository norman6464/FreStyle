import apiClient from '../lib/axios';
import type { DailyGoal } from '../types';

const STORAGE_KEY = 'freestyle_daily_goal';
const DEFAULT_TARGET = 3;

function getTodayStr(): string {
  return new Date().toISOString().split('T')[0];
}

function getLocalGoal(): DailyGoal {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (raw) {
    try {
      const goal: DailyGoal = JSON.parse(raw);
      if (goal.date === getTodayStr()) {
        return goal;
      }
    } catch {
      localStorage.removeItem(STORAGE_KEY);
    }
  }
  return { date: getTodayStr(), target: DEFAULT_TARGET, completed: 0 };
}

function saveLocalGoal(goal: DailyGoal): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(goal));
}

export const DailyGoalRepository = {
  async getToday(): Promise<DailyGoal> {
    try {
      const response = await apiClient.get<DailyGoal>('/api/daily-goals/today');
      return response.data;
    } catch {
      return getLocalGoal();
    }
  },

  async setTarget(target: number): Promise<void> {
    try {
      await apiClient.put('/api/daily-goals/target', { target });
    } catch {
      const goal = getLocalGoal();
      goal.target = target;
      saveLocalGoal(goal);
    }
  },

  async incrementCompleted(): Promise<DailyGoal> {
    try {
      const response = await apiClient.post<DailyGoal>('/api/daily-goals/increment');
      return response.data;
    } catch {
      const goal = getLocalGoal();
      goal.completed += 1;
      saveLocalGoal(goal);
      return goal;
    }
  },
};
