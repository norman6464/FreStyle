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
});
