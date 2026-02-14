import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ScoreStatsSummary from '../ScoreStatsSummary';

const mockHistory = [
  { sessionId: 1, overallScore: 8.5 },
  { sessionId: 2, overallScore: 7.0 },
  { sessionId: 3, overallScore: 9.2 },
];

describe('ScoreStatsSummary', () => {
  it('総セッション数が表示される', () => {
    render(<ScoreStatsSummary history={mockHistory} />);

    expect(screen.getByText('3')).toBeInTheDocument();
    expect(screen.getByText('総セッション')).toBeInTheDocument();
  });

  it('平均スコアが表示される', () => {
    render(<ScoreStatsSummary history={mockHistory} />);

    expect(screen.getByText('8.2')).toBeInTheDocument();
    expect(screen.getByText('平均スコア')).toBeInTheDocument();
  });

  it('最高スコアが表示される', () => {
    render(<ScoreStatsSummary history={mockHistory} />);

    expect(screen.getByText('9.2')).toBeInTheDocument();
    expect(screen.getByText('最高スコア')).toBeInTheDocument();
  });

  it('セッションがない場合は何も表示しない', () => {
    const { container } = render(<ScoreStatsSummary history={[]} />);

    expect(container.firstChild).toBeNull();
  });

  it('3カラムグリッドで表示される', () => {
    const { container } = render(<ScoreStatsSummary history={mockHistory} />);
    const grid = container.firstElementChild as HTMLElement;
    expect(grid.className).toContain('grid-cols-3');
  });

  it('最高スコアがアンバー色で表示される', () => {
    render(<ScoreStatsSummary history={mockHistory} />);
    const bestScore = screen.getByText('9.2');
    expect(bestScore.className).toContain('text-amber-400');
  });
});
