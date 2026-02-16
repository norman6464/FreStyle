import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import RecentSessionsCard from '../RecentSessionsCard';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

const mockSessions = [
  { sessionId: 1, sessionTitle: '練習: クレーム対応', overallScore: 8.5, scores: [], createdAt: '2025-01-15T10:00:00' },
  { sessionId: 2, sessionTitle: 'フリーチャット', overallScore: 7.2, scores: [], createdAt: '2025-01-14T09:00:00' },
  { sessionId: 3, sessionTitle: '練習: 会議ファシリ', overallScore: 6.8, scores: [], createdAt: '2025-01-13T08:00:00' },
];

describe('RecentSessionsCard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('タイトルが表示される', () => {
    render(<RecentSessionsCard sessions={mockSessions} />);

    expect(screen.getByText('直近のセッション')).toBeInTheDocument();
  });

  it('セッション名が表示される', () => {
    render(<RecentSessionsCard sessions={mockSessions} />);

    expect(screen.getByText('練習: クレーム対応')).toBeInTheDocument();
    expect(screen.getByText('フリーチャット')).toBeInTheDocument();
    expect(screen.getByText('練習: 会議ファシリ')).toBeInTheDocument();
  });

  it('スコアが表示される', () => {
    render(<RecentSessionsCard sessions={mockSessions} />);

    expect(screen.getByText('8.5')).toBeInTheDocument();
    expect(screen.getByText('7.2')).toBeInTheDocument();
    expect(screen.getByText('6.8')).toBeInTheDocument();
  });

  it('最大3件まで表示される', () => {
    const manySessions = [
      ...mockSessions,
      { sessionId: 4, sessionTitle: '追加セッション', overallScore: 5.0, scores: [], createdAt: '2025-01-12T08:00:00' },
    ];
    render(<RecentSessionsCard sessions={manySessions} />);

    expect(screen.queryByText('追加セッション')).not.toBeInTheDocument();
  });

  it('セッションがない場合は何も表示しない', () => {
    const { container } = render(<RecentSessionsCard sessions={[]} />);

    expect(container.firstChild).toBeNull();
  });

  it('セッション行をクリックするとAIセッションページに遷移する', () => {
    render(<RecentSessionsCard sessions={mockSessions} />);

    fireEvent.click(screen.getByText('練習: クレーム対応'));
    expect(mockNavigate).toHaveBeenCalledWith('/chat/ask-ai/1');
  });

  it('「すべて見る」をクリックするとスコア履歴ページに遷移する', () => {
    render(<RecentSessionsCard sessions={mockSessions} />);

    fireEvent.click(screen.getByText('すべて見る'));
    expect(mockNavigate).toHaveBeenCalledWith('/scores');
  });
});
