import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useBlockEditor } from '../useBlockEditor';

vi.mock('@tiptap/react', () => ({
  useEditor: vi.fn(() => mockEditor),
}));

vi.mock('@tiptap/starter-kit', () => ({ default: { configure: vi.fn(() => 'StarterKit') } }));
vi.mock('@tiptap/extension-placeholder', () => ({ default: { configure: vi.fn(() => 'Placeholder') } }));
vi.mock('@tiptap/extension-image', () => ({ default: { configure: vi.fn(() => 'Image') } }));
vi.mock('@tiptap/extension-task-list', () => ({ default: 'TaskList' }));
vi.mock('@tiptap/extension-task-item', () => ({ default: { configure: vi.fn(() => 'TaskItem') } }));
vi.mock('@tiptap/extension-code-block-lowlight', () => ({ default: { configure: vi.fn(() => 'CodeBlockLowlight') } }));
vi.mock('@tiptap/extension-highlight', () => ({ default: { configure: vi.fn(() => 'Highlight') } }));
vi.mock('@tiptap/extension-table', () => ({ default: { configure: vi.fn(() => 'Table') } }));
vi.mock('@tiptap/extension-table-row', () => ({ default: 'TableRow' }));
vi.mock('@tiptap/extension-table-cell', () => ({ default: 'TableCell' }));
vi.mock('@tiptap/extension-table-header', () => ({ default: 'TableHeader' }));
vi.mock('lowlight', () => ({ common: {}, createLowlight: vi.fn(() => 'lowlight') }));
vi.mock('../../extensions/SlashCommandExtension', () => ({
  SlashCommandExtension: { configure: vi.fn(() => 'SlashCommand') },
  executeCommand: vi.fn(),
}));
vi.mock('../../extensions/slashCommandRenderer', () => ({
  slashCommandRenderer: vi.fn(),
}));
vi.mock('../../extensions/ToggleListExtension', () => ({
  ToggleList: 'ToggleList',
  ToggleSummary: 'ToggleSummary',
  ToggleContent: 'ToggleContent',
}));
vi.mock('../../extensions/CalloutExtension', () => ({
  Callout: 'Callout',
}));
vi.mock('@tiptap/extension-link', () => ({ default: { configure: vi.fn(() => 'Link') } }));
vi.mock('@tiptap/extension-text-style', () => ({ default: 'TextStyle' }));
vi.mock('@tiptap/extension-color', () => ({ default: 'Color' }));
vi.mock('@tiptap/extension-text-align', () => ({ default: { configure: vi.fn(() => 'TextAlign') } }));
vi.mock('@tiptap/extension-underline', () => ({ default: 'Underline' }));
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

  it('useEditorにextensionsを渡して呼び出す', async () => {
    const { useEditor } = await import('@tiptap/react');
    const onChange = vi.fn();
    renderHook(() => useBlockEditor({ content: '', onChange }));

    expect(useEditor).toHaveBeenCalledWith(
      expect.objectContaining({
        extensions: expect.arrayContaining([
          'StarterKit',
          'CodeBlockLowlight',
          'Placeholder',
          'Image',
          'SlashCommand',
          'Highlight',
          'TaskList',
          'TaskItem',
          'ToggleList',
          'ToggleSummary',
          'ToggleContent',
        ]),
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

  it('21個の拡張が設定される', async () => {
    const { useEditor } = await import('@tiptap/react');
    const onChange = vi.fn();
    renderHook(() => useBlockEditor({ content: '', onChange }));

    const call = vi.mocked(useEditor).mock.calls[0]!;
    expect(call[0]!.extensions).toHaveLength(21);
  });
});
