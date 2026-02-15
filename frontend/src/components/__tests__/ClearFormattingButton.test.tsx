import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ClearFormattingButton from '../ClearFormattingButton';

describe('ClearFormattingButton', () => {
  it('ボタンが表示される', () => {
    render(<ClearFormattingButton onClearFormatting={vi.fn()} />);
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('aria-labelが設定されている', () => {
    render(<ClearFormattingButton onClearFormatting={vi.fn()} />);
    expect(screen.getByLabelText('書式クリア')).toBeInTheDocument();
  });

  it('クリックでonClearFormattingが呼ばれる', () => {
    const onClearFormatting = vi.fn();
    render(<ClearFormattingButton onClearFormatting={onClearFormatting} />);
    fireEvent.click(screen.getByRole('button'));
    expect(onClearFormatting).toHaveBeenCalledOnce();
  });
});
