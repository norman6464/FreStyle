import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import WeeklyGoalProgressCard from '../WeeklyGoalProgressCard';

describe('WeeklyGoalProgressCard', () => {
  it('タイトルが表示される', () => {
    render(<WeeklyGoalProgressCard sessionsThisWeek={0} weeklyGoal={5} />);

    expect(screen.getByText('今週の練習目標')).toBeInTheDocument();
  });

  it('進捗が正しく表示される', () => {
    render(<WeeklyGoalProgressCard sessionsThisWeek={3} weeklyGoal={5} />);

    expect(screen.getByText('3')).toBeInTheDocument();
    expect(screen.getByText('/ 5 回')).toBeInTheDocument();
  });

  it('0回の時に励ましメッセージが表示される', () => {
    render(<WeeklyGoalProgressCard sessionsThisWeek={0} weeklyGoal={5} />);

    expect(screen.getByText('今週はまだ練習していません。始めてみましょう！')).toBeInTheDocument();
  });

  it('達成途中のメッセージが表示される', () => {
    render(<WeeklyGoalProgressCard sessionsThisWeek={3} weeklyGoal={5} />);

    expect(screen.getByText('いい調子です！あと2回で目標達成！')).toBeInTheDocument();
  });

  it('目標達成時に祝福メッセージが表示される', () => {
    render(<WeeklyGoalProgressCard sessionsThisWeek={5} weeklyGoal={5} />);

    expect(screen.getByText('目標達成！素晴らしいです！')).toBeInTheDocument();
  });

  it('プログレスバーが正しい幅で表示される', () => {
    render(<WeeklyGoalProgressCard sessionsThisWeek={3} weeklyGoal={5} />);

    const bar = document.querySelector('[data-testid="progress-bar"]');
    expect(bar).toHaveStyle({ width: '60%' });
  });
});
