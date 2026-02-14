import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import PracticeCalendar from '../PracticeCalendar';

interface ScoreHistory {
  sessionId: number;
  sessionTitle: string;
  overallScore: number;
  createdAt: string;
}

describe('PracticeCalendar', () => {
  it('タイトルが表示される', () => {
    render(<PracticeCalendar practiceDates={[]} />);

    expect(screen.getByText('練習カレンダー')).toBeInTheDocument();
  });

  it('曜日ラベルが表示される', () => {
    render(<PracticeCalendar practiceDates={[]} />);

    expect(screen.getByText('月')).toBeInTheDocument();
    expect(screen.getByText('水')).toBeInTheDocument();
    expect(screen.getByText('金')).toBeInTheDocument();
  });

  it('練習日のセルがアクティブなスタイルになる', () => {
    const now = new Date();
    const today = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
    render(<PracticeCalendar practiceDates={[today]} />);

    const activeCells = document.querySelectorAll('[data-active="true"]');
    expect(activeCells.length).toBeGreaterThanOrEqual(1);
  });

  it('練習がない日は非アクティブなスタイルになる', () => {
    render(<PracticeCalendar practiceDates={[]} />);

    const inactiveCells = document.querySelectorAll('[data-active="false"]');
    expect(inactiveCells.length).toBeGreaterThan(0);
  });

  it('複数回練習した日はより濃い色になる', () => {
    const now = new Date();
    const today = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
    render(<PracticeCalendar practiceDates={[today, today, today]} />);

    const activeCells = document.querySelectorAll('[data-count="3"]');
    expect(activeCells.length).toBe(1);
  });
});
