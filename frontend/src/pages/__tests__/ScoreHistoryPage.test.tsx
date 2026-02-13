import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import type { Mock } from 'vitest';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

const mockHistory = [
  {
    sessionId: 1,
    sessionTitle: '会議フィードバック',
    overallScore: 7.4,
    scores: [
      { axis: '論理的構成力', score: 8, comment: '良い' },
      { axis: '配慮表現', score: 7, comment: '普通' },
      { axis: '要約力', score: 7, comment: '普通' },
      { axis: '提案力', score: 4, comment: '改善必要' },
      { axis: '質問・傾聴力', score: 6, comment: '普通' },
    ],
    createdAt: '2026-01-15T10:00:00Z',
  },
  {
    sessionId: 2,
    sessionTitle: '練習: 本番障害の緊急報告',
    overallScore: 6.0,
    scores: [
      { axis: '論理的構成力', score: 6, comment: '改善の余地あり' },
      { axis: '配慮表現', score: 5, comment: '改善必要' },
      { axis: '要約力', score: 7, comment: '良い' },
      { axis: '提案力', score: 5, comment: '改善必要' },
      { axis: '質問・傾聴力', score: 7, comment: '良い' },
    ],
    createdAt: '2026-01-16T10:00:00Z',
  },
  {
    sessionId: 3,
    sessionTitle: '練習: 設計方針の意見対立',
    overallScore: 8.2,
    scores: [
      { axis: '論理的構成力', score: 9, comment: '素晴らしい' },
      { axis: '配慮表現', score: 8, comment: '良い' },
      { axis: '要約力', score: 8, comment: '良い' },
      { axis: '提案力', score: 7, comment: '改善の余地あり' },
      { axis: '質問・傾聴力', score: 9, comment: '素晴らしい' },
    ],
    createdAt: '2026-01-17T10:00:00Z',
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
      expect(screen.getByText('練習: 本番障害の緊急報告')).toBeInTheDocument();
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

  it('スコア推移の棒グラフが表示される', async () => {
    const ScoreHistoryPage = await importScoreHistoryPage();

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      expect(screen.getByText('スコア推移')).toBeInTheDocument();
    }, { timeout: 3000 });

    const bars = document.querySelectorAll('[data-testid="trend-bar"]');
    expect(bars.length).toBe(3);
  });

  it('スコア変動の矢印が表示される', async () => {
    const ScoreHistoryPage = await importScoreHistoryPage();

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      // セッション2（6.0）はセッション1（7.4）より低いので↓
      expect(screen.getByText('−1.4')).toBeInTheDocument();
      // セッション3（8.2）はセッション2（6.0）より高いので↑
      expect(screen.getByText('+2.2')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('「練習」フィルタで練習セッションのみ表示される', async () => {
    const ScoreHistoryPage = await importScoreHistoryPage();

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      expect(screen.getByText('会議フィードバック')).toBeInTheDocument();
    }, { timeout: 3000 });

    fireEvent.click(screen.getByRole('button', { name: '練習' }));

    expect(screen.queryByText('会議フィードバック')).not.toBeInTheDocument();
    expect(screen.getByText('練習: 本番障害の緊急報告')).toBeInTheDocument();
    expect(screen.getByText('練習: 設計方針の意見対立')).toBeInTheDocument();
  });

  it('「フリー」フィルタでフリーセッションのみ表示される', async () => {
    const ScoreHistoryPage = await importScoreHistoryPage();

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      expect(screen.getByText('会議フィードバック')).toBeInTheDocument();
    }, { timeout: 3000 });

    fireEvent.click(screen.getByRole('button', { name: 'フリー' }));

    expect(screen.getByText('会議フィードバック')).toBeInTheDocument();
    expect(screen.queryByText('練習: 本番障害の緊急報告')).not.toBeInTheDocument();
  });

  it('弱点ベースのおすすめ練習セクションが表示される', async () => {
    const ScoreHistoryPage = await importScoreHistoryPage();

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      expect(screen.getByText('おすすめ練習')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('最新スコアの弱点軸名が表示される', async () => {
    const ScoreHistoryPage = await importScoreHistoryPage();

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      // 最新セッション(sessionId:3)の最低スコアは「提案力」(7)
      expect(screen.getByText(/提案力.*を伸ばす/)).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('おすすめ練習ボタンクリックで練習ページに遷移する', async () => {
    const ScoreHistoryPage = await importScoreHistoryPage();

    render(<ScoreHistoryPage />);

    await waitFor(() => {
      expect(screen.getByText('おすすめ練習')).toBeInTheDocument();
    }, { timeout: 3000 });

    fireEvent.click(screen.getByText('練習一覧を見る'));
    expect(mockNavigate).toHaveBeenCalledWith('/practice');
  });
});
