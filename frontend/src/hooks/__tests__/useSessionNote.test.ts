import { describe, it, expect, vi, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSessionNote } from '../useSessionNote';
import { SessionNoteRepository } from '../../repositories/SessionNoteRepository';

vi.mock('../../repositories/SessionNoteRepository');

const mockedRepo = vi.mocked(SessionNoteRepository);

describe('useSessionNote', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('セッションIDでメモを取得する', () => {
    mockedRepo.get.mockReturnValue({ sessionId: 1, note: 'テストメモ', updatedAt: '2026-02-13' });

    const { result } = renderHook(() => useSessionNote(1));

    expect(result.current.note).toBe('テストメモ');
    expect(mockedRepo.get).toHaveBeenCalledWith(1);
  });

  it('メモがない場合は空文字', () => {
    mockedRepo.get.mockReturnValue(null);

    const { result } = renderHook(() => useSessionNote(1));

    expect(result.current.note).toBe('');
  });

  it('メモを保存できる', () => {
    mockedRepo.get
      .mockReturnValueOnce(null)
      .mockReturnValueOnce({ sessionId: 1, note: '新しいメモ', updatedAt: '2026-02-13' });

    const { result } = renderHook(() => useSessionNote(1));

    act(() => {
      result.current.saveNote('新しいメモ');
    });

    expect(mockedRepo.save).toHaveBeenCalledWith(1, '新しいメモ');
    expect(result.current.note).toBe('新しいメモ');
  });
});
