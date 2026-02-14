import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useFavoritePhrase } from '../useFavoritePhrase';
import { FavoritePhraseRepository } from '../../repositories/FavoritePhraseRepository';

vi.mock('../../repositories/FavoritePhraseRepository');

const mockedRepo = vi.mocked(FavoritePhraseRepository);

describe('useFavoritePhrase', () => {
  beforeEach(() => {
    mockedRepo.getAll.mockReturnValue([]);
    mockedRepo.exists.mockReturnValue(false);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('初期状態でフレーズ一覧を取得する', () => {
    const mockPhrases = [
      { id: '1', originalText: 'テスト', rephrasedText: '言い換え', pattern: 'フォーマル', createdAt: '2026-01-01' },
    ];
    mockedRepo.getAll.mockReturnValue(mockPhrases);

    const { result } = renderHook(() => useFavoritePhrase());

    expect(result.current.phrases).toEqual(mockPhrases);
    expect(mockedRepo.getAll).toHaveBeenCalled();
  });

  it('フレーズを保存して一覧を更新する', () => {
    const newPhrase = { id: '1', originalText: 'テスト', rephrasedText: '言い換え', pattern: 'フォーマル', createdAt: '2026-01-01' };
    mockedRepo.getAll
      .mockReturnValueOnce([])
      .mockReturnValueOnce([newPhrase]);

    const { result } = renderHook(() => useFavoritePhrase());

    act(() => {
      result.current.saveFavorite('テスト', '言い換え', 'フォーマル');
    });

    expect(mockedRepo.save).toHaveBeenCalledWith({
      originalText: 'テスト',
      rephrasedText: '言い換え',
      pattern: 'フォーマル',
    });
    expect(result.current.phrases).toEqual([newPhrase]);
  });

  it('フレーズを削除して一覧を更新する', () => {
    const phrase = { id: '1', originalText: 'テスト', rephrasedText: '言い換え', pattern: 'フォーマル', createdAt: '2026-01-01' };
    mockedRepo.getAll
      .mockReturnValueOnce([phrase])
      .mockReturnValueOnce([]);

    const { result } = renderHook(() => useFavoritePhrase());

    act(() => {
      result.current.removeFavorite('1');
    });

    expect(mockedRepo.remove).toHaveBeenCalledWith('1');
    expect(result.current.phrases).toEqual([]);
  });

  it('重複チェックが機能する', () => {
    mockedRepo.exists.mockReturnValue(true);

    const { result } = renderHook(() => useFavoritePhrase());

    expect(result.current.isFavorite('言い換え', 'フォーマル')).toBe(true);
    expect(mockedRepo.exists).toHaveBeenCalledWith('言い換え', 'フォーマル');
  });

  it('未登録時にisFavoriteがfalseを返す', () => {
    mockedRepo.exists.mockReturnValue(false);

    const { result } = renderHook(() => useFavoritePhrase());

    expect(result.current.isFavorite('未登録', 'カジュアル')).toBe(false);
    expect(mockedRepo.exists).toHaveBeenCalledWith('未登録', 'カジュアル');
  });

  it('複数フレーズ保存で一覧が正しく更新される', () => {
    const phrase1 = { id: '1', originalText: 'テスト1', rephrasedText: '言い換え1', pattern: 'フォーマル', createdAt: '2026-01-01' };
    const phrase2 = { id: '2', originalText: 'テスト2', rephrasedText: '言い換え2', pattern: 'カジュアル', createdAt: '2026-01-02' };
    mockedRepo.getAll
      .mockReturnValueOnce([])
      .mockReturnValueOnce([phrase1])
      .mockReturnValueOnce([phrase1, phrase2]);

    const { result } = renderHook(() => useFavoritePhrase());

    act(() => {
      result.current.saveFavorite('テスト1', '言い換え1', 'フォーマル');
    });
    expect(result.current.phrases).toEqual([phrase1]);

    act(() => {
      result.current.saveFavorite('テスト2', '言い換え2', 'カジュアル');
    });
    expect(result.current.phrases).toEqual([phrase1, phrase2]);
  });

  it('初期状態でフレーズが空配列の場合', () => {
    mockedRepo.getAll.mockReturnValue([]);

    const { result } = renderHook(() => useFavoritePhrase());

    expect(result.current.phrases).toEqual([]);
  });

  it('searchQueryの初期値が空文字', () => {
    const { result } = renderHook(() => useFavoritePhrase());
    expect(result.current.searchQuery).toBe('');
  });

  it('patternFilterの初期値がすべて', () => {
    const { result } = renderHook(() => useFavoritePhrase());
    expect(result.current.patternFilter).toBe('すべて');
  });

  it('パターンフィルタで絞り込みできる', () => {
    const phrases = [
      { id: '1', originalText: 'テスト1', rephrasedText: '言い換え1', pattern: 'フォーマル', createdAt: '2026-01-01' },
      { id: '2', originalText: 'テスト2', rephrasedText: '言い換え2', pattern: 'ソフト', createdAt: '2026-01-02' },
    ];
    mockedRepo.getAll.mockReturnValue(phrases);

    const { result } = renderHook(() => useFavoritePhrase());

    act(() => {
      result.current.setPatternFilter('フォーマル');
    });

    expect(result.current.filteredPhrases).toHaveLength(1);
    expect(result.current.filteredPhrases[0].pattern).toBe('フォーマル');
  });

  it('検索クエリで絞り込みできる', () => {
    const phrases = [
      { id: '1', originalText: '会議', rephrasedText: 'ミーティング', pattern: 'フォーマル', createdAt: '2026-01-01' },
      { id: '2', originalText: '報告', rephrasedText: 'レポート', pattern: 'ソフト', createdAt: '2026-01-02' },
    ];
    mockedRepo.getAll.mockReturnValue(phrases);

    const { result } = renderHook(() => useFavoritePhrase());

    act(() => {
      result.current.setSearchQuery('会議');
    });

    expect(result.current.filteredPhrases).toHaveLength(1);
    expect(result.current.filteredPhrases[0].originalText).toBe('会議');
  });
});
