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

  it('複数回クリックでonBlockquoteが複数回呼ばれる', () => {
    const onBlockquote = vi.fn();
    render(<BlockquoteButton onBlockquote={onBlockquote} />);
    const button = screen.getByLabelText('引用');
    fireEvent.click(button);
    fireEvent.click(button);
    fireEvent.click(button);
    expect(onBlockquote).toHaveBeenCalledTimes(3);
  });

  it('SVGアイコンがレンダリングされる', () => {
    render(<BlockquoteButton onBlockquote={vi.fn()} />);
    const button = screen.getByLabelText('引用');
    expect(button.querySelector('svg')).toBeInTheDocument();
  });
});
