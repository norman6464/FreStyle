import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import DailyGoalCard from '../DailyGoalCard';

vi.mock('../../hooks/useDailyGoal', () => ({
  useDailyGoal: () => mockGoalState,
}));

let mockGoalState = {
  goal: { date: '2026-02-13', target: 3, completed: 1 },
  isAchieved: false,
  progress: 33,
  setTarget: vi.fn(),
  incrementCompleted: vi.fn(),
};

describe('DailyGoalCard', () => {
  it('今日の目標と進捗が表示される', () => {
    render(<DailyGoalCard />);

    expect(screen.getByText('今日の目標')).toBeInTheDocument();
    expect(screen.getByText(/1/)).toBeInTheDocument();
    expect(screen.getByText(/3/)).toBeInTheDocument();
  });

  it('プログレスバーが表示される', () => {
    const { container } = render(<DailyGoalCard />);

    const progressBar = container.querySelector('[role="progressbar"]');
    expect(progressBar).toBeInTheDocument();
  });

  it('目標達成時にメッセージが表示される', () => {
    mockGoalState = {
      goal: { date: '2026-02-13', target: 3, completed: 3 },
      isAchieved: true,
      progress: 100,
      setTarget: vi.fn(),
      incrementCompleted: vi.fn(),
    };

    render(<DailyGoalCard />);

    expect(screen.getByText(/達成/)).toBeInTheDocument();

    // 元に戻す
    mockGoalState = {
      goal: { date: '2026-02-13', target: 3, completed: 1 },
      isAchieved: false,
      progress: 33,
      setTarget: vi.fn(),
      incrementCompleted: vi.fn(),
    };
  });

  it('未達成時に達成メッセージが表示されない', () => {
    render(<DailyGoalCard />);
    expect(screen.queryByText(/達成/)).toBeNull();
  });

  it('プログレスバーにaria属性が設定される', () => {
    const { container } = render(<DailyGoalCard />);

    const progressBar = container.querySelector('[role="progressbar"]');
    expect(progressBar).toHaveAttribute('aria-valuenow', '33');
    expect(progressBar).toHaveAttribute('aria-valuemin', '0');
    expect(progressBar).toHaveAttribute('aria-valuemax', '100');
  });

  it('回ラベルが表示される', () => {
    render(<DailyGoalCard />);
    expect(screen.getByText(/回/)).toBeInTheDocument();
  });
});
