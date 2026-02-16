import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useFavoritePhrase } from '../useFavoritePhrase';

const mockGetAll = vi.fn();
const mockSave = vi.fn();
const mockRemove = vi.fn();
const mockExists = vi.fn();

vi.mock('../../repositories/FavoritePhraseRepository', () => ({
  FavoritePhraseRepository: {
    getAll: (...args: unknown[]) => mockGetAll(...args),
    save: (...args: unknown[]) => mockSave(...args),
    remove: (...args: unknown[]) => mockRemove(...args),
    exists: (...args: unknown[]) => mockExists(...args),
  },
}));

const phrase1 = { id: '1', originalText: 'テスト', rephrasedText: '言い換え', pattern: 'フォーマル', createdAt: '2026-01-01' };
const phrase2 = { id: '2', originalText: 'テスト2', rephrasedText: '言い換え2', pattern: 'ソフト', createdAt: '2026-01-02' };

describe('useFavoritePhrase', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetAll.mockResolvedValue([]);
    mockSave.mockResolvedValue(undefined);
    mockRemove.mockResolvedValue(undefined);
    mockExists.mockResolvedValue(false);
  });

  it('初期状態はloading=trueで空配列', () => {
    const { result } = renderHook(() => useFavoritePhrase());
    expect(result.current.phrases).toEqual([]);
    expect(result.current.loading).toBe(true);
  });

  it('マウント時にAPIからフレーズを取得する', async () => {
    mockGetAll.mockResolvedValue([phrase1]);

    const { result } = renderHook(() => useFavoritePhrase());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.phrases).toEqual([phrase1]);
  });

  it('フレーズを保存して一覧を更新する', async () => {
    mockGetAll.mockResolvedValueOnce([]).mockResolvedValueOnce([phrase1]);

    const { result } = renderHook(() => useFavoritePhrase());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.saveFavorite('テスト', '言い換え', 'フォーマル');
    });

    expect(mockSave).toHaveBeenCalledWith({
      originalText: 'テスト',
      rephrasedText: '言い換え',
      pattern: 'フォーマル',
    });
    expect(result.current.phrases).toEqual([phrase1]);
  });

  it('フレーズを削除すると即座にUIが更新される', async () => {
    mockGetAll.mockResolvedValue([phrase1, phrase2]);

    const { result } = renderHook(() => useFavoritePhrase());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.removeFavorite('1');
    });

    expect(mockRemove).toHaveBeenCalledWith('1');
    expect(result.current.phrases).toEqual([phrase2]);
  });

  it('isFavoriteでローカルstateから判定できる', async () => {
    mockGetAll.mockResolvedValue([phrase1]);

    const { result } = renderHook(() => useFavoritePhrase());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.isFavorite('言い換え', 'フォーマル')).toBe(true);
    expect(result.current.isFavorite('未登録', 'カジュアル')).toBe(false);
  });

  it('searchQueryの初期値が空文字', async () => {
    const { result } = renderHook(() => useFavoritePhrase());
    expect(result.current.searchQuery).toBe('');
  });

  it('patternFilterの初期値がすべて', async () => {
    const { result } = renderHook(() => useFavoritePhrase());
    expect(result.current.patternFilter).toBe('すべて');
  });

  it('パターンフィルタで絞り込みできる', async () => {
    mockGetAll.mockResolvedValue([phrase1, phrase2]);

    const { result } = renderHook(() => useFavoritePhrase());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    act(() => {
      result.current.setPatternFilter('フォーマル');
    });

    expect(result.current.filteredPhrases).toHaveLength(1);
    expect(result.current.filteredPhrases[0].pattern).toBe('フォーマル');
  });

  it('検索クエリで絞り込みできる', async () => {
    mockGetAll.mockResolvedValue([phrase1, phrase2]);

    const { result } = renderHook(() => useFavoritePhrase());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    act(() => {
      result.current.setSearchQuery('テスト2');
    });

    expect(result.current.filteredPhrases).toHaveLength(1);
    expect(result.current.filteredPhrases[0].originalText).toBe('テスト2');
  });
});
