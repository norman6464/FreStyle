import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { BookmarkRepository } from '../BookmarkRepository';

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
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('ブックマークを追加できる', () => {
    BookmarkRepository.add(1);
    const ids = BookmarkRepository.getAll();
    expect(ids).toContain(1);
  });

  it('ブックマークを削除できる', () => {
    BookmarkRepository.add(1);
    BookmarkRepository.add(2);
    BookmarkRepository.remove(1);
    const ids = BookmarkRepository.getAll();
    expect(ids).not.toContain(1);
    expect(ids).toContain(2);
  });

  it('ブックマーク済みか判定できる', () => {
    BookmarkRepository.add(5);
    expect(BookmarkRepository.isBookmarked(5)).toBe(true);
    expect(BookmarkRepository.isBookmarked(6)).toBe(false);
  });

  it('重複追加しても1つのみ保存される', () => {
    BookmarkRepository.add(1);
    BookmarkRepository.add(1);
    const ids = BookmarkRepository.getAll();
    expect(ids.filter(id => id === 1)).toHaveLength(1);
  });
});
