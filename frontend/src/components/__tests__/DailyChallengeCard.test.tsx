import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import DailyChallengeCard from '../DailyChallengeCard';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

describe('DailyChallengeCard', () => {
  it('タイトルが表示される', () => {
    render(<DailyChallengeCard />);

    expect(screen.getByText('本日のチャレンジ')).toBeInTheDocument();
  });

  it('チャレンジの内容が表示される', () => {
    render(<DailyChallengeCard />);

    const description = screen.getByTestId('challenge-description');
    expect(description.textContent).toBeTruthy();
  });

  it('難易度が表示される', () => {
    render(<DailyChallengeCard />);

    const difficulty = screen.getByTestId('challenge-difficulty');
    expect(difficulty.textContent).toMatch(/初級|中級|上級/);
  });

  it('チャレンジボタンが表示される', () => {
    render(<DailyChallengeCard />);

    expect(screen.getByText('チャレンジする')).toBeInTheDocument();
  });

  it('チャレンジカテゴリが表示される', () => {
    render(<DailyChallengeCard />);

    const category = screen.getByTestId('challenge-category');
    expect(category.textContent).toBeTruthy();
  });
});
