import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ScoreGoalCard from '../ScoreGoalCard';

function createMockStorage(): Storage {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
    removeItem: vi.fn((key: string) => { delete store[key]; }),
    clear: vi.fn(() => { store = {}; }),
    get length() { return Object.keys(store).length; },
    key: vi.fn((index: number) => Object.keys(store)[index] ?? null),
  };
}

describe('ScoreGoalCard', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
  });

  it('平均スコアが表示される', () => {
    render(<ScoreGoalCard averageScore={7.2} />);
    expect(screen.getByText('7.2')).toBeInTheDocument();
  });

  it('デフォルト目標スコア8.0が目標欄に表示される', () => {
    render(<ScoreGoalCard averageScore={7.0} />);
    // 目標欄のテキストを確認（selectのoptionにも8.0があるのでtext-primary-400で絞る）
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

  it('目標スコアを変更できる', () => {
    render(<ScoreGoalCard averageScore={7.0} />);
    const select = screen.getByRole('combobox');
    fireEvent.change(select, { target: { value: '9' } });
    const goalDisplay = screen.getByText('9.0', { selector: '.text-primary-400' });
    expect(goalDisplay).toBeInTheDocument();
  });

  it('目標スコアがLocalStorageに保存される', () => {
    render(<ScoreGoalCard averageScore={7.0} />);
    const select = screen.getByRole('combobox');
    fireEvent.change(select, { target: { value: '9' } });
    expect(localStorage.setItem).toHaveBeenCalledWith('scoreGoal', '9');
  });

  it('LocalStorageから目標スコアを復元する', () => {
    (localStorage.getItem as ReturnType<typeof vi.fn>).mockReturnValue('9.5');
    render(<ScoreGoalCard averageScore={6.0} />);
    const goalDisplay = screen.getByText('9.5', { selector: '.text-primary-400' });
    expect(goalDisplay).toBeInTheDocument();
  });
});
