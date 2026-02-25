import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useBookmark } from '../useBookmark';

const mockGetAll = vi.fn();
const mockAdd = vi.fn();
const mockRemove = vi.fn();

vi.mock('../../repositories/BookmarkRepository', () => ({
  BookmarkRepository: {
    getAll: (...args: unknown[]) => mockGetAll(...args),
    add: (...args: unknown[]) => mockAdd(...args),
    remove: (...args: unknown[]) => mockRemove(...args),
  },
}));

describe('useBookmark', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetAll.mockResolvedValue([1, 3]);
    mockAdd.mockResolvedValue(undefined);
    mockRemove.mockResolvedValue(undefined);
  });

  it('初期状態はloading=trueで空配列', () => {
    const { result } = renderHook(() => useBookmark());
    expect(result.current.bookmarkedIds).toEqual([]);
    expect(result.current.loading).toBe(true);
  });

  it('マウント時にAPIからブックマークを取得する', async () => {
    const { result } = renderHook(() => useBookmark());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.bookmarkedIds).toEqual([1, 3]);
    expect(mockGetAll).toHaveBeenCalled();
  });

  it('toggleBookmarkでブックマーク追加できる', async () => {
    const { result } = renderHook(() => useBookmark());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.toggleBookmark(5);
    });

    expect(mockAdd).toHaveBeenCalledWith(5);
    expect(result.current.bookmarkedIds).toContain(5);
  });

  it('ブックマーク済みの場合はtoggleで削除する', async () => {
    const { result } = renderHook(() => useBookmark());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.toggleBookmark(1);
    });

    expect(mockRemove).toHaveBeenCalledWith(1);
    expect(result.current.bookmarkedIds).not.toContain(1);
  });

  it('isBookmarkedで判定できる', async () => {
    const { result } = renderHook(() => useBookmark());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.isBookmarked(1)).toBe(true);
    expect(result.current.isBookmarked(2)).toBe(false);
  });

  it('初期ブックマークが空の場合に空配列を返す', async () => {
    mockGetAll.mockResolvedValue([]);

    const { result } = renderHook(() => useBookmark());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.bookmarkedIds).toEqual([]);
  });

  it('toggle追加後にbookmarkedIdsが即座に更新される', async () => {
    const { result } = renderHook(() => useBookmark());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.toggleBookmark(2);
    });

    expect(result.current.bookmarkedIds).toEqual([1, 3, 2]);
  });

  it('toggle削除後にbookmarkedIdsが即座に更新される', async () => {
    const { result } = renderHook(() => useBookmark());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.toggleBookmark(1);
    });

    expect(result.current.bookmarkedIds).toEqual([3]);
  });

  it('loading状態を返す', async () => {
    const { result } = renderHook(() => useBookmark());

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('ブックマーク追加API失敗時に状態がロールバックされる', async () => {
    mockAdd.mockRejectedValue(new Error('API error'));
    const { result } = renderHook(() => useBookmark());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.toggleBookmark(5);
    });

    expect(result.current.bookmarkedIds).not.toContain(5);
  });

  it('ブックマーク削除API失敗時に状態がロールバックされる', async () => {
    mockRemove.mockRejectedValue(new Error('API error'));
    const { result } = renderHook(() => useBookmark());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.toggleBookmark(1);
    });

    expect(result.current.bookmarkedIds).toContain(1);
  });
});
