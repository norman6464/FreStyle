import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ScoreComparisonCard from '../ScoreComparisonCard';

const firstScores = [
  { axis: '論理的構成力', score: 5, comment: '' },
  { axis: '配慮表現', score: 4, comment: '' },
  { axis: '要約力', score: 6, comment: '' },
];

const latestScores = [
  { axis: '論理的構成力', score: 8, comment: '' },
  { axis: '配慮表現', score: 7, comment: '' },
  { axis: '要約力', score: 6, comment: '' },
];

describe('ScoreComparisonCard', () => {
  it('タイトルが表示される', () => {
    render(<ScoreComparisonCard firstScores={firstScores} latestScores={latestScores} firstOverall={5.0} latestOverall={7.0} />);

    expect(screen.getByText('成長の記録')).toBeInTheDocument();
  });

  it('初回と最新の総合スコアが表示される', () => {
    render(<ScoreComparisonCard firstScores={firstScores} latestScores={latestScores} firstOverall={5.0} latestOverall={7.0} />);

    expect(screen.getByText('5.0')).toBeInTheDocument();
    expect(screen.getByText('7.0')).toBeInTheDocument();
  });

  it('スコア増加時にプラス表示される', () => {
    render(<ScoreComparisonCard firstScores={firstScores} latestScores={latestScores} firstOverall={5.0} latestOverall={7.0} />);

    expect(screen.getByText('+2.0')).toBeInTheDocument();
  });

  it('各軸の変化量が表示される', () => {
    render(<ScoreComparisonCard firstScores={firstScores} latestScores={latestScores} firstOverall={5.0} latestOverall={7.0} />);

    const plusThreeElements = screen.getAllByText('+3.0');
    expect(plusThreeElements.length).toBe(2);
  });

  it('変化なしの軸は±0と表示される', () => {
    render(<ScoreComparisonCard firstScores={firstScores} latestScores={latestScores} firstOverall={5.0} latestOverall={7.0} />);

    expect(screen.getByText('±0')).toBeInTheDocument();
  });
});
