import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { BookmarkRepository } from '../BookmarkRepository';

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

describe('BookmarkRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  describe('API正常系', () => {
    it('getAll: APIからブックマークIDリストを取得できる', async () => {
      mockGet.mockResolvedValue({ data: [1, 3, 5] });

      const ids = await BookmarkRepository.getAll();

      expect(ids).toEqual([1, 3, 5]);
      expect(mockGet).toHaveBeenCalledWith('/api/bookmarks');
    });

    it('add: APIでブックマークを追加できる', async () => {
      mockPost.mockResolvedValue({});

      await BookmarkRepository.add(1);

      expect(mockPost).toHaveBeenCalledWith('/api/bookmarks/1');
    });

    it('remove: APIでブックマークを削除できる', async () => {
      mockDelete.mockResolvedValue({});

      await BookmarkRepository.remove(1);

      expect(mockDelete).toHaveBeenCalledWith('/api/bookmarks/1');
    });

    it('isBookmarked: ブックマーク済みか判定できる', async () => {
      mockGet.mockResolvedValue({ data: [1, 3] });

      expect(await BookmarkRepository.isBookmarked(1)).toBe(true);
    });

    it('isBookmarked: 未ブックマークの場合falseを返す', async () => {
      mockGet.mockResolvedValue({ data: [1, 3] });

      expect(await BookmarkRepository.isBookmarked(2)).toBe(false);
    });
  });

  describe('APIエラー時のlocalStorageフォールバック', () => {
    it('getAll: APIエラー時はlocalStorageから取得する', async () => {
      mockGet.mockRejectedValue(new Error('Network Error'));
      localStorage.setItem('freestyle_scenario_bookmarks', JSON.stringify([2, 4]));

      const ids = await BookmarkRepository.getAll();

      expect(ids).toEqual([2, 4]);
    });

    it('add: APIエラー時はlocalStorageに保存する', async () => {
      mockPost.mockRejectedValue(new Error('Network Error'));

      await BookmarkRepository.add(7);

      expect(localStorage.setItem).toHaveBeenCalledWith(
        'freestyle_scenario_bookmarks',
        JSON.stringify([7])
      );
    });

    it('add: APIエラー時に重複追加しない', async () => {
      mockPost.mockRejectedValue(new Error('Network Error'));
      localStorage.setItem('freestyle_scenario_bookmarks', JSON.stringify([7]));

      await BookmarkRepository.add(7);

      const calls = (localStorage.setItem as ReturnType<typeof vi.fn>).mock.calls;
      const lastCall = calls[calls.length - 1];
      expect(JSON.parse(lastCall[1] as string).filter((id: number) => id === 7)).toHaveLength(1);
    });

    it('remove: APIエラー時はlocalStorageから削除する', async () => {
      mockDelete.mockRejectedValue(new Error('Network Error'));
      localStorage.setItem('freestyle_scenario_bookmarks', JSON.stringify([1, 2]));

      await BookmarkRepository.remove(1);

      expect(localStorage.setItem).toHaveBeenCalledWith(
        'freestyle_scenario_bookmarks',
        JSON.stringify([2])
      );
    });
  });
});
