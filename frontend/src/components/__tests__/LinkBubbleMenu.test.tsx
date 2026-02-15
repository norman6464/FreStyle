import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import LinkBubbleMenu from '../LinkBubbleMenu';

describe('LinkBubbleMenu', () => {
  const defaultProps = {
    url: 'https://example.com',
    onEdit: vi.fn(),
    onRemove: vi.fn(),
  };

  it('URLが表示される', () => {
    render(<LinkBubbleMenu {...defaultProps} />);
    expect(screen.getByText('https://example.com')).toBeInTheDocument();
  });

  it('編集ボタンが表示される', () => {
    render(<LinkBubbleMenu {...defaultProps} />);
    expect(screen.getByLabelText('リンクを編集')).toBeInTheDocument();
  });

  it('削除ボタンが表示される', () => {
    render(<LinkBubbleMenu {...defaultProps} />);
    expect(screen.getByLabelText('リンクを削除')).toBeInTheDocument();
  });

  it('編集ボタンクリックでonEditが呼ばれる', () => {
    const onEdit = vi.fn();
    render(<LinkBubbleMenu {...defaultProps} onEdit={onEdit} />);
    fireEvent.click(screen.getByLabelText('リンクを編集'));
    expect(onEdit).toHaveBeenCalled();
  });

  it('削除ボタンクリックでonRemoveが呼ばれる', () => {
    const onRemove = vi.fn();
    render(<LinkBubbleMenu {...defaultProps} onRemove={onRemove} />);
    fireEvent.click(screen.getByLabelText('リンクを削除'));
    expect(onRemove).toHaveBeenCalled();
  });

  it('URLリンクがhref属性を持つ', () => {
    render(<LinkBubbleMenu {...defaultProps} />);
    const link = screen.getByText('https://example.com');
    expect(link.closest('a')).toHaveAttribute('href', 'https://example.com');
  });
});
