import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import ScoreFilterSummary from '../ScoreFilterSummary';

describe('ScoreFilterSummary', () => {
  it('セッション数が表示される', () => {
    render(<ScoreFilterSummary scores={[7.5, 8.0, 6.5]} />);
    expect(screen.getByText('3')).toBeInTheDocument();
  });

  it('平均スコアが表示される', () => {
    render(<ScoreFilterSummary scores={[7.0, 8.0, 9.0]} />);
    expect(screen.getByText('8.0')).toBeInTheDocument();
  });

  it('最高スコアが表示される', () => {
    render(<ScoreFilterSummary scores={[5.0, 8.5, 7.0]} />);
    expect(screen.getByText('8.5')).toBeInTheDocument();
  });

  it('最低スコアが表示される', () => {
    render(<ScoreFilterSummary scores={[5.0, 8.5, 7.0]} />);
    expect(screen.getByText('5.0')).toBeInTheDocument();
  });

  it('スコアが空の場合は表示されない', () => {
    const { container } = render(<ScoreFilterSummary scores={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('単一スコアの場合も正しく表示される', () => {
    render(<ScoreFilterSummary scores={[7.5]} />);
    expect(screen.getByText('1')).toBeInTheDocument();
    expect(screen.getAllByText('7.5')).toHaveLength(3);
  });

  it('ラベルが表示される', () => {
    render(<ScoreFilterSummary scores={[7.0, 8.0]} />);
    expect(screen.getByText('件数')).toBeInTheDocument();
    expect(screen.getByText('平均')).toBeInTheDocument();
    expect(screen.getByText('最高')).toBeInTheDocument();
    expect(screen.getByText('最低')).toBeInTheDocument();
  });
});
