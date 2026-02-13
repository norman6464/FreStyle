import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import PrimaryButton from '../PrimaryButton';

describe('PrimaryButton', () => {
  it('子要素が表示される', () => {
    render(<PrimaryButton>ログイン</PrimaryButton>);

    expect(screen.getByText('ログイン')).toBeInTheDocument();
  });

  it('クリックでonClickが呼ばれる', () => {
    const mockOnClick = vi.fn();
    render(<PrimaryButton onClick={mockOnClick}>送信</PrimaryButton>);

    fireEvent.click(screen.getByText('送信'));
    expect(mockOnClick).toHaveBeenCalled();
  });

  it('disabled時にクリックが無効になる', () => {
    const mockOnClick = vi.fn();
    render(<PrimaryButton onClick={mockOnClick} disabled>送信</PrimaryButton>);

    const button = screen.getByText('送信');
    expect(button).toBeDisabled();
  });
});
