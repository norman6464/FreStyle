import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useBookmark } from '../useBookmark';

const mockGetAll = vi.fn();
const mockAdd = vi.fn();
const mockRemove = vi.fn();
const mockIsBookmarked = vi.fn();

vi.mock('../../repositories/BookmarkRepository', () => ({
  BookmarkRepository: {
    getAll: (...args: unknown[]) => mockGetAll(...args),
    add: (...args: unknown[]) => mockAdd(...args),
    remove: (...args: unknown[]) => mockRemove(...args),
    isBookmarked: (...args: unknown[]) => mockIsBookmarked(...args),
  },
}));

describe('useBookmark', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetAll.mockReturnValue([1, 3]);
    mockIsBookmarked.mockImplementation((id: number) => [1, 3].includes(id));
  });

  it('ブックマーク済みIDの一覧を返す', () => {
    const { result } = renderHook(() => useBookmark());
    expect(result.current.bookmarkedIds).toEqual([1, 3]);
  });

  it('toggleBookmarkでブックマーク追加・削除を切り替える', () => {
    mockGetAll.mockReturnValueOnce([1, 3]).mockReturnValueOnce([1, 3, 5]);

    const { result } = renderHook(() => useBookmark());

    act(() => {
      result.current.toggleBookmark(5);
    });

    expect(mockAdd).toHaveBeenCalledWith(5);
  });

  it('ブックマーク済みの場合はtoggleで削除する', () => {
    mockIsBookmarked.mockReturnValue(true);
    mockGetAll.mockReturnValueOnce([1, 3]).mockReturnValueOnce([3]);

    const { result } = renderHook(() => useBookmark());

    act(() => {
      result.current.toggleBookmark(1);
    });

    expect(mockRemove).toHaveBeenCalledWith(1);
  });

  it('isBookmarkedで判定できる', () => {
    const { result } = renderHook(() => useBookmark());
    expect(result.current.isBookmarked(1)).toBe(true);
    expect(result.current.isBookmarked(2)).toBe(false);
  });

  it('初期ブックマークが空の場合に空配列を返す', () => {
    mockGetAll.mockReturnValue([]);

    const { result } = renderHook(() => useBookmark());
    expect(result.current.bookmarkedIds).toEqual([]);
  });

  it('toggle後にbookmarkedIdsが更新される', () => {
    mockGetAll.mockReturnValueOnce([1]).mockReturnValueOnce([1, 2]);
    mockIsBookmarked.mockReturnValue(false);

    const { result } = renderHook(() => useBookmark());
    expect(result.current.bookmarkedIds).toEqual([1]);

    act(() => {
      result.current.toggleBookmark(2);
    });

    expect(result.current.bookmarkedIds).toEqual([1, 2]);
  });

  it('toggle削除後にbookmarkedIdsが更新される', () => {
    mockGetAll.mockReturnValueOnce([1, 3]).mockReturnValueOnce([3]);
    mockIsBookmarked.mockReturnValue(true);

    const { result } = renderHook(() => useBookmark());
    expect(result.current.bookmarkedIds).toEqual([1, 3]);

    act(() => {
      result.current.toggleBookmark(1);
    });

    expect(result.current.bookmarkedIds).toEqual([3]);
  });
});
