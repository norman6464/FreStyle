import type { DailyGoal } from '../types';

const STORAGE_KEY = 'freestyle_daily_goal';
const DEFAULT_TARGET = 3;

function getTodayStr(): string {
  return new Date().toISOString().split('T')[0];
}

export const DailyGoalRepository = {
  getToday(): DailyGoal {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      const goal: DailyGoal = JSON.parse(raw);
      if (goal.date === getTodayStr()) {
        return goal;
      }
    }
    return { date: getTodayStr(), target: DEFAULT_TARGET, completed: 0 };
  },

  setTarget(target: number): void {
    const goal = this.getToday();
    goal.target = target;
    localStorage.setItem(STORAGE_KEY, JSON.stringify(goal));
  },

  incrementCompleted(): void {
    const goal = this.getToday();
    goal.completed += 1;
    localStorage.setItem(STORAGE_KEY, JSON.stringify(goal));
  },
};
