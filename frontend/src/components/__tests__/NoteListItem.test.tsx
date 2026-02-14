import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import NoteListItem from '../NoteListItem';

const defaultProps = {
  noteId: 'note-1',
  title: 'テストノート',
  content: 'テスト内容のプレビュー',
  updatedAt: 1707900000000,
  isPinned: false,
  isActive: false,
  onSelect: vi.fn(),
  onDelete: vi.fn(),
  onTogglePin: vi.fn(),
};

describe('NoteListItem', () => {
  it('タイトルを表示する', () => {
    render(<NoteListItem {...defaultProps} />);
    expect(screen.getByText('テストノート')).toBeInTheDocument();
  });

  it('内容のプレビューを表示する', () => {
    render(<NoteListItem {...defaultProps} />);
    expect(screen.getByText('テスト内容のプレビュー')).toBeInTheDocument();
  });

  it('クリックでonSelectが呼ばれる', () => {
    render(<NoteListItem {...defaultProps} />);
    fireEvent.click(screen.getByText('テストノート'));
    expect(defaultProps.onSelect).toHaveBeenCalledWith('note-1');
  });

  it('削除ボタンでonDeleteが呼ばれる', () => {
    render(<NoteListItem {...defaultProps} />);
    const deleteBtn = screen.getByLabelText('ノートを削除');
    fireEvent.click(deleteBtn);
    expect(defaultProps.onDelete).toHaveBeenCalledWith('note-1');
  });

  it('アクティブ状態でスタイルが変わる', () => {
    const { container } = render(<NoteListItem {...defaultProps} isActive={true} />);
    const item = container.querySelector('[role="button"]');
    expect(item?.className).toContain('bg-surface-2');
  });

  it('タイトルが空の場合は「無題」を表示する', () => {
    render(<NoteListItem {...defaultProps} title="" />);
    expect(screen.getByText('無題')).toBeInTheDocument();
  });

  it('ピン留めボタンが表示される', () => {
    render(<NoteListItem {...defaultProps} />);
    expect(screen.getByLabelText('ピン留め')).toBeInTheDocument();
  });

  it('ピン留め状態のときピンアイコンがソリッドになる', () => {
    render(<NoteListItem {...defaultProps} isPinned={true} />);
    expect(screen.getByLabelText('ピン留め解除')).toBeInTheDocument();
  });

  it('ピン留めボタンクリックでonTogglePinが呼ばれる', () => {
    const onTogglePin = vi.fn();
    render(<NoteListItem {...defaultProps} onTogglePin={onTogglePin} />);
    fireEvent.click(screen.getByLabelText('ピン留め'));
    expect(onTogglePin).toHaveBeenCalledWith('note-1');
  });

  it('ピン留めボタンクリックでonSelectは呼ばれない', () => {
    const onSelect = vi.fn();
    const onTogglePin = vi.fn();
    render(<NoteListItem {...defaultProps} onSelect={onSelect} onTogglePin={onTogglePin} />);
    fireEvent.click(screen.getByLabelText('ピン留め'));
    expect(onSelect).not.toHaveBeenCalled();
  });
});
