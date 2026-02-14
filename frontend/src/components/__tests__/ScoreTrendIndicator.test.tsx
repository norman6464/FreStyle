import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import ScoreTrendIndicator from '../ScoreTrendIndicator';

describe('ScoreTrendIndicator', () => {
  it('上昇トレンドのメッセージが表示される', () => {
    render(<ScoreTrendIndicator scores={[5.0, 6.0, 7.0, 8.0, 9.0]} />);
    expect(screen.getByText('上昇傾向')).toBeInTheDocument();
  });

  it('下降トレンドのメッセージが表示される', () => {
    render(<ScoreTrendIndicator scores={[9.0, 8.0, 7.0, 6.0, 5.0]} />);
    expect(screen.getByText('下降傾向')).toBeInTheDocument();
  });

  it('横ばいトレンドのメッセージが表示される', () => {
    render(<ScoreTrendIndicator scores={[7.0, 7.0, 7.0, 7.0, 7.0]} />);
    expect(screen.getByText('安定')).toBeInTheDocument();
  });

  it('スコアが1件以下の場合は表示されない', () => {
    const { container } = render(<ScoreTrendIndicator scores={[7.0]} />);
    expect(container.firstChild).toBeNull();
  });

  it('スコアが空の場合は表示されない', () => {
    const { container } = render(<ScoreTrendIndicator scores={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('2件のスコアでもトレンドが計算される', () => {
    render(<ScoreTrendIndicator scores={[5.0, 8.0]} />);
    expect(screen.getByText('上昇傾向')).toBeInTheDocument();
  });

  it('トレンドラベルが表示される', () => {
    render(<ScoreTrendIndicator scores={[5.0, 6.0, 7.0]} />);
    expect(screen.getByText('直近トレンド')).toBeInTheDocument();
  });

  it('アイコンが表示される', () => {
    const { container } = render(<ScoreTrendIndicator scores={[5.0, 6.0, 7.0]} />);
    const svg = container.querySelector('svg');
    expect(svg).toBeTruthy();
  });
});
