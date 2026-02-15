import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ScoreSparkline from '../ScoreSparkline';

describe('ScoreSparkline', () => {
  it('スコアが2件以上の場合にSVGグラフが描画される', () => {
    const { container } = render(<ScoreSparkline scores={[6.0, 7.0, 8.0]} />);

    const svg = container.querySelector('svg');
    expect(svg).toBeTruthy();
  });

  it('スコアが1件以下の場合はグラフが描画されない', () => {
    const { container } = render(<ScoreSparkline scores={[7.0]} />);

    const svg = container.querySelector('svg');
    expect(svg).toBeNull();
    expect(screen.getByText('データ不足')).toBeInTheDocument();
  });

  it('最新スコアの数値が表示される', () => {
    render(<ScoreSparkline scores={[5.0, 6.5, 8.2]} />);

    expect(screen.getByText('8.2')).toBeInTheDocument();
  });

  it('上昇トレンドの場合に上向きインジケーターが表示される', () => {
    const { container } = render(<ScoreSparkline scores={[5.0, 6.0, 7.0, 8.0]} />);

    const trendIndicator = container.querySelector('[data-testid="trend-up"]');
    expect(trendIndicator).toBeTruthy();
  });

  it('下降トレンドの場合に下向きインジケーターが表示される', () => {
    const { container } = render(<ScoreSparkline scores={[8.0, 7.0, 6.0, 5.0]} />);

    const trendIndicator = container.querySelector('[data-testid="trend-down"]');
    expect(trendIndicator).toBeTruthy();
  });

  it('同じスコアが連続する場合はトレンドインジケーターが表示されない', () => {
    const { container } = render(<ScoreSparkline scores={[7.0, 7.0, 7.0]} />);

    expect(container.querySelector('[data-testid="trend-up"]')).toBeNull();
    expect(container.querySelector('[data-testid="trend-down"]')).toBeNull();
  });

  it('5件を超えるスコアの場合は直近5件のみ表示される', () => {
    const { container } = render(<ScoreSparkline scores={[3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0]} />);

    const circles = container.querySelectorAll('circle');
    expect(circles.length).toBe(5);
  });

  it('空配列の場合はデータ不足と表示される', () => {
    const { container } = render(<ScoreSparkline scores={[]} />);

    expect(screen.getByText('データ不足')).toBeInTheDocument();
    expect(container.querySelector('svg')).toBeNull();
  });
});
