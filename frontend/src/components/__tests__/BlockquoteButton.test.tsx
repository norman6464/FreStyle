import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import BlockquoteButton from '../BlockquoteButton';

describe('BlockquoteButton', () => {
  it('引用ボタンが表示される', () => {
    render(<BlockquoteButton onBlockquote={vi.fn()} />);
    expect(screen.getByLabelText('引用')).toBeInTheDocument();
  });

  it('クリックでonBlockquoteが呼ばれる', () => {
    const onBlockquote = vi.fn();
    render(<BlockquoteButton onBlockquote={onBlockquote} />);
    fireEvent.click(screen.getByLabelText('引用'));
    expect(onBlockquote).toHaveBeenCalledTimes(1);
  });

  it('ボタンがbutton要素である', () => {
    render(<BlockquoteButton onBlockquote={vi.fn()} />);
    const button = screen.getByLabelText('引用');
    expect(button.tagName).toBe('BUTTON');
    expect(button).toHaveAttribute('type', 'button');
  });
});
