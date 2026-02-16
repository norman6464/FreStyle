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

  it('handleSelectColorで空文字を渡すとunsetColorが呼ばれる', () => {
    const mockUnsetColor = vi.fn(() => ({ run: vi.fn() }));
    const mockSetColor = vi.fn(() => ({ run: vi.fn() }));
    const editor = {
      chain: vi.fn(() => ({
        focus: vi.fn(() => ({
          setColor: mockSetColor,
          unsetColor: mockUnsetColor,
          toggleBold: vi.fn(() => ({ run: vi.fn() })),
          toggleItalic: vi.fn(() => ({ run: vi.fn() })),
          toggleUnderline: vi.fn(() => ({ run: vi.fn() })),
          toggleStrike: vi.fn(() => ({ run: vi.fn() })),
          setTextAlign: vi.fn(() => ({ run: vi.fn() })),
        })),
      })),
    };

    const { result } = renderHook(() => useEditorFormat(editor as never));
    act(() => result.current.handleSelectColor(''));
    expect(mockUnsetColor).toHaveBeenCalled();
  });

  it('追加のハンドラーがすべて定義されている', () => {
    const editor = createMockEditor();
    const { result } = renderHook(() => useEditorFormat(editor as never));

    expect(result.current.handleHighlight).toBeDefined();
    expect(result.current.handleSuperscript).toBeDefined();
    expect(result.current.handleSubscript).toBeDefined();
    expect(result.current.handleUndo).toBeDefined();
    expect(result.current.handleRedo).toBeDefined();
    expect(result.current.handleClearFormatting).toBeDefined();
    expect(result.current.handleBlockquote).toBeDefined();
    expect(result.current.handleCodeBlock).toBeDefined();
    expect(result.current.handleBulletList).toBeDefined();
    expect(result.current.handleOrderedList).toBeDefined();
    expect(result.current.handleInlineCode).toBeDefined();
    expect(result.current.handleHeading).toBeDefined();
    expect(result.current.handleTaskList).toBeDefined();
    expect(result.current.handleInsertTable).toBeDefined();
  });
});
