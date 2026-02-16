import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ScoreGoalCard from '../ScoreGoalCard';

const mockSaveGoal = vi.fn();
let mockGoal = 8.0;

vi.mock('../../hooks/useScoreGoal', () => ({
  useScoreGoal: () => ({
    goal: mockGoal,
    saveGoal: mockSaveGoal,
    loading: false,
  }),
}));

describe('ScoreGoalCard', () => {
  beforeEach(() => {
    mockGoal = 8.0;
    mockSaveGoal.mockClear();
  });

  it('平均スコアが表示される', () => {
    render(<ScoreGoalCard averageScore={7.2} />);
    expect(screen.getByText('7.2')).toBeInTheDocument();
  });

  it('デフォルト目標スコア8.0が目標欄に表示される', () => {
    render(<ScoreGoalCard averageScore={7.0} />);
    const goalDisplay = screen.getByText('8.0', { selector: '.text-primary-400' });
    expect(goalDisplay).toBeInTheDocument();
  });

  it('目標達成時に達成メッセージが表示される', () => {
    render(<ScoreGoalCard averageScore={8.5} />);
    expect(screen.getByText(/目標達成/)).toBeInTheDocument();
  });

  it('目標未達成時に励ましメッセージが表示される', () => {
    render(<ScoreGoalCard averageScore={6.0} />);
    expect(screen.getByText(/あと/)).toBeInTheDocument();
  });

  it('目標スコアを変更するとsaveGoalが呼ばれる', () => {
    render(<ScoreGoalCard averageScore={7.0} />);
    const select = screen.getByRole('combobox');
    fireEvent.change(select, { target: { value: '9' } });
    expect(mockSaveGoal).toHaveBeenCalledWith(9);
  });
});
