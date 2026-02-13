import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import RecentSessionsCard from '../RecentSessionsCard';

const mockSessions = [
  { sessionId: 1, sessionTitle: '練習: クレーム対応', overallScore: 8.5, createdAt: '2025-01-15T10:00:00' },
  { sessionId: 2, sessionTitle: 'フリーチャット', overallScore: 7.2, createdAt: '2025-01-14T09:00:00' },
  { sessionId: 3, sessionTitle: '練習: 会議ファシリ', overallScore: 6.8, createdAt: '2025-01-13T08:00:00' },
];

describe('RecentSessionsCard', () => {
  it('タイトルが表示される', () => {
    render(<BrowserRouter><RecentSessionsCard sessions={mockSessions} /></BrowserRouter>);

    expect(screen.getByText('直近のセッション')).toBeInTheDocument();
  });

  it('セッション名が表示される', () => {
    render(<BrowserRouter><RecentSessionsCard sessions={mockSessions} /></BrowserRouter>);

    expect(screen.getByText('練習: クレーム対応')).toBeInTheDocument();
    expect(screen.getByText('フリーチャット')).toBeInTheDocument();
    expect(screen.getByText('練習: 会議ファシリ')).toBeInTheDocument();
  });

  it('スコアが表示される', () => {
    render(<BrowserRouter><RecentSessionsCard sessions={mockSessions} /></BrowserRouter>);

    expect(screen.getByText('8.5')).toBeInTheDocument();
    expect(screen.getByText('7.2')).toBeInTheDocument();
    expect(screen.getByText('6.8')).toBeInTheDocument();
  });

  it('最大3件まで表示される', () => {
    const manySessions = [
      ...mockSessions,
      { sessionId: 4, sessionTitle: '追加セッション', overallScore: 5.0, createdAt: '2025-01-12T08:00:00' },
    ];
    render(<BrowserRouter><RecentSessionsCard sessions={manySessions} /></BrowserRouter>);

    expect(screen.queryByText('追加セッション')).not.toBeInTheDocument();
  });

  it('セッションがない場合は何も表示しない', () => {
    const { container } = render(<BrowserRouter><RecentSessionsCard sessions={[]} /></BrowserRouter>);

    expect(container.firstChild).toBeNull();
  });
});
