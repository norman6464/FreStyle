import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import InlineCodeButton from '../InlineCodeButton';

describe('InlineCodeButton', () => {
  it('インラインコードボタンが表示される', () => {
    render(<InlineCodeButton onInlineCode={vi.fn()} />);
    expect(screen.getByLabelText('インラインコード')).toBeInTheDocument();
  });

  it('クリックでonInlineCodeが呼ばれる', () => {
    const onInlineCode = vi.fn();
    render(<InlineCodeButton onInlineCode={onInlineCode} />);
    fireEvent.click(screen.getByLabelText('インラインコード'));
    expect(onInlineCode).toHaveBeenCalledTimes(1);
  });

  it('ボタンがbutton要素である', () => {
    render(<InlineCodeButton onInlineCode={vi.fn()} />);
    const button = screen.getByLabelText('インラインコード');
    expect(button.tagName).toBe('BUTTON');
    expect(button).toHaveAttribute('type', 'button');
  });
});
