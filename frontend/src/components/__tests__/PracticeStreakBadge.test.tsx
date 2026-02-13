import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import PracticeStreakBadge from '../PracticeStreakBadge';

describe('PracticeStreakBadge', () => {
  it('ストリーク日数が表示される', () => {
    render(<PracticeStreakBadge streakDays={5} totalSessions={10} />);

    expect(screen.getByText('5')).toBeInTheDocument();
    // ストリークラベルのテキストを確認
    const allText = document.body.textContent;
    expect(allText).toContain('日連続');
  });

  it('累計練習回数が表示される', () => {
    render(<PracticeStreakBadge streakDays={3} totalSessions={12} />);

    expect(screen.getByText(/累計 12回/)).toBeInTheDocument();
  });

  it('初回練習バッジが達成済みで表示される', () => {
    render(<PracticeStreakBadge streakDays={1} totalSessions={1} />);

    expect(screen.getByText('初回練習')).toBeInTheDocument();
  });

  it('3日連続バッジが未達成の場合グレーで表示される', () => {
    render(<PracticeStreakBadge streakDays={1} totalSessions={1} />);

    const badge = screen.getByText('3日連続');
    expect(badge.closest('div')?.className).toContain('opacity-40');
  });

  it('3日連続バッジが達成済みの場合通常表示される', () => {
    render(<PracticeStreakBadge streakDays={3} totalSessions={5} />);

    const badge = screen.getByText('3日連続');
    expect(badge.closest('div')?.className).not.toContain('opacity-40');
  });

  it('ストリーク0日でも表示される', () => {
    render(<PracticeStreakBadge streakDays={0} totalSessions={0} />);

    expect(screen.getByText('0')).toBeInTheDocument();
  });
});
