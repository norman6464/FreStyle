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

  it('総合スコアのレベルラベルが表示される（実務レベル）', () => {
    render(<ScoreCard scoreCard={scoreCard} />);

    expect(screen.getByText('実務レベル')).toBeInTheDocument();
  });

  it('総合スコア8以上で「優秀」レベルが表示される', () => {
    const excellentCard: ScoreCardType = {
      ...scoreCard,
      overallScore: 8.5,
    };
    render(<ScoreCard scoreCard={excellentCard} />);

    expect(screen.getByText('優秀レベル')).toBeInTheDocument();
  });

  it('総合スコア5未満で「基礎レベル」が表示される', () => {
    const basicCard: ScoreCardType = {
      ...scoreCard,
      overallScore: 4.0,
    };
    render(<ScoreCard scoreCard={basicCard} />);

    expect(screen.getByText('基礎レベル')).toBeInTheDocument();
  });

  it('低スコアの軸に改善ヒントが表示される', () => {
    render(<ScoreCard scoreCard={scoreCard} />);

    // スコア5の「提案力」に改善ヒントが表示される
    expect(screen.getByText(/この項目を重点的に練習しましょう/)).toBeInTheDocument();
  });

  it('高スコアの軸には改善ヒントが表示されない', () => {
    const allHighCard: ScoreCardType = {
      sessionId: 1,
      scores: [
        { axis: '論理的構成力', score: 9, comment: '素晴らしい' },
      ],
      overallScore: 9.0,
    };
    render(<ScoreCard scoreCard={allHighCard} />);

    expect(screen.queryByText(/この項目を重点的に練習しましょう/)).not.toBeInTheDocument();
  });

  it('scoresがnullでもエラーにならない', () => {
    const nullScoresCard: ScoreCardType = {
      sessionId: 1,
      scores: null as unknown as ScoreCardType['scores'],
      overallScore: 7.0,
    };
    render(<ScoreCard scoreCard={nullScoresCard} />);
    expect(screen.getByText('スコアカード')).toBeInTheDocument();
  });

  it('スコアに応じたプログレスバーの色分けがされる', () => {
    render(<ScoreCard scoreCard={scoreCard} />);

    // スコア9の軸 → emerald（高スコア）
    const bars = document.querySelectorAll('[class*="rounded-full"][style]');
    const highScoreBar = Array.from(bars).find(
      (bar) => (bar as HTMLElement).style.width === '90%'
    );
    expect(highScoreBar?.className).toContain('bg-emerald');

    // スコア5の軸 → rose（低スコア）
    const lowScoreBar = Array.from(bars).find(
      (bar) => (bar as HTMLElement).style.width === '50%'
    );
    expect(lowScoreBar?.className).toContain('bg-rose');
  });
});
