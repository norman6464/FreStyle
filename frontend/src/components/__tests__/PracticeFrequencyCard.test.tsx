import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import PracticeFrequencyCard from '../PracticeFrequencyCard';

describe('PracticeFrequencyCard', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2025-07-14T10:00:00')); // 月曜日
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('タイトルが表示される', () => {
    render(<PracticeFrequencyCard dates={[]} />);
    expect(screen.getByText('週別練習頻度')).toBeInTheDocument();
  });

  it('練習日がない場合は全て0回と表示する', () => {
    render(<PracticeFrequencyCard dates={[]} />);
    const bars = screen.getAllByTestId('frequency-bar');
    expect(bars).toHaveLength(4);
    bars.forEach(bar => {
      expect(bar.style.height).toBe('0%');
    });
  });

  it('今週の練習回数を正しくカウントする', () => {
    const dates = [
      '2025-07-14T09:00:00', // 今週月曜
      '2025-07-14T15:00:00', // 今週月曜（2回目）
    ];
    render(<PracticeFrequencyCard dates={dates} />);
    expect(screen.getByText('2回')).toBeInTheDocument();
  });

  it('4週間分のバーが表示される', () => {
    render(<PracticeFrequencyCard dates={[]} />);
    const bars = screen.getAllByTestId('frequency-bar');
    expect(bars).toHaveLength(4);
  });

  it('今週ラベルがハイライトされる', () => {
    render(<PracticeFrequencyCard dates={[]} />);
    expect(screen.getByText('今週')).toBeInTheDocument();
  });

  it('前週比の増減を表示する', () => {
    const dates = [
      '2025-07-14T09:00:00', // 今週: 1回
      '2025-07-07T09:00:00', // 先週: 1回
      '2025-07-08T09:00:00', // 先週: 2回目
    ];
    render(<PracticeFrequencyCard dates={dates} />);
    // 今週1回 - 先週2回 = -1
    expect(screen.getByText(/\u22121/)).toBeInTheDocument();
  });

  it('前週比が増加の場合はプラス表示する', () => {
    const dates = [
      '2025-07-14T09:00:00', // 今週: 1回
      '2025-07-14T15:00:00', // 今週: 2回目
      '2025-07-14T20:00:00', // 今週: 3回目
      '2025-07-07T09:00:00', // 先週: 1回
    ];
    render(<PracticeFrequencyCard dates={dates} />);
    expect(screen.getByText('+2')).toBeInTheDocument();
  });
});
