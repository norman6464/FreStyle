import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import NoteEditor from '../NoteEditor';

const defaultProps = {
  title: 'テストノート',
  content: 'テスト内容',
  onTitleChange: vi.fn(),
  onContentChange: vi.fn(),
};

describe('NoteEditor', () => {
  it('タイトル入力欄を表示する', () => {
    render(<NoteEditor {...defaultProps} />);
    const input = screen.getByDisplayValue('テストノート');
    expect(input).toBeInTheDocument();
  });

  it('内容入力欄を表示する', () => {
    render(<NoteEditor {...defaultProps} />);
    const textarea = screen.getByDisplayValue('テスト内容');
    expect(textarea).toBeInTheDocument();
  });

  it('タイトル変更でonTitleChangeが呼ばれる', () => {
    render(<NoteEditor {...defaultProps} />);
    const input = screen.getByDisplayValue('テストノート');
    fireEvent.change(input, { target: { value: '更新タイトル' } });
    expect(defaultProps.onTitleChange).toHaveBeenCalledWith('更新タイトル');
  });

  it('内容変更でonContentChangeが呼ばれる', () => {
    render(<NoteEditor {...defaultProps} />);
    const textarea = screen.getByDisplayValue('テスト内容');
    fireEvent.change(textarea, { target: { value: '更新内容' } });
    expect(defaultProps.onContentChange).toHaveBeenCalledWith('更新内容');
  });

  it('タイトルのプレースホルダーが表示される', () => {
    render(<NoteEditor {...defaultProps} title="" />);
    expect(screen.getByPlaceholderText('無題')).toBeInTheDocument();
  });

  it('内容のプレースホルダーが表示される', () => {
    render(<NoteEditor {...defaultProps} content="" />);
    expect(screen.getByPlaceholderText('ここに入力...')).toBeInTheDocument();
  });

  it('文字数カウントが表示される', () => {
    render(<NoteEditor {...defaultProps} content="テスト内容" />);
    expect(screen.getByText('5文字')).toBeInTheDocument();
  });

  it('読了時間が表示される', () => {
    render(<NoteEditor {...defaultProps} content="テスト内容" />);
    expect(screen.getByText(/約\d+分/)).toBeInTheDocument();
  });

  it('空の内容で0文字と表示される', () => {
    render(<NoteEditor {...defaultProps} content="" />);
    expect(screen.getByText('0文字')).toBeInTheDocument();
  });
});
