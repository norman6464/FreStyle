import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useNoteKeyboardShortcuts } from '../useNoteKeyboardShortcuts';

describe('useNoteKeyboardShortcuts', () => {
  const mockOnCreateNote = vi.fn();
  const mockOnForceSave = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('Ctrl+Nでノート作成が呼ばれる', () => {
    renderHook(() => useNoteKeyboardShortcuts({ onCreateNote: mockOnCreateNote, onForceSave: mockOnForceSave }));

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'n', ctrlKey: true }));

    expect(mockOnCreateNote).toHaveBeenCalledTimes(1);
  });

  it('Meta+Nでノート作成が呼ばれる（Mac）', () => {
    renderHook(() => useNoteKeyboardShortcuts({ onCreateNote: mockOnCreateNote, onForceSave: mockOnForceSave }));

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'n', metaKey: true }));

    expect(mockOnCreateNote).toHaveBeenCalledTimes(1);
  });

  it('Ctrl+Sで即時保存が呼ばれる', () => {
    renderHook(() => useNoteKeyboardShortcuts({ onCreateNote: mockOnCreateNote, onForceSave: mockOnForceSave }));

    const event = new KeyboardEvent('keydown', { key: 's', ctrlKey: true, cancelable: true });
    const preventDefaultSpy = vi.spyOn(event, 'preventDefault');
    window.dispatchEvent(event);

    expect(mockOnForceSave).toHaveBeenCalledTimes(1);
    expect(preventDefaultSpy).toHaveBeenCalled();
  });

  it('Meta+Sで即時保存が呼ばれる（Mac）', () => {
    renderHook(() => useNoteKeyboardShortcuts({ onCreateNote: mockOnCreateNote, onForceSave: mockOnForceSave }));

    const event = new KeyboardEvent('keydown', { key: 's', metaKey: true, cancelable: true });
    const preventDefaultSpy = vi.spyOn(event, 'preventDefault');
    window.dispatchEvent(event);

    expect(mockOnForceSave).toHaveBeenCalledTimes(1);
    expect(preventDefaultSpy).toHaveBeenCalled();
  });

  it('修飾キーなしのNキーではノート作成が呼ばれない', () => {
    renderHook(() => useNoteKeyboardShortcuts({ onCreateNote: mockOnCreateNote, onForceSave: mockOnForceSave }));

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'n' }));

    expect(mockOnCreateNote).not.toHaveBeenCalled();
  });

  it('アンマウント時にイベントリスナーが解除される', () => {
    const { unmount } = renderHook(() => useNoteKeyboardShortcuts({ onCreateNote: mockOnCreateNote, onForceSave: mockOnForceSave }));

    unmount();

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'n', ctrlKey: true }));

    expect(mockOnCreateNote).not.toHaveBeenCalled();
  });
});
