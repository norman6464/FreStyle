import { describe, it, expect, vi, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSessionNote } from '../useSessionNote';
import { SessionNoteRepository } from '../../repositories/SessionNoteRepository';

vi.mock('../../repositories/SessionNoteRepository');

const mockedRepo = vi.mocked(SessionNoteRepository);

describe('useSessionNote', () => {
  afterEach(() => {
    vi.clearAllMocks();
    mockedRepo.get.mockReset();
    mockedRepo.save.mockReset();
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

  it('sessionIdがnullの場合は空文字を返す', () => {
    const { result } = renderHook(() => useSessionNote(null));

    expect(result.current.note).toBe('');
    expect(mockedRepo.get).not.toHaveBeenCalled();
  });

  it('sessionIdがnullの場合は保存しない', () => {
    const { result } = renderHook(() => useSessionNote(null));

    act(() => {
      result.current.saveNote('テスト');
    });

    expect(mockedRepo.save).not.toHaveBeenCalled();
  });

  it('保存後にnoteが即座に更新される', () => {
    mockedRepo.get.mockReturnValue({ sessionId: 2, note: '元のメモ', updatedAt: '2026-02-13' });

    const { result } = renderHook(() => useSessionNote(2));

    expect(result.current.note).toBe('元のメモ');

    act(() => {
      result.current.saveNote('更新メモ');
    });

    expect(result.current.note).toBe('更新メモ');
  });

  it('複数回保存できる', () => {
    mockedRepo.get.mockReturnValue(null);

    const { result } = renderHook(() => useSessionNote(1));

    act(() => {
      result.current.saveNote('メモ1');
    });
    expect(result.current.note).toBe('メモ1');

    act(() => {
      result.current.saveNote('メモ2');
    });
    expect(result.current.note).toBe('メモ2');
    expect(mockedRepo.save).toHaveBeenCalledTimes(2);
  });

  it('空文字で保存できる', () => {
    mockedRepo.get.mockReturnValue({ sessionId: 1, note: '既存メモ', updatedAt: '2026-02-13' });

    const { result } = renderHook(() => useSessionNote(1));

    act(() => {
      result.current.saveNote('');
    });

    expect(mockedRepo.save).toHaveBeenCalledWith(1, '');
    expect(result.current.note).toBe('');
  });
});
