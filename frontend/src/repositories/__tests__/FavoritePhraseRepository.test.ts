import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { FavoritePhraseRepository } from '../FavoritePhraseRepository';

const STORAGE_KEY = 'freestyle_favorite_phrases';

function createMockStorage(): Storage {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
    removeItem: vi.fn((key: string) => { delete store[key]; }),
    clear: vi.fn(() => { store = {}; }),
    get length() { return Object.keys(store).length; },
    key: vi.fn((index: number) => Object.keys(store)[index] ?? null),
  };
}

describe('FavoritePhraseRepository', () => {
  let mockStorage: Storage;

  beforeEach(() => {
    mockStorage = createMockStorage();
    vi.stubGlobal('localStorage', mockStorage);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('初期状態では空配列を返す', () => {
    const result = FavoritePhraseRepository.getAll();
    expect(result).toEqual([]);
  });

  it('フレーズを保存できる', () => {
    FavoritePhraseRepository.save({
      originalText: '確認お願いします',
      rephrasedText: 'ご確認いただけますでしょうか',
      pattern: 'フォーマル',
    });

    const all = FavoritePhraseRepository.getAll();
    expect(all).toHaveLength(1);
    expect(all[0].originalText).toBe('確認お願いします');
    expect(all[0].rephrasedText).toBe('ご確認いただけますでしょうか');
    expect(all[0].pattern).toBe('フォーマル');
    expect(all[0].id).toBeDefined();
    expect(all[0].createdAt).toBeDefined();
  });

  it('複数フレーズを保存でき、新しい順に返される', () => {
    vi.spyOn(Date, 'now')
      .mockReturnValueOnce(1000)
      .mockReturnValueOnce(1000)
      .mockReturnValueOnce(2000)
      .mockReturnValueOnce(2000);

    FavoritePhraseRepository.save({
      originalText: 'テスト1',
      rephrasedText: '言い換え1',
      pattern: 'ソフト',
    });
    FavoritePhraseRepository.save({
      originalText: 'テスト2',
      rephrasedText: '言い換え2',
      pattern: '簡潔',
    });

    const all = FavoritePhraseRepository.getAll();
    expect(all).toHaveLength(2);
    expect(all[0].originalText).toBe('テスト2');
    expect(all[1].originalText).toBe('テスト1');

    vi.restoreAllMocks();
  });

  it('フレーズを削除できる', () => {
    FavoritePhraseRepository.save({
      originalText: 'テスト',
      rephrasedText: '言い換え',
      pattern: 'フォーマル',
    });

    const all = FavoritePhraseRepository.getAll();
    expect(all).toHaveLength(1);

    FavoritePhraseRepository.remove(all[0].id);

    expect(FavoritePhraseRepository.getAll()).toHaveLength(0);
  });

  it('存在しないIDの削除はエラーにならない', () => {
    expect(() => FavoritePhraseRepository.remove('nonexistent')).not.toThrow();
  });

  it('重複チェックが機能する', () => {
    FavoritePhraseRepository.save({
      originalText: 'テスト',
      rephrasedText: '言い換え',
      pattern: 'フォーマル',
    });

    expect(FavoritePhraseRepository.exists('言い換え', 'フォーマル')).toBe(true);
    expect(FavoritePhraseRepository.exists('言い換え', 'ソフト')).toBe(false);
    expect(FavoritePhraseRepository.exists('別のテキスト', 'フォーマル')).toBe(false);
  });
});
