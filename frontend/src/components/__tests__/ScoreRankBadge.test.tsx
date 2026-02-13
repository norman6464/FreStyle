import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ScoreRankBadge from '../ScoreRankBadge';

describe('ScoreRankBadge', () => {
  it('スコア9.0以上でSランクが表示される', () => {
    render(<ScoreRankBadge score={9.5} />);

    expect(screen.getByText('S')).toBeInTheDocument();
    expect(screen.getByText('エキスパート')).toBeInTheDocument();
  });

  it('スコア8.0以上でAランクが表示される', () => {
    render(<ScoreRankBadge score={8.3} />);

    expect(screen.getByText('A')).toBeInTheDocument();
    expect(screen.getByText('上級')).toBeInTheDocument();
  });

  it('スコア7.0以上でBランクが表示される', () => {
    render(<ScoreRankBadge score={7.0} />);

    expect(screen.getByText('B')).toBeInTheDocument();
    expect(screen.getByText('中級')).toBeInTheDocument();
  });

  it('スコア6.0以上でCランクが表示される', () => {
    render(<ScoreRankBadge score={6.5} />);

    expect(screen.getByText('C')).toBeInTheDocument();
    expect(screen.getByText('初級')).toBeInTheDocument();
  });

  it('スコア6.0未満でDランクが表示される', () => {
    render(<ScoreRankBadge score={4.0} />);

    expect(screen.getByText('D')).toBeInTheDocument();
    expect(screen.getByText('入門')).toBeInTheDocument();
  });
});
