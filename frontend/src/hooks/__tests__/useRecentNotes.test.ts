import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useRecentNotes } from '../useRecentNotes';
import { SessionNoteRepository } from '../../repositories/SessionNoteRepository';

vi.mock('../../repositories/SessionNoteRepository', () => ({
  SessionNoteRepository: {
    getAll: vi.fn(),
  },
}));

const mockedRepo = vi.mocked(SessionNoteRepository);

describe('useRecentNotes', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('メモがない場合は空配列を返す', () => {
    mockedRepo.getAll.mockReturnValue({});

    const { result } = renderHook(() => useRecentNotes());

    expect(result.current.notes).toEqual([]);
    expect(result.current.totalCount).toBe(0);
  });

  it('メモを更新日時の降順でソートして返す', () => {
    mockedRepo.getAll.mockReturnValue({
      '1': { sessionId: 1, note: '古いメモ', updatedAt: '2024-01-01T00:00:00Z' },
      '2': { sessionId: 2, note: '新しいメモ', updatedAt: '2024-06-01T00:00:00Z' },
      '3': { sessionId: 3, note: '中間のメモ', updatedAt: '2024-03-01T00:00:00Z' },
    });

    const { result } = renderHook(() => useRecentNotes());

    expect(result.current.notes[0].sessionId).toBe(2);
    expect(result.current.notes[1].sessionId).toBe(3);
    expect(result.current.notes[2].sessionId).toBe(1);
  });

  it('limitパラメータで返すメモ数を制限する', () => {
    mockedRepo.getAll.mockReturnValue({
      '1': { sessionId: 1, note: 'メモ1', updatedAt: '2024-01-01T00:00:00Z' },
      '2': { sessionId: 2, note: 'メモ2', updatedAt: '2024-02-01T00:00:00Z' },
      '3': { sessionId: 3, note: 'メモ3', updatedAt: '2024-03-01T00:00:00Z' },
      '4': { sessionId: 4, note: 'メモ4', updatedAt: '2024-04-01T00:00:00Z' },
    });

    const { result } = renderHook(() => useRecentNotes(2));

    expect(result.current.notes).toHaveLength(2);
    expect(result.current.totalCount).toBe(4);
  });

  it('totalCountは全メモ数を返す', () => {
    mockedRepo.getAll.mockReturnValue({
      '1': { sessionId: 1, note: 'メモ1', updatedAt: '2024-01-01T00:00:00Z' },
      '2': { sessionId: 2, note: 'メモ2', updatedAt: '2024-02-01T00:00:00Z' },
      '3': { sessionId: 3, note: 'メモ3', updatedAt: '2024-03-01T00:00:00Z' },
    });

    const { result } = renderHook(() => useRecentNotes(1));

    expect(result.current.notes).toHaveLength(1);
    expect(result.current.totalCount).toBe(3);
  });
});
