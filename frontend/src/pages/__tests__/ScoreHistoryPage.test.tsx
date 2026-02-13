import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import type { Mock } from 'vitest';

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

const mockFetchScoreHistory: Mock = vi.fn();
const mockUseAiChat = {
  fetchScoreHistory: mockFetchScoreHistory,
  loading: false,
};

vi.mock('../../hooks/useAiChat', () => ({
  useAiChat: () => mockUseAiChat,
}));

// モックをクリアしてから動的にインポート
const importScoreHistoryPage = async () => {
  return (await import('../ScoreHistoryPage')).default;
};

describe('ScoreHistoryPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAiChat.loading = false;
    mockFetchScoreHistory.mockResolvedValue(mockHistory);
  });

  it('スコア履歴一覧が表示される', async () => {
    const ScoreHistoryPage = await importScoreHistoryPage();

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      expect(screen.getByText('会議フィードバック')).toBeInTheDocument();
      expect(screen.getByText('コードレビューフィードバック')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('総合スコアが表示される', async () => {
    const ScoreHistoryPage = await importScoreHistoryPage();

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      expect(screen.getByText('7.4')).toBeInTheDocument();
      expect(screen.getByText('6.0')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('履歴が空の場合メッセージが表示される', async () => {
    mockFetchScoreHistory.mockResolvedValue([]);
    const ScoreHistoryPage = await importScoreHistoryPage();

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      expect(screen.getByText('スコア履歴がありません')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('ローディング中はスケルトンが表示される', async () => {
    mockUseAiChat.loading = true;
    const ScoreHistoryPage = await importScoreHistoryPage();

    render(<ScoreHistoryPage />);

    expect(screen.getByText('スコア履歴を読み込み中...')).toBeInTheDocument();
  });
});
