import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { usePracticePage } from '../usePracticePage';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

const mockFetchScenarios = vi.fn();
const mockCreatePracticeSession = vi.fn();
vi.mock('../usePractice', () => ({
  usePractice: () => ({
    scenarios: [
      { id: 1, name: 'シナリオC', category: 'customer', difficulty: 'advanced' },
      { id: 2, name: 'シナリオA', category: 'senior', difficulty: 'beginner' },
      { id: 3, name: 'シナリオB', category: 'team', difficulty: 'intermediate' },
    ],
    loading: false,
    fetchScenarios: mockFetchScenarios,
    createPracticeSession: mockCreatePracticeSession,
  }),
}));

const mockBookmarkedIds = [1];
const mockToggleBookmark = vi.fn();
const mockIsBookmarked = vi.fn((id: number) => mockBookmarkedIds.includes(id));
vi.mock('../useBookmark', () => ({
  useBookmark: () => ({
    bookmarkedIds: mockBookmarkedIds,
    toggleBookmark: mockToggleBookmark,
    isBookmarked: mockIsBookmarked,
  }),
}));

describe('usePracticePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('初期カテゴリはすべて', () => {
    const { result } = renderHook(() => usePracticePage());
    expect(result.current.selectedCategory).toBe('すべて');
  });

  it('カテゴリを変更できる', () => {
    const { result } = renderHook(() => usePracticePage());
    act(() => {
      result.current.setSelectedCategory('顧客折衝');
    });
    expect(result.current.selectedCategory).toBe('顧客折衝');
  });

  it('すべてカテゴリで全シナリオを返す', () => {
    const { result } = renderHook(() => usePracticePage());
    expect(result.current.filteredScenarios).toHaveLength(3);
  });

  it('カテゴリフィルタでシナリオを絞り込む', () => {
    const { result } = renderHook(() => usePracticePage());
    act(() => {
      result.current.setSelectedCategory('顧客折衝');
    });
    expect(result.current.filteredScenarios).toHaveLength(1);
    expect(result.current.filteredScenarios[0].category).toBe('customer');
  });

  it('ブックマークフィルタでブックマーク済みのみ返す', () => {
    const { result } = renderHook(() => usePracticePage());
    act(() => {
      result.current.setSelectedCategory('ブックマーク');
    });
    expect(result.current.filteredScenarios).toHaveLength(1);
    expect(result.current.filteredScenarios[0].id).toBe(1);
  });

  it('handleSelectScenarioでセッション作成後にナビゲートする', async () => {
    mockCreatePracticeSession.mockResolvedValue({ id: 99 });
    const { result } = renderHook(() => usePracticePage());

    await act(async () => {
      await result.current.handleSelectScenario({ id: 1, name: 'テスト' } as any);
    });

    expect(mockCreatePracticeSession).toHaveBeenCalledWith({ scenarioId: 1 });
    expect(mockNavigate).toHaveBeenCalledWith('/chat/ask-ai/99', {
      state: {
        sessionType: 'practice',
        scenarioId: 1,
        scenarioName: 'テスト',
        initialPrompt: '練習開始',
      },
    });
  });

  it('handleSelectScenarioでセッション作成失敗時はナビゲートしない', async () => {
    mockCreatePracticeSession.mockResolvedValue(null);
    const { result } = renderHook(() => usePracticePage());

    await act(async () => {
      await result.current.handleSelectScenario({ id: 1, name: 'テスト' } as any);
    });

    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('マウント時にfetchScenariosを呼ぶ', () => {
    renderHook(() => usePracticePage());
    expect(mockFetchScenarios).toHaveBeenCalled();
  });

  it('loading状態を返す', () => {
    const { result } = renderHook(() => usePracticePage());
    expect(result.current.loading).toBe(false);
  });

  it('bookmark関連のプロパティを返す', () => {
    const { result } = renderHook(() => usePracticePage());
    expect(result.current.isBookmarked).toBeDefined();
    expect(result.current.toggleBookmark).toBeDefined();
  });

  it('初期ソートはdefault', () => {
    const { result } = renderHook(() => usePracticePage());
    expect(result.current.selectedSort).toBe('default');
  });

  it('難易度昇順でソートできる', () => {
    const { result } = renderHook(() => usePracticePage());
    act(() => {
      result.current.setSelectedSort('difficulty-asc');
    });
    const names = result.current.filteredScenarios.map((s) => s.name);
    expect(names).toEqual(['シナリオA', 'シナリオB', 'シナリオC']);
  });

  it('難易度降順でソートできる', () => {
    const { result } = renderHook(() => usePracticePage());
    act(() => {
      result.current.setSelectedSort('difficulty-desc');
    });
    const names = result.current.filteredScenarios.map((s) => s.name);
    expect(names).toEqual(['シナリオC', 'シナリオB', 'シナリオA']);
  });

  it('名前順でソートできる', () => {
    const { result } = renderHook(() => usePracticePage());
    act(() => {
      result.current.setSelectedSort('name');
    });
    const names = result.current.filteredScenarios.map((s) => s.name);
    expect(names).toEqual(['シナリオA', 'シナリオB', 'シナリオC']);
  });
});
