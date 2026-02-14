import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ScoreGrowthTrendCard from '../ScoreGrowthTrendCard';

describe('ScoreGrowthTrendCard', () => {
  it('セッションが2件未満の場合は何も表示しない', () => {
    const { container } = render(<ScoreGrowthTrendCard scores={[7.0]} />);
    expect(container.firstChild).toBeNull();
  });

  it('成長トレンドタイトルが表示される', () => {
    render(<ScoreGrowthTrendCard scores={[6.0, 7.0, 8.0]} />);
    expect(screen.getByText('成長トレンド')).toBeInTheDocument();
  });

  it('スコアが上昇傾向の場合に上昇メッセージが表示される', () => {
    render(<ScoreGrowthTrendCard scores={[5.0, 6.0, 7.0, 8.0, 9.0]} />);
    expect(screen.getAllByText(/上昇/).length).toBeGreaterThanOrEqual(1);
  });

  it('スコアが下降傾向の場合に下降メッセージが表示される', () => {
    render(<ScoreGrowthTrendCard scores={[9.0, 8.0, 7.0, 6.0, 5.0]} />);
    expect(screen.getAllByText(/低下/).length).toBeGreaterThanOrEqual(1);
  });

  it('スコアが横ばいの場合に安定メッセージが表示される', () => {
    render(<ScoreGrowthTrendCard scores={[7.0, 7.0, 7.0, 7.0]} />);
    expect(screen.getAllByText(/安定/).length).toBeGreaterThanOrEqual(1);
  });

  it('差分値が表示される', () => {
    // 前半平均: 6.0, 後半平均: 8.0 → 差分 +2.0
    render(<ScoreGrowthTrendCard scores={[6.0, 6.0, 8.0, 8.0]} />);
    expect(screen.getByText(/\+2\.0/)).toBeInTheDocument();
  });

  it('最新スコアが表示される', () => {
    render(<ScoreGrowthTrendCard scores={[5.0, 7.5]} />);
    expect(screen.getByText('7.5')).toBeInTheDocument();
  });
});
