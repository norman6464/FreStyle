import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import NoteEditor from '../NoteEditor';
import type { SaveStatus } from '../../hooks/useNoteEditor';

vi.mock('../../hooks/useBlockEditor', () => ({
  useBlockEditor: () => ({ editor: null }),
}));

vi.mock('../../hooks/useToast', () => ({
  useToast: () => ({ showToast: vi.fn(), toasts: [], removeToast: vi.fn() }),
}));

const defaultProps = {
  title: 'テストノート',
  content: '',
  noteId: 'note-1',
  saveStatus: 'idle' as SaveStatus,
  onTitleChange: vi.fn(),
  onContentChange: vi.fn(),
};

describe('NoteEditor', () => {
  it('タイトル入力欄を表示する', () => {
    render(<NoteEditor {...defaultProps} />);
    const input = screen.getByDisplayValue('テストノート');
    expect(input).toBeInTheDocument();
  });

  it('タイトル変更でonTitleChangeが呼ばれる', () => {
    render(<NoteEditor {...defaultProps} />);
    const input = screen.getByDisplayValue('テストノート');
    fireEvent.change(input, { target: { value: '更新タイトル' } });
    expect(defaultProps.onTitleChange).toHaveBeenCalledWith('更新タイトル');
  });

  it('タイトルのプレースホルダーが表示される', () => {
    render(<NoteEditor {...defaultProps} title="" />);
    expect(screen.getByPlaceholderText('無題')).toBeInTheDocument();
  });

  it('ブロックエディタが表示される', () => {
    render(<NoteEditor {...defaultProps} />);
    expect(screen.getByTestId('block-editor')).toBeInTheDocument();
  });

  it('文字数カウントが表示される', () => {
    render(<NoteEditor {...defaultProps} content="" />);
    expect(screen.getByText('0文字')).toBeInTheDocument();
  });

  it('ノート統計が表示される', () => {
    render(<NoteEditor {...defaultProps} />);
    expect(screen.getByLabelText('ノート統計')).toBeInTheDocument();
  });

  it('saveStatusがidleのとき保存インジケーターが表示されない', () => {
    render(<NoteEditor {...defaultProps} saveStatus="idle" />);
    expect(screen.queryByLabelText('保存状態')).not.toBeInTheDocument();
  });

  it('saveStatusがunsavedのとき「未保存」が表示される', () => {
    render(<NoteEditor {...defaultProps} saveStatus="unsaved" />);
    expect(screen.getByText('未保存')).toBeInTheDocument();
  });

  it('saveStatusがsavingのとき「保存中...」が表示される', () => {
    render(<NoteEditor {...defaultProps} saveStatus="saving" />);
    expect(screen.getByText('保存中...')).toBeInTheDocument();
  });

  it('saveStatusがsavedのとき「保存済み」が表示される', () => {
    render(<NoteEditor {...defaultProps} saveStatus="saved" />);
    expect(screen.getByText('保存済み')).toBeInTheDocument();
  });
});
