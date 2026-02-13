import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SNSSignInButton from '../SNSSignInButton';

describe('SNSSignInButton', () => {
  it('Googleログインボタンが表示される', () => {
    render(<SNSSignInButton provider="google" onClick={vi.fn()} />);

    expect(screen.getByText('Googleでログイン')).toBeInTheDocument();
  });

  it('クリックでonClickが呼ばれる', () => {
    const mockOnClick = vi.fn();
    render(<SNSSignInButton provider="google" onClick={mockOnClick} />);

    fireEvent.click(screen.getByText('Googleでログイン'));
    expect(mockOnClick).toHaveBeenCalled();
  });

  it('Facebookログインボタンが表示される', () => {
    render(<SNSSignInButton provider="facebook" onClick={vi.fn()} />);

    expect(screen.getByText('Facebookでログイン')).toBeInTheDocument();
  });
});
