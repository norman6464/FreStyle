import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import PracticeReminderCard from '../PracticeReminderCard';

describe('PracticeReminderCard', () => {
  it('今日練習した場合は「今日も練習済み」と表示される', () => {
    const today = new Date().toISOString();
    render(<BrowserRouter><PracticeReminderCard lastPracticeDate={today} /></BrowserRouter>);

    expect(screen.getByText('今日も練習済み！')).toBeInTheDocument();
  });

  it('1日前の場合は「1日前に練習」と表示される', () => {
    const yesterday = new Date(Date.now() - 86400000).toISOString();
    render(<BrowserRouter><PracticeReminderCard lastPracticeDate={yesterday} /></BrowserRouter>);

    expect(screen.getByText('最後の練習: 1日前')).toBeInTheDocument();
  });

  it('3日以上前の場合は練習を促すメッセージが表示される', () => {
    const threeDaysAgo = new Date(Date.now() - 86400000 * 3).toISOString();
    render(<BrowserRouter><PracticeReminderCard lastPracticeDate={threeDaysAgo} /></BrowserRouter>);

    expect(screen.getByText('最後の練習: 3日前')).toBeInTheDocument();
    expect(screen.getByText('そろそろ練習しませんか？')).toBeInTheDocument();
  });

  it('「練習する」ボタンが表示される', () => {
    const yesterday = new Date(Date.now() - 86400000).toISOString();
    render(<BrowserRouter><PracticeReminderCard lastPracticeDate={yesterday} /></BrowserRouter>);

    expect(screen.getByText('練習する')).toBeInTheDocument();
  });

  it('練習日がない場合は何も表示しない', () => {
    const { container } = render(<BrowserRouter><PracticeReminderCard lastPracticeDate={null} /></BrowserRouter>);

    expect(container.firstChild).toBeNull();
  });
});
