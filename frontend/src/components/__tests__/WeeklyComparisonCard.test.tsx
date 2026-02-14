import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import WeeklyComparisonCard from '../WeeklyComparisonCard';

describe('WeeklyComparisonCard', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    // 2026-02-13 (金) に固定
    vi.setSystemTime(new Date('2026-02-13T12:00:00'));
  });

  it('タイトルが表示される', () => {
    render(<WeeklyComparisonCard sessions={[]} />);
    expect(screen.getByText('週間比較')).toBeInTheDocument();
  });

  it('今週と先週のラベルが表示される', () => {
    render(<WeeklyComparisonCard sessions={[]} />);
    expect(screen.getByText('今週')).toBeInTheDocument();
    expect(screen.getByText('先週')).toBeInTheDocument();
  });

  it('今週のセッション数と平均スコアが表示される', () => {
    const sessions = [
      { overallScore: 8.0, createdAt: '2026-02-10T10:00:00' }, // 火 (今週)
      { overallScore: 6.0, createdAt: '2026-02-11T10:00:00' }, // 水 (今週)
    ];
    render(<WeeklyComparisonCard sessions={sessions} />);
    const counts = screen.getAllByTestId('session-count');
    expect(counts[0]).toHaveTextContent('2');
    const averages = screen.getAllByTestId('avg-score');
    expect(averages[0]).toHaveTextContent('7.0');
  });

  it('先週のセッション数と平均スコアが表示される', () => {
    const sessions = [
      { overallScore: 7.0, createdAt: '2026-02-03T10:00:00' }, // 先週火
      { overallScore: 5.0, createdAt: '2026-02-04T10:00:00' }, // 先週水
      { overallScore: 9.0, createdAt: '2026-02-05T10:00:00' }, // 先週木
    ];
    render(<WeeklyComparisonCard sessions={sessions} />);
    const counts = screen.getAllByTestId('session-count');
    expect(counts[1]).toHaveTextContent('3');
    const averages = screen.getAllByTestId('avg-score');
    expect(averages[1]).toHaveTextContent('7.0');
  });

  it('スコア上昇時にプラス表示される', () => {
    const sessions = [
      { overallScore: 5.0, createdAt: '2026-02-03T10:00:00' }, // 先週
      { overallScore: 8.0, createdAt: '2026-02-10T10:00:00' }, // 今週
    ];
    render(<WeeklyComparisonCard sessions={sessions} />);
    expect(screen.getByTestId('score-delta')).toHaveTextContent('+3.0');
  });

  it('セッションがない場合でもクラッシュしない', () => {
    render(<WeeklyComparisonCard sessions={[]} />);
    const counts = screen.getAllByTestId('session-count');
    expect(counts[0]).toHaveTextContent('0');
    expect(counts[1]).toHaveTextContent('0');
  });

  it('今週のみセッションがある場合に変動値が表示されない', () => {
    const sessions = [
      { overallScore: 8.0, createdAt: '2026-02-10T10:00:00' },
    ];
    render(<WeeklyComparisonCard sessions={sessions} />);
    expect(screen.queryByTestId('score-delta')).not.toBeInTheDocument();
  });
});
