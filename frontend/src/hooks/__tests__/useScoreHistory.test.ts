import { describe, it, expect, vi, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useScoreHistory } from '../useScoreHistory';

const mockFetchScoreHistory = vi.fn();

vi.mock('../useAiChat', () => ({
  useAiChat: () => ({
    fetchScoreHistory: mockFetchScoreHistory,
    loading: false,
  }),
}));

describe('useScoreHistory', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('スコア履歴を取得する', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: 'テスト1', overallScore: 7.0, scores: [], createdAt: '2026-02-10' },
      { sessionId: 2, sessionTitle: 'テスト2', overallScore: 8.5, scores: [], createdAt: '2026-02-11' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(2);
    });

    expect(result.current.history[0].sessionTitle).toBe('テスト1');
  });

  it('latestSessionが最新のセッションを返す', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: '古い', overallScore: 6.0, scores: [], createdAt: '2026-02-10' },
      { sessionId: 2, sessionTitle: '最新', overallScore: 9.0, scores: [], createdAt: '2026-02-11' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.latestSession).not.toBeNull();
    });

    expect(result.current.latestSession?.sessionTitle).toBe('最新');
  });

  it('weakestAxisが最低スコアの軸を返す', async () => {
    const mockData = [
      {
        sessionId: 1,
        sessionTitle: 'テスト',
        overallScore: 7.0,
        scores: [
          { axis: '論理的構成力', score: 8, comment: '' },
          { axis: '要約力', score: 5, comment: '' },
          { axis: '提案力', score: 9, comment: '' },
        ],
        createdAt: '2026-02-10',
      },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.weakestAxis).not.toBeNull();
    });

    expect(result.current.weakestAxis?.axis).toBe('要約力');
  });

  it('フィルタが練習セッションを正しく絞り込む', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: '練習: テスト', overallScore: 7.0, scores: [], createdAt: '2026-02-10' },
      { sessionId: 2, sessionTitle: 'フリーチャット', overallScore: 8.0, scores: [], createdAt: '2026-02-11' },
      { sessionId: 3, sessionTitle: '練習：別テスト', overallScore: 6.0, scores: [], createdAt: '2026-02-12' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(3);
    });

    act(() => {
      result.current.setFilter('練習');
    });

    expect(result.current.filteredHistory).toHaveLength(2);
  });

  it('空の履歴でlatestSessionがnullになる', async () => {
    mockFetchScoreHistory.mockResolvedValue([]);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(mockFetchScoreHistory).toHaveBeenCalled();
    });

    expect(result.current.latestSession).toBeNull();
    expect(result.current.weakestAxis).toBeNull();
  });

  it('フリーフィルタでフリーセッションのみ返す', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: '練習: テスト', overallScore: 7.0, scores: [], createdAt: '2026-02-10' },
      { sessionId: 2, sessionTitle: 'フリーチャット', overallScore: 8.0, scores: [], createdAt: '2026-02-11' },
      { sessionId: 3, sessionTitle: '練習：別テスト', overallScore: 6.0, scores: [], createdAt: '2026-02-12' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(3);
    });

    act(() => {
      result.current.setFilter('フリー');
    });

    expect(result.current.filteredHistory).toHaveLength(1);
    expect(result.current.filteredHistory[0].sessionTitle).toBe('フリーチャット');
  });

  it('「すべて」フィルタで全セッションを返す', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: '練習: テスト', overallScore: 7.0, scores: [], createdAt: '2026-02-10' },
      { sessionId: 2, sessionTitle: 'フリーチャット', overallScore: 8.0, scores: [], createdAt: '2026-02-11' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(2);
    });

    // まず練習フィルタに変更
    act(() => {
      result.current.setFilter('練習');
    });
    expect(result.current.filteredHistory).toHaveLength(1);

    // すべてフィルタに戻す
    act(() => {
      result.current.setFilter('すべて');
    });
    expect(result.current.filteredHistory).toHaveLength(2);
  });

  it('selectedSessionの初期値がnullである', async () => {
    mockFetchScoreHistory.mockResolvedValue([]);
    const { result } = renderHook(() => useScoreHistory());
    expect(result.current.selectedSession).toBeNull();
  });

  it('setSelectedSessionでセッションを選択できる', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: 'テスト', overallScore: 7.0, scores: [], createdAt: '2026-02-10' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(1);
    });

    act(() => {
      result.current.setSelectedSession(result.current.history[0]);
    });

    expect(result.current.selectedSession?.sessionId).toBe(1);

    act(() => {
      result.current.setSelectedSession(null);
    });

    expect(result.current.selectedSession).toBeNull();
  });

  it('scoresがnullのセッションでもエラーにならない', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: 'テスト', overallScore: 7.0, scores: null, createdAt: '2026-02-10' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(1);
    });

    expect(result.current.history[0].scores).toEqual([]);
    expect(result.current.weakestAxis).toBeNull();
  });

  it('scoresがundefinedのセッションでもエラーにならない', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: 'テスト', overallScore: 7.0, createdAt: '2026-02-10' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(1);
    });

    expect(result.current.history[0].scores).toEqual([]);
    expect(result.current.weakestAxis).toBeNull();
  });

  it('averageScoreが全セッションの平均を返す', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: 'テスト1', overallScore: 7.0, scores: [], createdAt: '2026-02-10' },
      { sessionId: 2, sessionTitle: 'テスト2', overallScore: 8.5, scores: [], createdAt: '2026-02-11' },
      { sessionId: 3, sessionTitle: 'テスト3', overallScore: 9.0, scores: [], createdAt: '2026-02-12' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(3);
    });

    // (7.0 + 8.5 + 9.0) / 3 = 8.166... → 8.2
    expect(result.current.averageScore).toBe(8.2);
  });

  it('空の履歴でaverageScoreが0を返す', async () => {
    mockFetchScoreHistory.mockResolvedValue([]);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(mockFetchScoreHistory).toHaveBeenCalled();
    });

    expect(result.current.averageScore).toBe(0);
  });

  it('filteredHistoryWithDeltaがデルタ値を含む', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: 'テスト1', overallScore: 7.0, scores: [], createdAt: '2026-02-10' },
      { sessionId: 2, sessionTitle: 'テスト2', overallScore: 8.5, scores: [], createdAt: '2026-02-11' },
      { sessionId: 3, sessionTitle: 'テスト3', overallScore: 6.0, scores: [], createdAt: '2026-02-12' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.filteredHistoryWithDelta).toHaveLength(3);
    });

    // 最初のアイテムはデルタがnull
    expect(result.current.filteredHistoryWithDelta[0].delta).toBeNull();
    // 2番目: 8.5 - 7.0 = 1.5
    expect(result.current.filteredHistoryWithDelta[1].delta).toBe(1.5);
    // 3番目: 6.0 - 8.5 = -2.5
    expect(result.current.filteredHistoryWithDelta[2].delta).toBe(-2.5);
  });

  it('filteredHistoryWithDeltaがフィルタ適用時もオリジナルインデックスでデルタを計算する', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: '練習: A', overallScore: 5.0, scores: [], createdAt: '2026-02-10' },
      { sessionId: 2, sessionTitle: 'フリー', overallScore: 7.0, scores: [], createdAt: '2026-02-11' },
      { sessionId: 3, sessionTitle: '練習: B', overallScore: 9.0, scores: [], createdAt: '2026-02-12' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(3);
    });

    act(() => {
      result.current.setFilter('練習');
    });

    // 練習フィルタ: セッション1(5.0)とセッション3(9.0)
    expect(result.current.filteredHistoryWithDelta).toHaveLength(2);
    // セッション1は最初なのでデルタnull
    expect(result.current.filteredHistoryWithDelta[0].delta).toBeNull();
    // セッション3: 9.0 - 7.0(前のセッション2) = 2.0 (オリジナルインデックスで計算)
    expect(result.current.filteredHistoryWithDelta[1].delta).toBe(2.0);
  });

  it('最新セッションのスコアが空の場合にweakestAxisがnullになる', async () => {
    const mockData = [
      { sessionId: 1, sessionTitle: 'テスト', overallScore: 7.0, scores: [], createdAt: '2026-02-10' },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.latestSession).not.toBeNull();
    });

    expect(result.current.weakestAxis).toBeNull();
  });

  it('期間フィルタ「1週間」で直近1週間のセッションのみ返す', async () => {
    const now = new Date();
    const twoDaysAgo = new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000).toISOString();
    const tenDaysAgo = new Date(now.getTime() - 10 * 24 * 60 * 60 * 1000).toISOString();

    const mockData = [
      { sessionId: 1, sessionTitle: '古い', overallScore: 6.0, scores: [], createdAt: tenDaysAgo },
      { sessionId: 2, sessionTitle: '新しい', overallScore: 8.0, scores: [], createdAt: twoDaysAgo },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(2);
    });

    act(() => {
      result.current.setPeriodFilter('1週間');
    });

    expect(result.current.filteredHistory).toHaveLength(1);
    expect(result.current.filteredHistory[0].sessionTitle).toBe('新しい');
  });

  it('期間フィルタ「1ヶ月」で直近1ヶ月のセッションのみ返す', async () => {
    const now = new Date();
    const fiveDaysAgo = new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000).toISOString();
    const fiftyDaysAgo = new Date(now.getTime() - 50 * 24 * 60 * 60 * 1000).toISOString();

    const mockData = [
      { sessionId: 1, sessionTitle: '古い', overallScore: 5.0, scores: [], createdAt: fiftyDaysAgo },
      { sessionId: 2, sessionTitle: '新しい', overallScore: 7.5, scores: [], createdAt: fiveDaysAgo },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(2);
    });

    act(() => {
      result.current.setPeriodFilter('1ヶ月');
    });

    expect(result.current.filteredHistory).toHaveLength(1);
    expect(result.current.filteredHistory[0].sessionTitle).toBe('新しい');
  });

  it('期間フィルタ「3ヶ月」で直近3ヶ月のセッションのみ返す', async () => {
    const now = new Date();
    const twentyDaysAgo = new Date(now.getTime() - 20 * 24 * 60 * 60 * 1000).toISOString();
    const hundredDaysAgo = new Date(now.getTime() - 100 * 24 * 60 * 60 * 1000).toISOString();

    const mockData = [
      { sessionId: 1, sessionTitle: '古い', overallScore: 4.0, scores: [], createdAt: hundredDaysAgo },
      { sessionId: 2, sessionTitle: '新しい', overallScore: 8.0, scores: [], createdAt: twentyDaysAgo },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(2);
    });

    act(() => {
      result.current.setPeriodFilter('3ヶ月');
    });

    expect(result.current.filteredHistory).toHaveLength(1);
    expect(result.current.filteredHistory[0].sessionTitle).toBe('新しい');
  });

  it('期間フィルタ「全期間」で全セッションを返す', async () => {
    const now = new Date();
    const oneYearAgo = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000).toISOString();
    const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000).toISOString();

    const mockData = [
      { sessionId: 1, sessionTitle: '古い', overallScore: 5.0, scores: [], createdAt: oneYearAgo },
      { sessionId: 2, sessionTitle: '新しい', overallScore: 9.0, scores: [], createdAt: yesterday },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(2);
    });

    act(() => {
      result.current.setPeriodFilter('全期間');
    });

    expect(result.current.filteredHistory).toHaveLength(2);
  });

  it('期間フィルタとセッション種別フィルタを組み合わせて使える', async () => {
    const now = new Date();
    const twoDaysAgo = new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000).toISOString();
    const tenDaysAgo = new Date(now.getTime() - 10 * 24 * 60 * 60 * 1000).toISOString();

    const mockData = [
      { sessionId: 1, sessionTitle: '練習: 古い', overallScore: 5.0, scores: [], createdAt: tenDaysAgo },
      { sessionId: 2, sessionTitle: 'フリー新しい', overallScore: 7.0, scores: [], createdAt: twoDaysAgo },
      { sessionId: 3, sessionTitle: '練習: 新しい', overallScore: 8.0, scores: [], createdAt: twoDaysAgo },
    ];
    mockFetchScoreHistory.mockResolvedValue(mockData);

    const { result } = renderHook(() => useScoreHistory());

    await waitFor(() => {
      expect(result.current.history).toHaveLength(3);
    });

    act(() => {
      result.current.setPeriodFilter('1週間');
      result.current.setFilter('練習');
    });

    expect(result.current.filteredHistory).toHaveLength(1);
    expect(result.current.filteredHistory[0].sessionTitle).toBe('練習: 新しい');
  });
});
