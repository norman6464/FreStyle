import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ScoreCard from '../ScoreCard';
import type { ScoreCard as ScoreCardType } from '../../types';

describe('ScoreCard', () => {
  const scoreCard: ScoreCardType = {
    sessionId: 1,
    scores: [
      { axis: '論理的構成力', score: 8, comment: '良い構成です' },
      { axis: '配慮表現', score: 6, comment: '改善余地あり' },
      { axis: '要約力', score: 7, comment: '簡潔です' },
      { axis: '提案力', score: 5, comment: '要改善' },
      { axis: '質問・傾聴力', score: 9, comment: '優秀' },
    ],
    overallScore: 7.0,
  };

  it('総合スコアが表示される', () => {
    render(<ScoreCard scoreCard={scoreCard} />);

    expect(screen.getByText('7.0')).toBeInTheDocument();
  });

  it('各評価軸の名前が表示される', () => {
    render(<ScoreCard scoreCard={scoreCard} />);

    expect(screen.getByText('論理的構成力')).toBeInTheDocument();
    expect(screen.getByText('配慮表現')).toBeInTheDocument();
    expect(screen.getByText('要約力')).toBeInTheDocument();
    expect(screen.getByText('提案力')).toBeInTheDocument();
    expect(screen.getByText('質問・傾聴力')).toBeInTheDocument();
  });

  it('各評価軸のスコアが表示される', () => {
    render(<ScoreCard scoreCard={scoreCard} />);

    expect(screen.getByText('8')).toBeInTheDocument();
    expect(screen.getByText('6')).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
    expect(screen.getByText('9')).toBeInTheDocument();
  });

  it('各評価軸のコメントが表示される', () => {
    render(<ScoreCard scoreCard={scoreCard} />);

    expect(screen.getByText('良い構成です')).toBeInTheDocument();
    expect(screen.getByText('改善余地あり')).toBeInTheDocument();
    expect(screen.getByText('優秀')).toBeInTheDocument();
  });
});
