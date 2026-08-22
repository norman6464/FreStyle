import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import DocumentListItem from '../DocumentListItem';

const props = {
  id: 'a',
  title: 'メモA',
  updatedAt: '2026-08-02T00:00:00Z',
  isActive: false,
  onSelect: vi.fn(),
  onDelete: vi.fn(),
};

describe('DocumentListItem', () => {
  it('タイトルを表示する（空なら「無題」）', () => {
    const { rerender } = render(<DocumentListItem {...props} />);
    expect(screen.getByText('メモA')).toBeInTheDocument();
    rerender(<DocumentListItem {...props} title="" />);
    expect(screen.getByText('無題')).toBeInTheDocument();
  });

  it('クリックで onSelect(id) を呼ぶ', () => {
    const onSelect = vi.fn();
    render(<DocumentListItem {...props} onSelect={onSelect} />);
    fireEvent.click(screen.getByLabelText('ノート「メモA」を選択'));
    expect(onSelect).toHaveBeenCalledWith('a');
  });

  it('削除ボタンで onDelete(id) を呼び、選択はしない（stopPropagation）', () => {
    const onSelect = vi.fn();
    const onDelete = vi.fn();
    render(<DocumentListItem {...props} onSelect={onSelect} onDelete={onDelete} />);
    fireEvent.click(screen.getByLabelText('ノート「メモA」を削除'));
    expect(onDelete).toHaveBeenCalledWith('a');
    expect(onSelect).not.toHaveBeenCalled();
  });

  it('isActive のとき aria-pressed=true', () => {
    render(<DocumentListItem {...props} isActive />);
    expect(screen.getByLabelText('ノート「メモA」を選択')).toHaveAttribute('aria-pressed', 'true');
  });
});
