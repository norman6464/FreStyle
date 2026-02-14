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

  it('デフォルトのtype属性がbuttonである', () => {
    render(<PrimaryButton>テスト</PrimaryButton>);
    expect(screen.getByText('テスト')).toHaveAttribute('type', 'button');
  });

  it('type=submitを指定できる', () => {
    render(<PrimaryButton type="submit">送信</PrimaryButton>);
    expect(screen.getByText('送信')).toHaveAttribute('type', 'submit');
  });

  it('全幅のスタイルが適用される', () => {
    render(<PrimaryButton>テスト</PrimaryButton>);
    expect(screen.getByText('テスト').className).toContain('w-full');
  });

  it('loading時にスピナーが表示される', () => {
    const { container } = render(<PrimaryButton loading>送信</PrimaryButton>);
    expect(container.querySelector('[data-testid="loading-spinner"]')).toBeInTheDocument();
  });

  it('loading時にボタンが無効化される', () => {
    render(<PrimaryButton loading>送信</PrimaryButton>);
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('loading=false時にスピナーが表示されない', () => {
    const { container } = render(<PrimaryButton>送信</PrimaryButton>);
    expect(container.querySelector('[data-testid="loading-spinner"]')).not.toBeInTheDocument();
  });

  it('loading時にchildrenも表示される', () => {
    render(<PrimaryButton loading>送信中</PrimaryButton>);
    expect(screen.getByText('送信中')).toBeInTheDocument();
  });

  it('loadingとdisabled両方設定時にボタンが無効化される', () => {
    render(<PrimaryButton loading disabled>送信</PrimaryButton>);
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('loading時にonClickが呼ばれない', () => {
    const onClick = vi.fn();
    render(<PrimaryButton loading onClick={onClick}>送信</PrimaryButton>);
    fireEvent.click(screen.getByRole('button'));
    expect(onClick).not.toHaveBeenCalled();
  });
});
