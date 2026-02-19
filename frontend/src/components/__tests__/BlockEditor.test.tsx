import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import BlockEditor from '../BlockEditor';
import type { BlockEditorHandle } from '../BlockEditor';

const mockEditor = {
  chain: vi.fn(() => mockEditor),
  focus: vi.fn(() => mockEditor),
  insertContent: vi.fn(() => mockEditor),
  run: vi.fn(() => true),
  storage: {
    slashCommand: {
      onImageUpload: null as (() => void) | null,
      onEmojiPicker: null as (() => void) | null,
      onYoutubeUrl: null as (() => void) | null,
    },
  },
  commands: {},
  on: vi.fn(),
  off: vi.fn(),
  state: { selection: { from: 0, to: 0, empty: true } },
};

vi.mock('../../hooks/useBlockEditor', () => ({
  useBlockEditor: () => ({ editor: mockEditor }),
}));

vi.mock('../../hooks/useImageUpload', () => ({
  useImageUpload: () => ({
    openFileDialog: vi.fn(),
    handleDrop: vi.fn(),
    handlePaste: vi.fn(),
  }),
}));

vi.mock('../../hooks/useLinkEditor', () => ({
  useLinkEditor: () => ({
    linkBubble: null,
    handleEditorClick: vi.fn(),
    handleEditLink: vi.fn(),
    handleRemoveLink: vi.fn(),
  }),
}));

vi.mock('@tiptap/react', () => ({
  EditorContent: ({ 'aria-label': ariaLabel }: { 'aria-label'?: string }) => (
    <div data-testid="editor-content" aria-label={ariaLabel}>エディタ</div>
  ),
}));

vi.mock('../BlockInserterButton', () => ({
  default: ({ visible }: { visible: boolean }) =>
    visible ? <div data-testid="block-inserter">+</div> : null,
}));

vi.mock('../SearchReplaceBar', () => ({
  default: ({ isOpen }: { isOpen: boolean }) =>
    isOpen ? <div data-testid="search-replace-bar">検索</div> : null,
}));

vi.mock('../SelectionToolbar', () => ({
  default: () => null,
}));

vi.mock('../EmojiPicker', () => ({
  default: ({ onClose }: { onClose: () => void }) => (
    <div data-testid="emoji-picker">
      <button onClick={onClose}>閉じる</button>
    </div>
  ),
}));

vi.mock('../LinkBubbleMenu', () => ({
  default: () => null,
}));

vi.mock('../../extensions/SlashCommandExtension', () => ({
  executeCommand: vi.fn(),
}));

describe('BlockEditor', () => {
  const defaultProps = {
    content: '',
    onChange: vi.fn(),
    noteId: 'test-note-1',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('エディタコンテナが表示される', () => {
    render(<BlockEditor {...defaultProps} />);
    expect(screen.getByTestId('block-editor')).toBeInTheDocument();
  });

  it('EditorContentが表示される', () => {
    render(<BlockEditor {...defaultProps} />);
    expect(screen.getByTestId('editor-content')).toBeInTheDocument();
  });

  it('aria-labelが設定されている', () => {
    render(<BlockEditor {...defaultProps} />);
    expect(screen.getByLabelText('ノートの内容')).toBeInTheDocument();
  });

  it('初期状態で検索バーは非表示', () => {
    render(<BlockEditor {...defaultProps} />);
    expect(screen.queryByTestId('search-replace-bar')).toBeNull();
  });

  it('Ctrl+Fで検索バーが表示される', () => {
    render(<BlockEditor {...defaultProps} />);
    fireEvent.keyDown(document, { key: 'f', ctrlKey: true });
    expect(screen.getByTestId('search-replace-bar')).toBeInTheDocument();
  });

  it('初期状態で絵文字ピッカーは非表示', () => {
    render(<BlockEditor {...defaultProps} />);
    expect(screen.queryByTestId('emoji-picker')).toBeNull();
  });

  it('初期状態でブロック挿入ボタンは非表示', () => {
    render(<BlockEditor {...defaultProps} />);
    expect(screen.queryByTestId('block-inserter')).toBeNull();
  });

  it('data-testid="block-editor"が正しいクラスを持つ', () => {
    render(<BlockEditor {...defaultProps} />);
    const container = screen.getByTestId('block-editor');
    expect(container.className).toContain('block-editor');
    expect(container.className).toContain('pl-10');
  });

  it('contentの先頭でBackspaceを押すとonBackspaceAtStartが呼ばれる', () => {
    const onBackspaceAtStart = vi.fn();
    mockEditor.state.selection = { from: 1, to: 1, empty: true };
    render(<BlockEditor {...defaultProps} onBackspaceAtStart={onBackspaceAtStart} />);
    const container = screen.getByTestId('block-editor');
    fireEvent.keyDown(container, { key: 'Backspace' });
    expect(onBackspaceAtStart).toHaveBeenCalledTimes(1);
  });

  it('contentの途中でBackspaceを押してもonBackspaceAtStartは呼ばれない', () => {
    const onBackspaceAtStart = vi.fn();
    mockEditor.state.selection = { from: 5, to: 5, empty: true };
    render(<BlockEditor {...defaultProps} onBackspaceAtStart={onBackspaceAtStart} />);
    const container = screen.getByTestId('block-editor');
    fireEvent.keyDown(container, { key: 'Backspace' });
    expect(onBackspaceAtStart).not.toHaveBeenCalled();
  });

  it('focusAtStartでエディタの先頭にフォーカスする', () => {
    const ref = React.createRef<BlockEditorHandle>();
    render(<BlockEditor ref={ref} {...defaultProps} />);
    act(() => ref.current!.focusAtStart());
    expect(mockEditor.chain).toHaveBeenCalled();
    expect(mockEditor.focus).toHaveBeenCalledWith('start');
    expect(mockEditor.run).toHaveBeenCalled();
  });

  it('YouTube入力の外側クリックで閉じる', () => {
    render(<BlockEditor {...defaultProps} />);
    const openYoutube = mockEditor.storage.slashCommand.onYoutubeUrl;
    expect(openYoutube).not.toBeNull();
    act(() => openYoutube!());
    expect(screen.getByLabelText('YouTube URL')).toBeInTheDocument();

    fireEvent.mouseDown(document.body);
    expect(screen.queryByLabelText('YouTube URL')).not.toBeInTheDocument();
  });
});
