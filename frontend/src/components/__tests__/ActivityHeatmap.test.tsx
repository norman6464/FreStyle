import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ActivityHeatmap from '../ActivityHeatmap';

function formatDate(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

describe('ActivityHeatmap', () => {
  it('ヒートマップが表示される', () => {
    render(<ActivityHeatmap practiceDates={[]} />);

    expect(screen.getByText('年間の学習活動')).toBeInTheDocument();
  });

  it('練習日数が表示される', () => {
    const today = new Date();
    const dates = [formatDate(today)];

    render(<ActivityHeatmap practiceDates={dates} />);

    expect(screen.getByText(/1日/)).toBeInTheDocument();
  });

  it('練習した日のセルが色付きで表示される', () => {
    const today = new Date();
    const dateStr = formatDate(today);

    render(<ActivityHeatmap practiceDates={[dateStr]} />);

    const activeCell = screen.getByTitle(dateStr);
    expect(activeCell.className).toContain('bg-emerald');
  });

  it('練習していない日のセルが薄い色で表示される', () => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    const dateStr = formatDate(yesterday);

    // 空の練習日リストなので全セルが非アクティブ
    render(<ActivityHeatmap practiceDates={[]} />);

    const cell = screen.getByTitle(dateStr);
    expect(cell.className).toContain('bg-surface');
  });

  it('月ラベルが表示される', () => {
    render(<ActivityHeatmap practiceDates={[]} />);

    // 少なくとも1つの月ラベルが表示される
    const monthLabels = ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月'];
    const found = monthLabels.some((label) => screen.queryByText(label) !== null);
    expect(found).toBe(true);
  });

  it('複数の練習日を正しく反映する', () => {
    const today = new Date();
    const dates = [];
    for (let i = 0; i < 5; i++) {
      const d = new Date(today);
      d.setDate(d.getDate() - i);
      dates.push(formatDate(d));
    }

    render(<ActivityHeatmap practiceDates={dates} />);

    expect(screen.getByText(/5日/)).toBeInTheDocument();
  });
});
