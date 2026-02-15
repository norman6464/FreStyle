import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import NoteEditor from '../NoteEditor';

vi.mock('../../hooks/useBlockEditor', () => ({
  useBlockEditor: () => ({ editor: null }),
}));

const defaultProps = {
  title: 'テストノート',
  content: '',
  noteId: 'note-1',
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
});
