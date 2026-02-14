import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import StreakCalendarCard from '../StreakCalendarCard';

describe('StreakCalendarCard', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2025-06-15'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('タイトルが表示される', () => {
    render(<StreakCalendarCard practiceDates={[]} />);
    expect(screen.getByText('練習カレンダー')).toBeInTheDocument();
  });

  it('練習日がない場合、連続日数が0と表示される', () => {
    render(<StreakCalendarCard practiceDates={[]} />);
    expect(screen.getByText('0')).toBeInTheDocument();
    expect(screen.getByText('日連続')).toBeInTheDocument();
  });

  it('28日分のセルが表示される', () => {
    render(<StreakCalendarCard practiceDates={[]} />);
    const cells = screen.getAllByTestId(/^calendar-cell-/);
    expect(cells).toHaveLength(28);
  });

  it('練習した日のセルがアクティブスタイルになる', () => {
    render(<StreakCalendarCard practiceDates={['2025-06-15', '2025-06-14']} />);
    const today = screen.getByTestId('calendar-cell-2025-06-15');
    expect(today.className).toContain('bg-primary-500');
  });

  it('現在の連続日数が正しく計算される', () => {
    render(
      <StreakCalendarCard
        practiceDates={['2025-06-15', '2025-06-14', '2025-06-13']}
      />
    );
    expect(screen.getByText('3')).toBeInTheDocument();
  });
});
