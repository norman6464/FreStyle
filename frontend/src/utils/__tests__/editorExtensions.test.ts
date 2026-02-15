import { describe, it, expect, vi } from 'vitest';

vi.mock('@tiptap/starter-kit', () => ({ default: { configure: vi.fn(() => 'StarterKit') } }));
vi.mock('@tiptap/extension-placeholder', () => ({ default: { configure: vi.fn(() => 'Placeholder') } }));
vi.mock('@tiptap/extension-image', () => ({ default: { configure: vi.fn(() => 'Image') } }));
vi.mock('@tiptap/extension-task-list', () => ({ default: 'TaskList' }));
vi.mock('@tiptap/extension-task-item', () => ({ default: { configure: vi.fn(() => 'TaskItem') } }));
vi.mock('@tiptap/extension-code-block-lowlight', () => ({ default: { configure: vi.fn(() => 'CodeBlockLowlight') } }));
vi.mock('@tiptap/extension-highlight', () => ({ default: { configure: vi.fn(() => 'Highlight') } }));
vi.mock('@tiptap/extension-table', () => ({ Table: { configure: vi.fn(() => 'Table') } }));
vi.mock('@tiptap/extension-table-row', () => ({ TableRow: 'TableRow' }));
vi.mock('@tiptap/extension-table-cell', () => ({ TableCell: 'TableCell' }));
vi.mock('@tiptap/extension-table-header', () => ({ TableHeader: 'TableHeader' }));
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
vi.mock('../../extensions/CodeBlockExtension', () => ({
  CodeBlock: { configure: vi.fn(() => 'CodeBlock') },
}));
vi.mock('@tiptap/extension-link', () => ({ default: { configure: vi.fn(() => 'Link') } }));
vi.mock('@tiptap/extension-text-style', () => ({ TextStyle: 'TextStyle' }));
vi.mock('@tiptap/extension-color', () => ({ default: 'Color' }));
vi.mock('@tiptap/extension-text-align', () => ({ default: { configure: vi.fn(() => 'TextAlign') } }));
vi.mock('@tiptap/extension-underline', () => ({ default: 'Underline' }));
vi.mock('@tiptap/extension-superscript', () => ({ default: 'Superscript' }));
vi.mock('@tiptap/extension-subscript', () => ({ default: 'Subscript' }));
vi.mock('@tiptap/extension-youtube', () => ({ default: { configure: vi.fn(() => 'Youtube') } }));

import { createEditorExtensions } from '../editorExtensions';

describe('createEditorExtensions', () => {
  it('24個のエクステンションを返す', () => {
    const extensions = createEditorExtensions();
    expect(extensions).toHaveLength(24);
  });

  it('主要なエクステンションが含まれる', () => {
    const extensions = createEditorExtensions();
    expect(extensions).toContain('StarterKit');
    expect(extensions).toContain('Placeholder');
    expect(extensions).toContain('Image');
    expect(extensions).toContain('TaskList');
    expect(extensions).toContain('ToggleList');
    expect(extensions).toContain('Callout');
    expect(extensions).toContain('Underline');
    expect(extensions).toContain('Superscript');
    expect(extensions).toContain('Subscript');
  });

  it('毎回新しい配列を返す', () => {
    const a = createEditorExtensions();
    const b = createEditorExtensions();
    expect(a).not.toBe(b);
    expect(a).toEqual(b);
  });
});
