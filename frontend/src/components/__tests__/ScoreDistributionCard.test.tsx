import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ScoreDistributionCard from '../ScoreDistributionCard';

describe('ScoreDistributionCard', () => {
  it('タイトルが表示される', () => {
    render(<ScoreDistributionCard scores={[5, 7, 8]} />);
    expect(screen.getByText('スコア分布')).toBeInTheDocument();
  });

  it('4つのレンジラベルが表示される', () => {
    render(<ScoreDistributionCard scores={[5, 7, 8]} />);
    expect(screen.getByText('9-10')).toBeInTheDocument();
    expect(screen.getByText('7-8')).toBeInTheDocument();
    expect(screen.getByText('4-6')).toBeInTheDocument();
    expect(screen.getByText('1-3')).toBeInTheDocument();
  });

  it('各レンジの件数が表示される', () => {
    render(<ScoreDistributionCard scores={[2, 5, 7, 8, 9]} />);
    // 1-3: 1件, 4-6: 1件, 7-8: 2件, 9-10: 1件
    const counts = screen.getAllByTestId('range-count');
    expect(counts[0]).toHaveTextContent('1'); // 9-10
    expect(counts[1]).toHaveTextContent('2'); // 7-8
    expect(counts[2]).toHaveTextContent('1'); // 4-6
    expect(counts[3]).toHaveTextContent('1'); // 1-3
  });

  it('最頻レンジに応じたメッセージが表示される（高スコア帯）', () => {
    render(<ScoreDistributionCard scores={[9, 9.5, 10, 8]} />);
    expect(screen.getByText(/高スコア帯/)).toBeInTheDocument();
  });

  it('最頻レンジに応じたメッセージが表示される（中スコア帯）', () => {
    render(<ScoreDistributionCard scores={[7, 7.5, 8, 5]} />);
    expect(screen.getByText(/安定した実力/)).toBeInTheDocument();
  });

  it('最頻レンジに応じたメッセージが表示される（低スコア帯）', () => {
    render(<ScoreDistributionCard scores={[2, 3, 1, 5]} />);
    expect(screen.getByText(/伸びしろ/)).toBeInTheDocument();
  });

  it('スコアが空の場合は何も表示しない', () => {
    const { container } = render(<ScoreDistributionCard scores={[]} />);
    expect(container.firstChild).toBeNull();
  });
});
