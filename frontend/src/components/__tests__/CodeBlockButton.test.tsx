import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import CodeBlockButton from '../CodeBlockButton';

describe('CodeBlockButton', () => {
  it('コードブロックボタンが表示される', () => {
    render(<CodeBlockButton onCodeBlock={vi.fn()} />);
    expect(screen.getByLabelText('コードブロック')).toBeInTheDocument();
  });

  it('クリックでonCodeBlockが呼ばれる', () => {
    const onCodeBlock = vi.fn();
    render(<CodeBlockButton onCodeBlock={onCodeBlock} />);
    fireEvent.click(screen.getByLabelText('コードブロック'));
    expect(onCodeBlock).toHaveBeenCalledTimes(1);
  });

  it('ボタンがbutton要素である', () => {
    render(<CodeBlockButton onCodeBlock={vi.fn()} />);
    const button = screen.getByLabelText('コードブロック');
    expect(button.tagName).toBe('BUTTON');
    expect(button).toHaveAttribute('type', 'button');
  });
});
