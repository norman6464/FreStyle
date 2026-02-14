import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useNoteEditor } from '../useNoteEditor';
import type { Note } from '../../types';

const mockUpdateNote = vi.fn();

const baseNote: Note = {
  noteId: 'n1',
  userId: 1,
  title: '元タイトル',
  content: '元内容',
  isPinned: false,
  createdAt: 1000,
  updatedAt: 2000,
};

describe('useNoteEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('selectedNoteからeditTitleとeditContentが初期化される', () => {
    const { result } = renderHook(() => useNoteEditor('n1', baseNote, mockUpdateNote));
    expect(result.current.editTitle).toBe('元タイトル');
    expect(result.current.editContent).toBe('元内容');
  });

  it('selectedNoteがnullの場合は空文字で初期化される', () => {
    const { result } = renderHook(() => useNoteEditor(null, null, mockUpdateNote));
    expect(result.current.editTitle).toBe('');
    expect(result.current.editContent).toBe('');
  });

  it('handleTitleChangeでeditTitleが更新される', () => {
    const { result } = renderHook(() => useNoteEditor('n1', baseNote, mockUpdateNote));

    act(() => {
      result.current.handleTitleChange('新タイトル');
    });

    expect(result.current.editTitle).toBe('新タイトル');
  });

  it('handleContentChangeでeditContentが更新される', () => {
    const { result } = renderHook(() => useNoteEditor('n1', baseNote, mockUpdateNote));

    act(() => {
      result.current.handleContentChange('新内容');
    });

    expect(result.current.editContent).toBe('新内容');
  });

  it('タイトル変更後800msでupdateNoteが呼ばれる', () => {
    const { result } = renderHook(() => useNoteEditor('n1', baseNote, mockUpdateNote));

    act(() => {
      result.current.handleTitleChange('新タイトル');
    });

    expect(mockUpdateNote).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(800);
    });

    expect(mockUpdateNote).toHaveBeenCalledWith('n1', {
      title: '新タイトル',
      content: '元内容',
      isPinned: false,
    });
  });

  it('内容変更後800msでupdateNoteが呼ばれる', () => {
    const { result } = renderHook(() => useNoteEditor('n1', baseNote, mockUpdateNote));

    act(() => {
      result.current.handleContentChange('新内容');
    });

    act(() => {
      vi.advanceTimersByTime(800);
    });

    expect(mockUpdateNote).toHaveBeenCalledWith('n1', {
      title: '元タイトル',
      content: '新内容',
      isPinned: false,
    });
  });

  it('連続入力でデバウンスされる', () => {
    const { result } = renderHook(() => useNoteEditor('n1', baseNote, mockUpdateNote));

    act(() => { result.current.handleTitleChange('A'); });
    act(() => { vi.advanceTimersByTime(400); });
    act(() => { result.current.handleTitleChange('AB'); });
    act(() => { vi.advanceTimersByTime(400); });
    act(() => { result.current.handleTitleChange('ABC'); });
    act(() => { vi.advanceTimersByTime(800); });

    expect(mockUpdateNote).toHaveBeenCalledTimes(1);
    expect(mockUpdateNote).toHaveBeenCalledWith('n1', {
      title: 'ABC',
      content: '元内容',
      isPinned: false,
    });
  });

  it('selectedNoteIdが変わるとeditTitle/editContentがリセットされる', () => {
    const newNote: Note = { ...baseNote, noteId: 'n2', title: '別ノート', content: '別内容' };

    const { result, rerender } = renderHook(
      ({ noteId, note }) => useNoteEditor(noteId, note, mockUpdateNote),
      { initialProps: { noteId: 'n1' as string | null, note: baseNote as Note | null } }
    );

    expect(result.current.editTitle).toBe('元タイトル');

    rerender({ noteId: 'n2', note: newNote });

    expect(result.current.editTitle).toBe('別ノート');
    expect(result.current.editContent).toBe('別内容');
  });

  it('selectedNoteIdがnullになるとeditTitle/editContentが空になる', () => {
    const { result, rerender } = renderHook(
      ({ noteId, note }) => useNoteEditor(noteId, note, mockUpdateNote),
      { initialProps: { noteId: 'n1' as string | null, note: baseNote as Note | null } }
    );

    rerender({ noteId: null, note: null });

    expect(result.current.editTitle).toBe('');
    expect(result.current.editContent).toBe('');
  });

  it('ピン留めノートの自動保存でisPinnedがtrueで送信される', () => {
    const pinnedNote: Note = { ...baseNote, isPinned: true };
    const { result } = renderHook(() => useNoteEditor('n1', pinnedNote, mockUpdateNote));

    act(() => { result.current.handleTitleChange('更新'); });
    act(() => { vi.advanceTimersByTime(800); });

    expect(mockUpdateNote).toHaveBeenCalledWith('n1', {
      title: '更新',
      content: '元内容',
      isPinned: true,
    });
  });

  it('selectedNoteIdがnullの場合800ms後もupdateNoteが呼ばれない', () => {
    const { result } = renderHook(() => useNoteEditor(null, null, mockUpdateNote));

    act(() => { result.current.handleTitleChange('テスト'); });
    act(() => { vi.advanceTimersByTime(800); });

    expect(mockUpdateNote).not.toHaveBeenCalled();
  });

  it('タイトルと内容を交互に変更してもデバウンスが正しく動作する', () => {
    const { result } = renderHook(() => useNoteEditor('n1', baseNote, mockUpdateNote));

    act(() => { result.current.handleTitleChange('新タイトル'); });
    act(() => { vi.advanceTimersByTime(400); });
    act(() => { result.current.handleContentChange('新内容'); });
    act(() => { vi.advanceTimersByTime(800); });

    expect(mockUpdateNote).toHaveBeenCalledTimes(1);
    expect(mockUpdateNote).toHaveBeenCalledWith('n1', {
      title: '新タイトル',
      content: '新内容',
      isPinned: false,
    });
  });
});
