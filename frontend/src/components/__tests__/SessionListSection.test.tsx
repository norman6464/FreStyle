import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import SessionListSection from '../SessionListSection';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

const mockHistory = [
  {
    sessionId: 1,
    sessionTitle: '会議フィードバック',
    overallScore: 7.4,
    scores: [{ axis: '論理的構成力', score: 8, comment: '良い' }],
    createdAt: '2026-01-15T10:00:00Z',
    delta: null,
  },
  {
    sessionId: 2,
    sessionTitle: '練習: テスト',
    overallScore: 8.0,
    scores: [{ axis: '論理的構成力', score: 9, comment: '素晴らしい' }],
    createdAt: '2026-01-16T10:00:00Z',
    delta: 0.6,
  },
];

describe('SessionListSection', () => {
  it('セッション一覧が表示される', () => {
    render(
      <SessionListSection
        history={mockHistory}
        filteredHistoryWithDelta={mockHistory}
        filter="すべて"
        setFilter={vi.fn()}
        selectedSession={null}
        setSelectedSession={vi.fn()}
      />
    );
    expect(screen.getByText('会議フィードバック')).toBeInTheDocument();
    expect(screen.getByText('練習: テスト')).toBeInTheDocument();
  });

  it('フィルタタブが表示される', () => {
    render(
      <SessionListSection
        history={mockHistory}
        filteredHistoryWithDelta={mockHistory}
        filter="すべて"
        setFilter={vi.fn()}
        selectedSession={null}
        setSelectedSession={vi.fn()}
      />
    );
    expect(screen.getByRole('tab', { name: 'すべて' })).toBeInTheDocument();
  });
});
