import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ProfileStatsSection from '../ProfileStatsSection';
import type { ProfileStats } from '../../../repositories/ProfileStatsRepository';

const mockStats: ProfileStats = {
  totalSessions: 42,
  averageScore: 78.5,
  currentStreak: 7,
  longestStreak: 14,
  totalAchievedDays: 30,
  followerCount: 5,
  followingCount: 8,
};

describe('ProfileStatsSection', () => {
  it('統計データを表示する', () => {
    render(<ProfileStatsSection stats={mockStats} loading={false} />);

    expect(screen.getByText('学習統計')).toBeInTheDocument();
    expect(screen.getByText('42回')).toBeInTheDocument();
    expect(screen.getByText('78.5点')).toBeInTheDocument();
    expect(screen.getByText('7日')).toBeInTheDocument();
    expect(screen.getByText('14日')).toBeInTheDocument();
    expect(screen.getByText('5人')).toBeInTheDocument();
    expect(screen.getByText('8人')).toBeInTheDocument();
  });

  it('平均スコアが0の場合「未計測」を表示する', () => {
    render(<ProfileStatsSection stats={{ ...mockStats, averageScore: 0 }} loading={false} />);

    expect(screen.getByText('未計測')).toBeInTheDocument();
  });

  it('ローディング中はスケルトンを表示する', () => {
    render(<ProfileStatsSection stats={null} loading={true} />);

    expect(screen.getByText('学習統計')).toBeInTheDocument();
    expect(screen.queryByText('42回')).not.toBeInTheDocument();
  });

  it('statsがnullでloadingがfalseの場合は何も表示しない', () => {
    const { container } = render(<ProfileStatsSection stats={null} loading={false} />);

    expect(container.innerHTML).toBe('');
  });
});
