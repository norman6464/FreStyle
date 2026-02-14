import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import AchievementBadgeCard from '../AchievementBadgeCard';

describe('AchievementBadgeCard', () => {
  it('タイトルが表示される', () => {
    render(<AchievementBadgeCard totalSessions={0} />);

    expect(screen.getByText('達成バッジ')).toBeInTheDocument();
  });

  it('セッション0回で全バッジが未達成', () => {
    render(<AchievementBadgeCard totalSessions={0} />);

    const unlocked = screen.queryAllByTestId('badge-unlocked');
    expect(unlocked).toHaveLength(0);
  });

  it('セッション1回でファーストステップバッジが達成', () => {
    render(<AchievementBadgeCard totalSessions={1} />);

    const unlocked = screen.getAllByTestId('badge-unlocked');
    expect(unlocked.length).toBeGreaterThanOrEqual(1);
  });

  it('セッション50回で複数バッジが達成', () => {
    render(<AchievementBadgeCard totalSessions={50} />);

    const unlocked = screen.getAllByTestId('badge-unlocked');
    expect(unlocked.length).toBeGreaterThanOrEqual(4);
  });

  it('次のバッジまでの情報が表示される', () => {
    render(<AchievementBadgeCard totalSessions={3} />);

    expect(screen.getByTestId('next-badge-info')).toBeInTheDocument();
  });
});
