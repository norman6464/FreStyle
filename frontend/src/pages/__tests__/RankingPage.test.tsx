import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import RankingPage from '../RankingPage';
import { useRanking } from '../../hooks/useRanking';

vi.mock('../../hooks/useRanking');
const mockedUseRanking = vi.mocked(useRanking);

function renderPage() {
  return render(
    <MemoryRouter>
      <RankingPage />
    </MemoryRouter>
  );
}

describe('RankingPage', () => {
  const mockRanking = {
    entries: [
      { rank: 1, userId: 1, username: 'TopUser', iconUrl: null, averageScore: 9.5, sessionCount: 10 },
      { rank: 2, userId: 2, username: 'SecondUser', iconUrl: null, averageScore: 8.0, sessionCount: 5 },
    ],
    myRanking: { rank: 2, userId: 2, username: 'SecondUser', iconUrl: null, averageScore: 8.0, sessionCount: 5 },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockedUseRanking.mockReturnValue({
      ranking: mockRanking,
      period: 'weekly',
      changePeriod: vi.fn(),
      loading: false,
      error: null,
    });
  });

  it('ランキングタイトルを表示する', () => {
    renderPage();
    expect(screen.getByText('ランキング')).toBeInTheDocument();
  });

  it('ランキングエントリーを表示する', () => {
    renderPage();
    expect(screen.getByText('TopUser')).toBeInTheDocument();
    // SecondUserはmyRankingとentriesの両方に表示される
    expect(screen.getAllByText('SecondUser').length).toBeGreaterThanOrEqual(1);
  });

  it('自分の順位セクションを表示する', () => {
    renderPage();
    expect(screen.getByText('あなたの順位')).toBeInTheDocument();
  });

  it('ローディング中はローディング表示する', () => {
    mockedUseRanking.mockReturnValue({
      ranking: null,
      period: 'weekly',
      changePeriod: vi.fn(),
      loading: true,
      error: null,
    });
    renderPage();
    expect(screen.getByText('ランキングを読み込み中...')).toBeInTheDocument();
  });

  it('エラー時はエラーメッセージを表示する', () => {
    mockedUseRanking.mockReturnValue({
      ranking: null,
      period: 'weekly',
      changePeriod: vi.fn(),
      loading: false,
      error: 'ランキングの取得に失敗しました',
    });
    renderPage();
    expect(screen.getByText('ランキングの取得に失敗しました')).toBeInTheDocument();
  });

  it('期間ボタンをクリックするとchangePeriodが呼ばれる', () => {
    const changePeriod = vi.fn();
    mockedUseRanking.mockReturnValue({
      ranking: mockRanking,
      period: 'weekly',
      changePeriod,
      loading: false,
      error: null,
    });
    renderPage();
    fireEvent.click(screen.getByTestId('period-monthly'));
    expect(changePeriod).toHaveBeenCalledWith('monthly');
  });
});
