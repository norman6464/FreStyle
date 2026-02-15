import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useBlockEditor } from '../useBlockEditor';

vi.mock('@tiptap/react', () => ({
  useEditor: vi.fn(() => mockEditor),
}));

const mockExtensions = ['ext1', 'ext2'];
vi.mock('../../utils/editorExtensions', () => ({
  createEditorExtensions: vi.fn(() => mockExtensions),
}));
vi.mock('../../utils/isLegacyMarkdown', () => ({
  isLegacyMarkdown: vi.fn(() => false),
}));
vi.mock('../../utils/markdownToTiptap', () => ({
  markdownToTiptap: vi.fn(),
}));

const mockEditor = {
  getJSON: vi.fn(() => ({ type: 'doc', content: [] })),
  commands: {
    clearContent: vi.fn(),
    setContent: vi.fn(),
  },
};

describe('useBlockEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('editorインスタンスを返す', () => {
    const onChange = vi.fn();
    const { result } = renderHook(() => useBlockEditor({ content: '', onChange }));

    expect(result.current.editor).toBe(mockEditor);
  });

  it('createEditorExtensionsの結果をuseEditorに渡す', async () => {
    const { useEditor } = await import('@tiptap/react');
    const onChange = vi.fn();
    renderHook(() => useBlockEditor({ content: '', onChange }));

    expect(useEditor).toHaveBeenCalledWith(
      expect.objectContaining({
        extensions: mockExtensions,
      })
    );
  });

  it('JSONコンテンツをパースしてinitialContentとして設定する', async () => {
    const { useEditor } = await import('@tiptap/react');
    const jsonContent = JSON.stringify({ type: 'doc', content: [{ type: 'paragraph' }] });
    const onChange = vi.fn();
    renderHook(() => useBlockEditor({ content: jsonContent, onChange }));

    expect(useEditor).toHaveBeenCalledWith(
      expect.objectContaining({
        content: { type: 'doc', content: [{ type: 'paragraph' }] },
      })
    );
  });

  it('空コンテンツの場合はundefinedを設定する', async () => {
    const { useEditor } = await import('@tiptap/react');
    const onChange = vi.fn();
    renderHook(() => useBlockEditor({ content: '', onChange }));

    expect(useEditor).toHaveBeenCalledWith(
      expect.objectContaining({
        content: undefined,
      })
    );
  });

  it('レガシーMarkdownコンテンツをTipTap JSONに変換する', async () => {
    const { isLegacyMarkdown } = await import('../../utils/isLegacyMarkdown');
    const { markdownToTiptap } = await import('../../utils/markdownToTiptap');
    const { useEditor } = await import('@tiptap/react');

    vi.mocked(isLegacyMarkdown).mockReturnValue(true);
    const tiptapDoc = { type: 'doc', content: [{ type: 'paragraph' }] };
    vi.mocked(markdownToTiptap).mockReturnValue(tiptapDoc);

    const onChange = vi.fn();
    renderHook(() => useBlockEditor({ content: '# Hello', onChange }));

    expect(isLegacyMarkdown).toHaveBeenCalledWith('# Hello');
    expect(markdownToTiptap).toHaveBeenCalledWith('# Hello');
    expect(useEditor).toHaveBeenCalledWith(
      expect.objectContaining({ content: tiptapDoc })
    );
  });

  it('不正なJSONの場合はundefinedを設定する', async () => {
    const { useEditor } = await import('@tiptap/react');
    const { isLegacyMarkdown } = await import('../../utils/isLegacyMarkdown');
    vi.mocked(isLegacyMarkdown).mockReturnValue(false);

    const onChange = vi.fn();
    renderHook(() => useBlockEditor({ content: 'invalid json{{{', onChange }));

    expect(useEditor).toHaveBeenCalledWith(
      expect.objectContaining({ content: undefined })
    );
  });

  it('onUpdateコールバックがuseEditorに設定される', async () => {
    const { useEditor } = await import('@tiptap/react');
    const onChange = vi.fn();
    renderHook(() => useBlockEditor({ content: '', onChange }));

    const call = vi.mocked(useEditor).mock.calls[0]!;
    expect(call[0]).toHaveProperty('onUpdate');
  });
});
