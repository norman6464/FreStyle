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
});
