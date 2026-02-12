import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import ScoreHistoryPage from '../ScoreHistoryPage';

const mockHistory = [
  {
    sessionId: 1,
    sessionTitle: '会議フィードバック',
    overallScore: 7.4,
    scores: [
      { axis: '論理的構成力', score: 8, comment: '良い' },
      { axis: '配慮表現', score: 7, comment: '普通' },
    ],
    createdAt: '2026-01-15T10:00:00Z',
  },
  {
    sessionId: 2,
    sessionTitle: 'コードレビューフィードバック',
    overallScore: 6.0,
    scores: [
      { axis: '論理的構成力', score: 6, comment: '改善の余地あり' },
    ],
    createdAt: '2026-01-16T10:00:00Z',
  },
];

describe('ScoreHistoryPage', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('スコア履歴一覧が表示される', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockHistory),
    });

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      expect(screen.getByText('会議フィードバック')).toBeInTheDocument();
      expect(screen.getByText('コードレビューフィードバック')).toBeInTheDocument();
    });
  });

  it('総合スコアが表示される', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockHistory),
    });

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      expect(screen.getByText('7.4')).toBeInTheDocument();
      expect(screen.getByText('6.0')).toBeInTheDocument();
    });
  });

  it('履歴が空の場合メッセージが表示される', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([]),
    });

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      expect(screen.getByText('スコア履歴がありません')).toBeInTheDocument();
    });
  });

  it('ローディング中はスピナーが表示される', () => {
    global.fetch = vi.fn().mockReturnValue(new Promise(() => {}));

    render(<ScoreHistoryPage />);

    expect(document.querySelector('.animate-spin')).toBeInTheDocument();
  });
});
