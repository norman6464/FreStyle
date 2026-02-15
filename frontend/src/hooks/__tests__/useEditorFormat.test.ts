import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useEditorFormat } from '../useEditorFormat';

describe('useEditorFormat', () => {
  const createMockEditor = () => ({
    chain: vi.fn(() => ({
      focus: vi.fn(() => ({
        toggleBold: vi.fn(() => ({ run: vi.fn() })),
        toggleItalic: vi.fn(() => ({ run: vi.fn() })),
        toggleUnderline: vi.fn(() => ({ run: vi.fn() })),
        toggleStrike: vi.fn(() => ({ run: vi.fn() })),
        setTextAlign: vi.fn(() => ({ run: vi.fn() })),
        setColor: vi.fn(() => ({ run: vi.fn() })),
        unsetColor: vi.fn(() => ({ run: vi.fn() })),
      })),
    })),
  });

  it('全てのハンドラーを返す', () => {
    const editor = createMockEditor();
    const { result } = renderHook(() => useEditorFormat(editor as never));

    expect(result.current.handleBold).toBeDefined();
    expect(result.current.handleItalic).toBeDefined();
    expect(result.current.handleUnderline).toBeDefined();
    expect(result.current.handleStrike).toBeDefined();
    expect(result.current.handleAlign).toBeDefined();
    expect(result.current.handleSelectColor).toBeDefined();
  });

  it('handleBoldがeditor.chain()を呼ぶ', () => {
    const editor = createMockEditor();
    const { result } = renderHook(() => useEditorFormat(editor as never));

    act(() => result.current.handleBold());
    expect(editor.chain).toHaveBeenCalled();
  });

  it('editorがnullでもエラーにならない', () => {
    const { result } = renderHook(() => useEditorFormat(null));

    expect(() => act(() => result.current.handleBold())).not.toThrow();
    expect(() => act(() => result.current.handleAlign('center'))).not.toThrow();
    expect(() => act(() => result.current.handleSelectColor('#ff0000'))).not.toThrow();
  });
});
