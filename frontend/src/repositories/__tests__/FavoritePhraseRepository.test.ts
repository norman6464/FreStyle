import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { FavoritePhraseRepository } from '../FavoritePhraseRepository';

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockDelete = vi.fn();

vi.mock('../../lib/axios', () => ({
  default: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
    delete: (...args: unknown[]) => mockDelete(...args),
  },
}));

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
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  describe('API正常系', () => {
    it('getAll: APIからフレーズ一覧を取得し、IDをstring変換する', async () => {
      mockGet.mockResolvedValue({
        data: [
          { id: 1, originalText: '元文', rephrasedText: '変換文', pattern: 'フォーマル版', createdAt: '2026-01-01T00:00:00Z' },
        ],
      });

      const result = await FavoritePhraseRepository.getAll();

      expect(result).toHaveLength(1);
      expect(result[0].id).toBe('1');
      expect(result[0].originalText).toBe('元文');
      expect(mockGet).toHaveBeenCalledWith('/api/favorite-phrases');
    });

    it('save: APIでフレーズを保存できる', async () => {
      mockPost.mockResolvedValue({});

      await FavoritePhraseRepository.save({
        originalText: '確認お願い',
        rephrasedText: 'ご確認ください',
        pattern: 'フォーマル版',
      });

      expect(mockPost).toHaveBeenCalledWith('/api/favorite-phrases', {
        originalText: '確認お願い',
        rephrasedText: 'ご確認ください',
        pattern: 'フォーマル版',
      });
    });

    it('remove: APIでフレーズを削除できる', async () => {
      mockDelete.mockResolvedValue({});

      await FavoritePhraseRepository.remove('5');

      expect(mockDelete).toHaveBeenCalledWith('/api/favorite-phrases/5');
    });

    it('exists: フレーズの存在判定ができる', async () => {
      mockGet.mockResolvedValue({
        data: [
          { id: 1, originalText: '元', rephrasedText: '変換文', pattern: 'フォーマル版', createdAt: '2026-01-01' },
        ],
      });

      expect(await FavoritePhraseRepository.exists('変換文', 'フォーマル版')).toBe(true);
      expect(await FavoritePhraseRepository.exists('別文', 'フォーマル版')).toBe(false);
    });
  });

  describe('APIエラー時のlocalStorageフォールバック', () => {
    it('getAll: APIエラー時はlocalStorageから取得する', async () => {
      mockGet.mockRejectedValue(new Error('Network Error'));
      localStorage.setItem('freestyle_favorite_phrases', JSON.stringify([
        { id: 'fav_1', originalText: 'テスト', rephrasedText: '変換', pattern: 'ソフト版', createdAt: '2026-01-01' },
      ]));

      const result = await FavoritePhraseRepository.getAll();

      expect(result).toHaveLength(1);
      expect(result[0].id).toBe('fav_1');
    });

    it('save: APIエラー時はlocalStorageに保存する', async () => {
      mockPost.mockRejectedValue(new Error('Network Error'));

      await FavoritePhraseRepository.save({
        originalText: 'テスト',
        rephrasedText: '変換',
        pattern: 'ソフト版',
      });

      expect(localStorage.setItem).toHaveBeenCalled();
    });

    it('remove: APIエラー時はlocalStorageから削除する', async () => {
      mockDelete.mockRejectedValue(new Error('Network Error'));
      localStorage.setItem('freestyle_favorite_phrases', JSON.stringify([
        { id: 'fav_1', originalText: 'テスト', rephrasedText: '変換', pattern: 'ソフト版', createdAt: '2026-01-01' },
      ]));

      await FavoritePhraseRepository.remove('fav_1');

      const calls = (localStorage.setItem as ReturnType<typeof vi.fn>).mock.calls;
      const lastCall = calls[calls.length - 1];
      expect(JSON.parse(lastCall[1] as string)).toHaveLength(0);
    });
  });
});
