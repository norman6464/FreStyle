import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useLinkEditor } from '../useLinkEditor';
import type { Editor } from '@tiptap/core';

function createMockEditor(): Editor {
  const chain = {
    focus: vi.fn().mockReturnThis(),
    extendMarkRange: vi.fn().mockReturnThis(),
    setLink: vi.fn().mockReturnThis(),
    unsetLink: vi.fn().mockReturnThis(),
    run: vi.fn(),
  };
  return {
    getAttributes: vi.fn(() => ({ href: '' })),
    chain: vi.fn(() => chain),
    view: {
      dom: document.createElement('div'),
    },
    _chain: chain,
  } as unknown as Editor & { _chain: typeof chain };
}

describe('useLinkEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('初期状態でlinkBubbleがnull', () => {
    const editor = createMockEditor();
    const containerRef = { current: document.createElement('div') };
    const { result } = renderHook(() => useLinkEditor(editor, containerRef));

    expect(result.current.linkBubble).toBeNull();
  });

  it('handleEditorClickでリンクがないときlinkBubbleがnull', () => {
    const editor = createMockEditor();
    const container = document.createElement('div');
    const containerRef = { current: container };
    const { result } = renderHook(() => useLinkEditor(editor, containerRef));

    const event = { target: document.createElement('p') } as unknown as React.MouseEvent;
    act(() => {
      result.current.handleEditorClick(event);
    });

    expect(result.current.linkBubble).toBeNull();
  });

  it('dismissLinkBubbleでlinkBubbleをnullにする', () => {
    const editor = createMockEditor();
    const containerRef = { current: document.createElement('div') };
    const { result } = renderHook(() => useLinkEditor(editor, containerRef));

    act(() => {
      result.current.dismissLinkBubble();
    });

    expect(result.current.linkBubble).toBeNull();
  });

  it('handleRemoveLinkでunsetLinkが呼ばれる', () => {
    const editor = createMockEditor();
    const containerRef = { current: document.createElement('div') };
    const { result } = renderHook(() => useLinkEditor(editor, containerRef));

    act(() => {
      result.current.handleRemoveLink();
    });

    const chain = (editor as unknown as { _chain: { unsetLink: ReturnType<typeof vi.fn> } })._chain;
    expect(chain.unsetLink).toHaveBeenCalled();
  });
});
